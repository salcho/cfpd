package main

import (
	"cfdp/messages"
	"net"
	"testing"
)

func TestCFPDService_ProcessMessage_Empty(t *testing.T) {
	service := &CFPDService{}
	_, err := service.ProcessMessage([]byte{})
	if err == nil {
		t.Error("Expected error for empty message, got nil")
	}
}

func TestCFPDService_ProcessMessage_Metadata(t *testing.T) {
	service := &CFPDService{}
	metadataPDU, _ := messages.NewMetadataPDU(false, 1, 2, 123, false, "/remote", "/local", 0, messages.CRC32, nil)
	pdu, err := service.ProcessMessage(metadataPDU.ToBytes())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if pdu.GetType() != messages.FileDirective {
		t.Errorf("Expected FileDirective PDU, got %v", pdu.GetType())
	}
}

func TestCFPDService_RequestBytes(t *testing.T) {
	// Create a listener to act as the remote entity
	addr := "127.0.0.1:0" // Use port 0 to let the OS choose a free port
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		t.Fatalf("Failed to resolve UDP address: %v", err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatalf("Failed to listen on UDP: %v", err)
	}
	defer conn.Close()

	service := &CFPDService{
		Config: ServiceConfig{
			entityID: 1,
			addressTable: map[uint16]string{
				2: conn.LocalAddr().String(),
			},
		},
	}

	metadataPDU, _ := messages.NewMetadataPDU(false, 1, 2, 123, false, "/remote", "/local", 0, messages.CRC32, nil)
	err = service.RequestBytes(metadataPDU.ToBytes(), 2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Read from the connection to verify the PDU was sent
	buf := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("Failed to read from UDP: %v", err)
	}

	if n == 0 {
		t.Fatal("No data received")
	}
}
