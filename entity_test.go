package main

import (
	"fmt"
	filesystem "cfdp/filesystem"
	"cfdp/messages"
	"cfdp/statemachine"
	"testing"
)

func TestNewEntity(t *testing.T) {
	entity := NewEntity(1, "testEntity", ServiceConfig{})
	if entity.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", entity.ID)
	}
	if entity.Name != "testEntity" {
		t.Errorf("Expected Name to be testEntity, got %s", entity.Name)
	}
}

func TestCFDPEntity_PutRequest(t *testing.T) {
	mockService := &mockCFDPService{}
	entity := NewEntity(1, "testEntity", ServiceConfig{})
	entity.service = mockService

	dstEntityID := uint16(2)
	err := entity.PutRequest(PutParameters{
		DstEntityID:         &dstEntityID,
		SrcFileName:         "/local",
		DestinationFileName: "/remote",
		MessagesToUser: []messages.Message{
			&messages.DirectoryListingRequest{
				DirToList:     "/remote",
				PathToRespond: "/local",
			},
		},
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if mockService.sentPDU == nil {
		t.Fatal("Expected a PDU to be sent")
	}
}

func TestCFDPEntity_HandlePDU_Metadata(t *testing.T) {
	entity := NewEntity(1, "testEntity", ServiceConfig{})
	entity.fs = &filesystem.MockFS{}
	entity.ic = func(i Indication) {}

	metadataPDU, _ := messages.NewMetadataPDU(false, 2, 1, 123, false, "/remote", "/local", 0, messages.CRC32, nil)
	err := entity.HandlePDU(metadataPDU)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if entity.sm.GetState() != statemachine.WaitingForFileData {
		t.Errorf("Expected state WaitingForFileData, got %v", entity.sm.GetState())
	}
}

func TestCFDPEntity_HandlePDU_EOF(t *testing.T) {
	entity := NewEntity(1, "testEntity", ServiceConfig{})
	entity.fs = &filesystem.MockFS{}
	sm := statemachine.NewStateMachine()
	entity.sm = sm
	sm.SetState(statemachine.WaitingForFileData)
	sm.GetContext().ChecksumType = messages.CRC32
	sm.GetContext().FileData = []byte("test data")
	sm.GetContext().FilePath = "/local/test.txt"

	checksum := messages.GetChecksumAlgorithm(messages.CRC32)([]byte("test data"))
	eofPDU, _ := messages.NewEOFPDU(false, 2, 1, 123, messages.NoError, checksum)

	err := entity.HandlePDU(eofPDU)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if sm.GetState() != statemachine.ReceivedDirectoryListing {
		t.Errorf("Expected state ReceivedDirectoryListing, got %v", sm.GetState())
	}

	mockFS, ok := entity.fs.(*filesystem.MockFS)
	if !ok {
		t.Fatal("Expected a MockFS")
	}

	if mockFS.WrittenFile != "/local/test.txt" {
		t.Errorf("Expected file /local/test.txt to be written, got %s", mockFS.WrittenFile)
	}
}

func TestCFDPEntity_HandlePDU_FileData(t *testing.T) {
	entity := NewEntity(1, "testEntity", ServiceConfig{})
	sm := statemachine.NewStateMachine()
	entity.sm = sm
	sm.SetState(statemachine.WaitingForFileData)

	fileDataPDU := &messages.FileDataPDU{
		Header:   messages.NewPDUHeader(false, 2, 1, 123, messages.FileData),
		FileData: []byte("test data"),
	}

	err := entity.HandlePDU(fileDataPDU)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if string(sm.GetContext().FileData) != "test data" {
		t.Errorf("Expected file data 'test data', got '%s'", string(sm.GetContext().FileData))
	}
}

func TestCFDPEntity_UploadDirectoryListing(t *testing.T) {
	mockService := &mockCFDPService{}
	entity := NewEntity(1, "testEntity", ServiceConfig{})
	entity.service = mockService
	entity.fs = &filesystem.MockFS{
		DirListing: "file1.txt\nfile2.txt",
	}

	metadataPDU, _ := messages.NewMetadataPDU(false, 2, 1, 123, false, "/remote", "/local", 0, messages.CRC32, []messages.Message{
		&messages.DirectoryListingRequest{},
	})
	metadataContents := messages.MetadataPDUContents{}
	metadataContents.FromBytes(metadataPDU.Data, metadataPDU.Header)
	metadataContents.SourceFileName = "/remote"
	metadataContents.DestinationFileName = "/local"

	err := entity.UploadDirectoryListing(metadataPDU, metadataContents)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if mockService.sentPDUs == nil || len(mockService.sentPDUs) != 3 {
		t.Fatalf("Expected 3 PDUs to be sent, got %d", len(mockService.sentPDUs))
	}

	// Check metadata PDU
	metadata, ok := mockService.sentPDUs[0].(*messages.FileDirectivePDU)
	if !ok {
		t.Fatalf("Expected a FileDirectivePDU, got %T", mockService.sentPDUs[0])
	}
	if metadata.DirCode != messages.MetadataPDU {
		t.Errorf("Expected MetadataPDU, got %v", metadata.DirCode)
	}

	// Check file data PDU
	fileData, ok := mockService.sentPDUs[1].(*messages.FileDataPDU)
	if !ok {
		t.Fatalf("Expected a FileDataPDU, got %T", mockService.sentPDUs[1])
	}
	if string(fileData.FileData) != "file1.txt\nfile2.txt" {
		t.Errorf("Unexpected file data: %s", string(fileData.FileData))
	}

	// Check EOF PDU
	eof, ok := mockService.sentPDUs[2].(*messages.FileDirectivePDU)
	if !ok {
		t.Fatalf("Expected a FileDirectivePDU, got %T", mockService.sentPDUs[2])
	}
	if eof.DirCode != messages.EOFPDU {
		t.Errorf("Expected EOFPDU, got %v", eof.DirCode)
	}
}

// mockCFDPService is a mock implementation of the CFPDService for testing.
type mockCFDPService struct {
	sentPDU  messages.PDU
	sentPDUs []messages.PDU
}

func (m *mockCFDPService) RequestBytes(p []byte, dEntityID uint16) error {
	header := messages.ProtocolDataUnitHeader{}
	_, err := header.FromBytes(p)
	if err != nil {
		return err
	}

	var pdu messages.PDU
	if header.PduType == messages.FileDirective {
		fdp := &messages.FileDirectivePDU{}
		err := fdp.FromBytes(p)
		if err != nil {
			return err
		}
		pdu = fdp
	} else if header.PduType == messages.FileData {
		fdp := &messages.FileDataPDU{}
		err := fdp.FromBytes(p)
		if err != nil {
			return err
		}
		pdu = fdp
	} else {
		return fmt.Errorf("unknown pdu type")
	}
	m.sentPDU = pdu
	m.sentPDUs = append(m.sentPDUs, pdu)
	return nil
}

func (m *mockCFDPService) Bind(e *CFDPEntity) error {
	return nil
}

func (m *mockCFDPService) Listen(e *CFDPEntity) {
}

func (m *mockCFDPService) ProcessMessage(bytes []byte) (messages.PDU, error) {
	return nil, nil
}
