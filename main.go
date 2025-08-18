package main

import (
	"main/messages"
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
	DstEntityID             uint16
	SrcEntityID             uint16
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
	return []byte{}

	// bytes := make(bytes.Buffer)

	// // magic number
	// bytes.WriteString("cfpd")

	// // message type
	// bytes.WriteString(r.ReqType.String())
}

func NewPutRequest(dst uint16, msgs ...messages.Message) RequestPrimitive {
	return RequestPrimitive{
		ReqType:          PutRequest,
		DstEntityID:      dst,
		MessagesToUser:   msgs,
		TransmissionMode: messages.Unacknowledged,
	}
}

func main() {
	srvConf := ServiceConfig{
		entityID: 1,
		address:  "127.0.0.1:11234",
		addressTable: map[uint16]string{
			0: "127.0.0.1:11235",
		},
	}
	s := NewService(srvConf)
	s.Bind()

	clientConf := ServiceConfig{
		entityID: 0,
		address:  "127.0.0.1:11235",
		addressTable: map[uint16]string{
			1: "127.0.0.1:11234",
		},
	}
	cEntity := NewEntity(0, "Client", NewService(clientConf))
	app := NewCFDPApp(&cEntity)
	app.ListDirectory(1, "/path/to/local/dir", "/path/to/remote/dir")

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
