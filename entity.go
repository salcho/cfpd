package main

import (
	"fmt"
	"log/slog"
	filesystem "main/filesystem"
	"main/messages"
	"main/statemachine"
)

type CFDPEntity struct {
	ID      uint16
	Name    string
	service *CFPDService
	fs      filesystem.FS
	sm      *statemachine.StateMachine
	ic      IndicationCallback
}

func (c CFDPEntity) GetID() uint16 {
	return c.ID
}

func (c CFDPEntity) GetName() string {
	return c.Name
}

type PutParameters struct {
	DstEntityID           *uint16
	SrcFileName           string
	DestinationFileName   string
	SegmentationControl   bool
	FaultHandlerOverrides map[messages.ConditionCode]messages.Action
	FlowLabel             string
	TransmissionMode      messages.TransmissionMode
	ClosureRequested      bool
	MessagesToUser        []messages.Message
	FilestoreRequests     []string
}

// The Put.request primitive shall be used by the application to request delivery of a file from the source filestore to a destination filestore.
func (c CFDPEntity) PutRequest(p PutParameters) error {
	if p.DstEntityID == nil {
		fmt.Println("Destination Entity ID is required")
		return fmt.Errorf("destination entity ID is required")
	}

	// check that all conditions in faultHandlerOverrides are error conditions
	if p.FaultHandlerOverrides != nil {
		for condition := range p.FaultHandlerOverrides {
			if !condition.IsError() {
				fmt.Println("Fault handler overrides must only contain error conditions")
				return fmt.Errorf("fault handler overrides must only contain error conditions")
			}
		}
	}

	// TODO: select a default error handler if none is provided, according to the contents of the MIB
	// TODO: implement flow label handling
	// TODO: implement closure requested handling

	dstID := *p.DstEntityID
	slog.Debug("Triggering Put.request", "from", c.ID, "to", dstID)
	for _, msg := range p.MessagesToUser {
		switch msg.GetMessageType() {
		case messages.MessageTypeDirectoryRequest:
			listingRequest := msg.(*messages.DirectoryListingRequest)
			slog.Debug("Sending Directory Request", "remote", listingRequest.DirToList, "local", listingRequest.PathToRespond, "from", c.ID, "dst", dstID)

			// TODO: assign sequence number and transaction ID properly
			pdu, err := messages.NewMetadataPDU(false, c.ID, dstID, 12345, p.ClosureRequested,
				listingRequest.PathToRespond, listingRequest.DirToList, 0, messages.CRC32, []messages.Message{listingRequest})
			if err != nil {
				fmt.Println("Error creating File Directive PDU:", err)
				return err
			}
			if err := c.service.RequestBytes(pdu.ToBytes(), dstID); err != nil {
				fmt.Println("Error sending PDU:", err)
			}
			c.sm.SetState(statemachine.WaitingForDirectoryListingMetadata)
		default:
			fmt.Println(c.ID, "Unknown request primitive: ", msg.GetMessageType())
			return fmt.Errorf("unknown request primitive: %s", msg.GetMessageType())
		}
	}
	return nil
}

func (c CFDPEntity) HandlePDU(pdu messages.PDU) error {
	if pdu.GetType() == messages.FileDirective {
		pdu := pdu.(*messages.FileDirectivePDU)
		switch pdu.DirCode {
		case messages.MetadataPDU:
			slog.Debug("Handling Metadata PDU", "entityID", c.ID)
			metadata := messages.MetadataPDUContents{}
			err := metadata.FromBytes(pdu.Data, pdu.Header)
			if err != nil {
				return fmt.Errorf("failed to decode metadata: %v", err)
			}

			if hasUploadRequest(metadata.MessagesToUser) {
				return c.UploadDirectoryListing(pdu, metadata)
			}

			// TODO: 4.6.1.2.6 pass any errors to the indication callback
			c.ic(MetadataRecv)
			c.sm.SetState(statemachine.WaitingForFileData)
			c.sm.Context.FilePath = metadata.DestinationFileName
			c.sm.Context.ChecksumType = messages.ChecksumType(metadata.ChecksumType)
			return nil
		case messages.EOFPDU:
			slog.Debug("Handling EOF PDU", "entityID", c.ID)
			eofContents := messages.EOFPDUContents{}
			err := eofContents.FromBytes(pdu.Data, pdu.Header)
			if err != nil {
				return fmt.Errorf("failed to decode EOF contents: %v", err)
			}

			if eofContents.ConditionCode != messages.NoError {
				return fmt.Errorf("EOF PDU with error condition: %v", eofContents.ConditionCode)
			}

			if c.sm.CurrentState != statemachine.WaitingForFileData {
				return fmt.Errorf("unexpected state: %v, expected WaitingForFileData", c.sm.CurrentState)
			}

			var checksum = messages.GetChecksumAlgorithm(c.sm.Context.ChecksumType)(c.sm.Context.FileData)
			if eofContents.FileChecksum != checksum {
				return fmt.Errorf("checksum mismatch: expected %d, got %d", eofContents.FileChecksum, checksum)
			}

			err = c.fs.WriteFile(c.sm.Context.FilePath, c.sm.Context.FileData)
			if err != nil {
				return fmt.Errorf("failed to write file data: %v", err)
			}
			slog.Info("File data written successfully", "fileName", c.sm.Context.FilePath, "entityID", c.ID)
			c.sm.SetState(statemachine.ReceivedDirectoryListing)
			return nil

		default:
			return fmt.Errorf("unknown directive code: %v", pdu.DirCode)
		}
	}

	if pdu.GetType() == messages.FileData {
		pdu := pdu.(*messages.FileDataPDU)
		// TODO: implement segment metadata handling & file chunking
		slog.Info("Handling File Data PDU", "entityID", c.ID, "dataLength", len(pdu.FileData))
		if c.sm.CurrentState != statemachine.WaitingForFileData {
			return fmt.Errorf("unexpected state: %v, expected WaitingForFileData", c.sm.CurrentState)
		}

		c.sm.Context.FileData = pdu.FileData
		return nil
	}

	return nil
}

