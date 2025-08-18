package main

import (
	"fmt"
	"log/slog"
	"main/messages"
	"net"
	"time"
)

type RequestType uint

const (
	PutRequest RequestType = iota
)

func (r RequestType) String() string {
	switch r {
	case PutRequest:
		return "PutRequest"
	default:
		return "UnknownRequest"
	}
}

type RequestPrimitive struct {
	ReqType                 RequestType
	DstEntityID             uint
	SrcEntityID             uint
	SrcFileName             string
	DstFileName             string
	SegmentControl          string
	TransactionID           string
	ConditionCode           string
	StatusReport            string
	FaultHandlerOverride    string
	TransmissionMode        messages.TransmissionMode // optional
	MessagesToUser          []messages.Message        // optional
	FilestoreRequests       []string                  // optional
	FilestoreResponses      []string                  // optional
	FlowLabel               string                    // optional
	Offset                  uint64
	Length                  uint64
	Progress                float64
	ClosureRequested        bool // optional
	FileSize                uint64
	SegmentData             []byte // optional
	SegmentMetadataLength   uint64 // optional
	RecordContinuationState string // optional
}

func (r RequestPrimitive) ToBytes() []byte {

}

func NewPutRequest(dst uint, msgs ...messages.Message) RequestPrimitive {
	return RequestPrimitive{
		ReqType:          PutRequest,
		DstEntityID:      dst,
		MessagesToUser:   msgs,
		TransmissionMode: messages.Unacknowledged,
	}
}

type ServiceConfig struct {
	entityID     uint
	address      string
	addressTable map[uint]string // entity IDs to addresses
}

type Service interface {
	Serve(p RequestPrimitive)
	RequestPrimitive(p RequestPrimitive) error
	RequestPDU(p messages.ProtocolDataUnit) error
	Bind() error
}

type CFPDService struct {
	config      ServiceConfig
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
	fmt.Println(decoded)

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

				if err := s.RequestPDU(pduContents.ToBytes(pduHeader), p.SrcEntityID); err != nil {
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

func (s *CFPDService) Request(p RequestPrimitive, entityID uint) error {
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

func (s *CFPDService) RequestPDU(p []byte, dEntityID uint) error {
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

func (s *CFPDService) BindAndListen() error {
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
	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		_, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error reading from UDP:", err)
			s.isListening = false
			break
		}
		s.ProcessMessage(buf)
	}

	s.isListening = false
	return nil
}

func main() {
	srvConf := ServiceConfig{
		entityID: 1,
		address:  "127.0.0.1:11234",
		addressTable: map[uint]string{
			0: "127.0.0.1:11235",
		},
	}
	s := NewService(srvConf)
	go s.BindAndListen()

	clientConf := ServiceConfig{
		entityID: 0,
		address:  "127.0.0.1:11235",
		addressTable: map[uint]string{
			1: "127.0.0.1:11234",
		},
	}
	client := NewEntity(0, "Client", NewService(clientConf))
	client.ListDirectory(1, "/path/to/local/dir", "/path/to/remote/dir")

	// listingRequest := RequestPrimitive{
	// 	ReqType:          PutRequest,
	// 	DstEntityID:      1,
	// 	SrcEntityID:      0,
	// 	TransmissionMode: messages.Unacknowledged,
	// 	MessagesToUser: []messages.Message{
	// 		messages.NewDirectoryListingRequest("/path/to/directory", "/path/to/directory/listing.txt"),
	// 	},
	// }

	// err := c.Request(listingRequest, 1)

	// fmt.Println("CFDP ID:", s.GetID())
	// fmt.Println("CFDP Name:", s.GetName())
	// cfdp.ProxyOperation()
	// cfdp.StatusReportOperation()
	// cfdp.SuspendOperation()
	// cfdp.ResumeOperation()
	for {
		// Prevent busy loop
		time.Sleep(10000 * time.Millisecond)
	}
}
