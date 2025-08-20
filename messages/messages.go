package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type TransmissionMode uint

const (
	// In acknowledged mode, the receiver informs the sender of any undelivered file segments or ancillary data.
	// These are then retransmitted, guaranteeing complete file delivery.
	Acknowledged TransmissionMode = iota
	// In unacknowledged mode, data delivery failures are not reported to the sender and therefore cannot be repaired.
	Unacknowledged
)

// ============= Protocol Messages
type Message interface {
	GetMessageType() MessageType
	ToBytes() ([]byte, error)
	FromBytes(data []byte) error
}

func DeserializeMessageType(data []byte) (MessageType, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("data too short to contain message type")
	}
	buf := bytes.NewReader(data)
	var magic [4]byte
	if _, err := buf.Read(magic[:]); err != nil {
		return 0, fmt.Errorf("failed to read magic: %v", err)
	}
	var messageType MessageType
	if err := binary.Read(buf, binary.LittleEndian, &messageType); err != nil {
		return 0, fmt.Errorf("failed to read message type: %v", err)
	}
	return messageType, nil
}

type MessageType byte

const (
	MessageTypeProxyOperation           MessageType = 0x1
	MessageTypeStatusReportOperation    MessageType = 0x2
	MessageTypeSuspendOperation         MessageType = 0x3
	MessageTypeResumeOperation          MessageType = 0x4
	MessageTypeDirectoryRequest         MessageType = 0x10
	MessageTypeDirectoryResponse        MessageType = 0x11
	MessageTypeOriginatingTransactionID MessageType = 0x0A
)

func (mt MessageType) String() string {
	switch mt {
	case MessageTypeProxyOperation:
		return "ProxyOperation"
	case MessageTypeStatusReportOperation:
		return "StatusReportOperation"
	case MessageTypeSuspendOperation:
		return "SuspendOperation"
	case MessageTypeResumeOperation:
		return "ResumeOperation"
	case MessageTypeDirectoryRequest:
		return "DirectoryRequest"
	case MessageTypeDirectoryResponse:
		return "DirectoryResponse"
	default:
		return "Unknown"
	}
}
