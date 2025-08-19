package main

import (
	"fmt"
	"log/slog"
	"main/messages"
	"net"
)

type ServiceConfig struct {
	entityID     uint16
	address      string
	addressTable map[uint16]string // entity IDs to addresses
}

type Service interface {
	ProcessMessage(bytes []byte) error
	RequestBytes(p []byte, dEntityID uint16) error
	Bind() error
	Listen()
}

type CFPDService struct {
	config      ServiceConfig
	conn        *net.UDPConn
	isListening bool
}

func NewService(sc ServiceConfig) *CFPDService {
	return &CFPDService{config: sc}
}

func (s *CFPDService) ProcessMessage(bytes []byte) error {
	if len(bytes) == 0 {
		return fmt.Errorf("no data received")
	}

	header := messages.ProtocolDataUnitHeader{}
	_, err := header.FromBytes(bytes)
	if err != nil {
		return fmt.Errorf("failed to decode header: %v", err)
	}

	if header.PduType == messages.FileDirective {
		fmt.Println("Received File Directive PDU")
		decoded := messages.FileDirectivePDU{}
		err := decoded.FromBytes(bytes)
		if err != nil {
			return fmt.Errorf("failed to decode File Directive PDU: %v", err)
		}

		metadata := messages.MetadataPDUContents{}
		err = metadata.FromBytes(decoded.Data, decoded.Header)
		if err != nil {
			return fmt.Errorf("failed to decode metadata: %v", err)
		}

		fmt.Printf("Metadata PDU Contents: %#v", metadata)
	}

	return nil
}

func (s *CFPDService) RequestBytes(p []byte, dEntityID uint16) error {
	addr, exists := s.config.addressTable[dEntityID]
	if !exists {
		return fmt.Errorf("entity ID %d not found in address table", dEntityID)
	}

	resolved, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to resolve address %s: %w", addr, err)
	}

	slog.Info("Sending PDU", "entityID", dEntityID, "resolved", resolved)
	conn, err := net.DialUDP("udp", nil, resolved)
	if err != nil {
		return fmt.Errorf("failed to dial UDP: %w", err)
	}

	defer conn.Close()

	if _, err := conn.Write(p); err != nil {
		return fmt.Errorf("failed to send PDU: %w", err)
	}

	slog.Info("PDU sent successfully", "entityID", dEntityID)
	return nil
}

func (s *CFPDService) Bind() error {
	udpAddr, err := net.ResolveUDPAddr("udp", s.config.address)
	if err != nil {
		return fmt.Errorf("failed to resolve address: %w", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP: %w", err)
	}

	fmt.Printf("Entity %d bound to address %s\n", s.config.entityID, s.config.address)
	s.isListening = true

	s.conn = conn
	go s.Listen()

	return nil
}

func (s *CFPDService) Listen() {
	defer s.conn.Close()
	for {
		buf := make([]byte, 1024)
		_, _, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error reading from UDP:", err)
			s.isListening = false
			break
		}
		err = s.ProcessMessage(buf)
		if err != nil {
			fmt.Println("Error processing message:", err)
			continue
		}
	}

	s.isListening = false
	fmt.Println("Stopped listening on", s.config.address)
}
