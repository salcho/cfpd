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

type CFPDService struct {
	Config      ServiceConfig
	conn        *net.UDPConn
	isListening bool
}

func (s *CFPDService) ProcessMessage(bytes []byte) (messages.PDU, error) {
	if len(bytes) == 0 {
		return nil, fmt.Errorf("no data received")
	}

	header := messages.ProtocolDataUnitHeader{}
	_, err := header.FromBytes(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %v", err)
	}

	if header.PduType == messages.FileDirective {
		slog.Debug("Received File Directive PDU", "entityID", s.Config.entityID)
		fdp := messages.FileDirectivePDU{}
		err := fdp.FromBytes(bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to decode File Directive PDU: %v", err)
		}

		switch fdp.DirCode {
		case messages.MetadataPDU:
			slog.Debug("Received Metadata PDU", "entityID", s.Config.entityID)

			metadata := messages.MetadataPDUContents{}
			err = metadata.FromBytes(fdp.Data, fdp.Header)
			if err != nil {
				return nil, fmt.Errorf("failed to decode metadata: %v", err)
			}

			return &fdp, nil

		case messages.EOFPDU:
			slog.Debug("Received EOF PDU", "entityID", s.Config.entityID)

			eofContents := messages.EOFPDUContents{}
			err = eofContents.FromBytes(fdp.Data, fdp.Header)
			if err != nil {
				return nil, fmt.Errorf("failed to decode EOF contents: %v", err)
			}

			if eofContents.ConditionCode != messages.NoError {
				return nil, fmt.Errorf("EOF PDU with error condition: %v", eofContents.ConditionCode)
			}

			return &fdp, nil

		default:
			return nil, fmt.Errorf("unknown directive code: %v", fdp.DirCode)
		}
	}

	if header.PduType == messages.FileData {
		slog.Debug("Received File Data PDU", "entityID", s.Config.entityID)
		fdp := messages.FileDataPDU{}
		err := fdp.FromBytes(bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to decode File Data PDU: %v", err)
		}
		return &fdp, nil
	}

	return nil, fmt.Errorf("unsupported PDU type: %v", header.PduType)
}

func (s *CFPDService) RequestBytes(p []byte, dEntityID uint16) error {
	addr, exists := s.Config.addressTable[dEntityID]
	if !exists {
		return fmt.Errorf("entity ID %d not found in address table", dEntityID)
	}

	resolved, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to resolve address %s: %w", addr, err)
	}

	slog.Debug("Sending PDU", "from", s.Config.entityID, "to", dEntityID)
	conn, err := net.DialUDP("udp", nil, resolved)
	if err != nil {
		return fmt.Errorf("failed to dial UDP: %w", err)
	}

	defer conn.Close()

	if _, err := conn.Write(p); err != nil {
		return fmt.Errorf("failed to send PDU: %w", err)
	}

	slog.Debug("PDU sent successfully", "from", s.Config.entityID, "to", dEntityID)
	return nil
}

func (s *CFPDService) Bind(e *CFDPEntity) error {
	udpAddr, err := net.ResolveUDPAddr("udp", s.Config.address)
	if err != nil {
		return fmt.Errorf("failed to resolve address: %w", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP: %w", err)
	}

	slog.Info("CFDP service up!", "ID", s.Config.entityID, "address", s.Config.address)
	s.isListening = true

	s.conn = conn
	go s.Listen(e)

	return nil
}

func (s *CFPDService) Listen(e *CFDPEntity) {
	defer s.conn.Close()
	for {
		buf := make([]byte, 1024)
		n, _, err := s.conn.ReadFromUDP(buf)
		buf = buf[:n]
		if err != nil {
			fmt.Println("Error reading from UDP:", err)
			s.isListening = false
			break
		}
		pdu, err := s.ProcessMessage(buf)
		if err != nil {
			slog.Info("Error processing message", "error", err, "entityID", e.ID)
			continue
		}
		if err := e.HandlePDU(pdu); err != nil {
			fmt.Println("Error handling PDU:", err)
		}
	}

	s.isListening = false
	fmt.Println("Stopped listening on", s.Config.address)
}
