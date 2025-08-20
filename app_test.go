package main

import (
	"cfdp/messages"
	"testing"
)

func TestNewCFDPApp(t *testing.T) {
	app := NewCFDPApp(1, "testApp", ServiceConfig{})
	if app.Entity == nil {
		t.Error("Expected Entity to be initialized")
	}
	if app.Entity.ID != 1 {
		t.Errorf("Expected Entity ID to be 1, got %d", app.Entity.ID)
	}
	if app.Entity.Name != "testApp" {
		t.Errorf("Expected Entity Name to be testApp, got %s", app.Entity.Name)
	}
}

func TestCFDPApp_ListDirectory(t *testing.T) {
	// Create a mock service to avoid actual network calls
	mockService := &mockCFDPService{}
	app := NewCFDPApp(1, "testApp", ServiceConfig{})
	app.Entity.service = mockService

	err := app.ListDirectory(2, "/local", "/remote")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if mockService.sentPDU == nil {
		t.Fatal("Expected a PDU to be sent")
	}

	pdu, ok := mockService.sentPDU.(*messages.FileDirectivePDU)
	if !ok {
		t.Fatalf("Expected a FileDirectivePDU, got %T", mockService.sentPDU)
	}

	if pdu.Header.SourceEntityID != 1 {
		t.Errorf("Expected source entity ID 1, got %d", pdu.Header.SourceEntityID)
	}

	if pdu.Header.GetDestinationEntityID() != 2 {
		t.Errorf("Expected destination entity ID 2, got %d", pdu.Header.GetDestinationEntityID())
	}

	metadata := messages.MetadataPDUContents{}
	err = metadata.FromBytes(pdu.Data, pdu.Header)
	if err != nil {
		t.Fatalf("Failed to decode metadata: %v", err)
	}

	if len(metadata.MessagesToUser) != 1 {
		t.Fatalf("Expected 1 message to user, got %d", len(metadata.MessagesToUser))
	}

	dirReq, ok := metadata.MessagesToUser[0].(*messages.DirectoryListingRequest)
	if !ok {
		t.Fatalf("Expected a DirectoryListingRequest, got %T", metadata.MessagesToUser[0])
	}

	if dirReq.DirToList != "/remote" {
		t.Errorf("Expected remote path /remote, got %s", dirReq.DirToList)
	}

	if dirReq.PathToRespond != "/local" {
		t.Errorf("Expected local path /local, got %s", dirReq.PathToRespond)
	}
}
