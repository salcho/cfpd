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
	Serve(p RequestPrimitive)
	RequestPrimitive(p RequestPrimitive) error
	RequestPDU(p messages.ProtocolDataUnit) error
	Bind() error
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

	// TODO deserialize RequestPrimitive and handle it

	decoded := messages.MetadataPDUContents{}
	decoded.FromBytes(bytes, messages.ProtocolDataUnitHeader{LargeFileFlag: false})
	fmt.Println("decoded: ", decoded)

	return nil
}

func (s *CFPDService) Serve(p RequestPrimitive) error {
	switch p.ReqType {
	case PutRequest:
		slog.Info("Serving Put.request", "dst", p.DstEntityID, "src", p.SrcEntityID)
		for _, msg := range p.MessagesToUser {
			switch msg.GetMessageType() {
			case messages.MessageTypeDirectoryRequest:
				listingRequest := msg.(messages.DirectoryListingRequest)
				slog.Info("Handling Directory Request", "dir", listingRequest.DirToList, "file", listingRequest.PathToRespond, "entityID", s.config.entityID)

				fakeListing := []string{"file1.txt", "file2.txt", "file3.txt"}
				// Create a FileDirective PDU to send metadata about the directory listing
				pduHeader := messages.NewPDUHeader(false, s.config.entityID, p.SrcEntityID, 12345)
				// pdu := messages.FileDirectivePDU{
				// 	Header: pduHeader,
				// 	Dc:     messages.MetadataPDU,
				// }
				pduContents := messages.MetadataPDUContents{
					ClosureRequested:    false,
					ChecksumType:        0, // Assuming no checksum for simplicity
					FileSize:            uint64(len(fakeListing)),
					SourceFileName:      listingRequest.DirToList,
					DestinationFileName: listingRequest.PathToRespond,
				}

				if err := s.RequestBytes(pduContents.ToBytes(pduHeader), p.SrcEntityID); err != nil {
					fmt.Println("Error sending PDU:", err)
				}

				if err := s.Request(
					RequestPrimitive{
						ReqType:          PutRequest,
						DstEntityID:      p.SrcEntityID,
						SrcEntityID:      s.config.entityID,
						TransmissionMode: messages.Acknowledged,
						MessagesToUser: []messages.Message{
							messages.NewDirectoryListingResponse(listingRequest.DirToList, listingRequest.PathToRespond),
							messages.NewOriginatingTransactionID(1, 12345),
						},
					},
					p.SrcEntityID,
				); err != nil {
					fmt.Println("Error handling directory request:", err)
					return err
				}
			case messages.MessageTypeDirectoryResponse:
				listingResponse := msg.(messages.DirectoryListingResponse)
				slog.Info("Handling Directory Response", "dir", listingResponse.DirToList, "file", listingResponse.PathToRespond, "entityID", s.config.entityID)
				return nil
			case messages.MessageTypeOriginatingTransactionID:
				originatingTransactionID := msg.(messages.OriginatingTransactionID)
				fmt.Println(s.config.entityID, "Handling Originating Transaction ID:", originatingTransactionID.SourceEntityID, originatingTransactionID.TransactionSequenceNumber)
				return nil
			default:
				fmt.Println(s.config.entityID, "Unknown request primitive: ", msg.GetMessageType())
				return fmt.Errorf("unknown request primitive: %s", msg.GetMessageType())
			}
		}
	}
	return nil
}

func (s *CFPDService) Request(p RequestPrimitive, entityID uint16) error {
	addr, exists := s.config.addressTable[entityID]
	if !exists {
		return fmt.Errorf("entity ID %d not found in address table", entityID)
	}

	resolved, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to resolve address %s: %w", addr, err)
	}

	slog.Info("Sending request", "type", p.ReqType, "entityID", entityID, "resolved", resolved)
	conn, err := net.DialUDP("udp", nil, resolved)
	if err != nil {
		return fmt.Errorf("failed to dial UDP: %w", err)
	}

	defer conn.Close()

	// var buf bytes.Buffer
	// enc := gob.NewEncoder(&buf)
	// gob.Register(messages.DirectoryListingRequest{})
	// gob.Register(messages.MessageHeaderImpl{})
	// gob.Register(messages.DirectoryListingResponse{})
	// gob.Register(messages.OriginatingTransactionID{})
	// if err := enc.Encode(p); err != nil {
	// return fmt.Errorf("failed to encode request primitive: %w", err)
	// }

	if _, err := conn.Write(p.ToBytes()); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
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
		s.ProcessMessage(buf)
	}

	s.isListening = false
	fmt.Println("Stopped listening on", s.config.address)
}