func hasUploadRequest(m []messages.Message) bool {
	for _, msg := range m {
		if msg.GetMessageType() == messages.MessageTypeDirectoryRequest {
			return true
		}
	}
	return false
}

func (c CFDPEntity) UploadDirectoryListing(pdu *messages.FileDirectivePDU, m messages.MetadataPDUContents) error {
	slog.Info("Uploading directory listing", "local", m.DestinationFileName, "uploadTo", m.SourceFileName, "from", c.ID, "to", pdu.Header.SourceEntityID)
	dstEntityID := pdu.Header.SourceEntityID
	srcFileName := m.DestinationFileName
	dstFileName := m.SourceFileName

	l, err := c.fs.ListDirectory(m.DestinationFileName)
	if err != nil {
		// TODO: handle error properly
		return fmt.Errorf("failed to list directory: %v", err)
	}
	// 4.6.1.1.2 the sender shall forward a Metadata PDU
	c.sm.SetState(statemachine.SendingDirectoryListingMetadata)
	var checksumAlgorithm messages.ChecksumType = messages.CRC32
	metadata, err := messages.NewMetadataPDU(false, c.ID, dstEntityID, 12345, true,
		srcFileName, dstFileName, uint64(len(l)), checksumAlgorithm, []messages.Message{
			&messages.DirectoryListingResponse{
				WasSuccessful: true,
				DirToList:     dstFileName,
				PathToRespond: srcFileName,
			},
		})
	if err != nil {
		fmt.Println("Error creating File Directive PDU:", err)
		return nil
	}
	if err := c.service.RequestBytes(metadata.ToBytes(), dstEntityID); err != nil {
		fmt.Println("Error sending PDU:", err)
		return err
	}

	c.sm.SetState(statemachine.UploadingDirectoryListing)
	fdp := messages.FileDataPDU{
		Header:   messages.NewPDUHeader(false, c.ID, dstEntityID, 12345, messages.FileData),
		FileData: []byte(l),
		Offset:   0,
	}

	if err := c.service.RequestBytes(fdp.ToBytes(), dstEntityID); err != nil {
		fmt.Println("Error sending File Data PDU:", err)
		return err
	}

	c.sm.SetState(statemachine.SendingDirectoryListingEOF)
	checksum := messages.GetChecksumAlgorithm(checksumAlgorithm)([]byte(l))
	eof, err := messages.NewEOFPDU(false, c.ID, dstEntityID, 12345, messages.NoError, checksum)
	if err != nil {
		fmt.Println("Error creating EOF PDU:", err)
		return err
	}
	if err := c.service.RequestBytes(eof.ToBytes(), dstEntityID); err != nil {
		fmt.Println("Error sending File Data PDU:", err)
		return err
	}

	return nil
}

type Indication byte

const (
	Transaction Indication = iota
	EOFSent
	TransactionFinished
	MetadataRecv
	FileSegmentRecv
	Report
	Suspended
	Resumed
	Fault
	Abandoned
	EOFRecv
)

type IndicationCallback func(i Indication)

func NewEntity(id uint16, name string, sc ServiceConfig) *CFDPEntity {
	service := &CFPDService{Config: sc}
	entity := CFDPEntity{
		ID:      id,
		Name:    name,
		service: service,
		fs:      &filesystem.LocalFS{},
		sm:      statemachine.NewStateMachine(),
		ic:      nil, // no callback set by default
	}
	service.Bind(&entity)
	return &entity
}

func (c *CFDPEntity) SetIndicationCallback(ic IndicationCallback) {
	c.ic = ic
}
