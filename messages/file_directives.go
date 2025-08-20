package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type DirectiveCode uint8

// 0x0-0x3 and 0xD-0xF are reserved for future use
const (
	EOFPDU       DirectiveCode = 0x4
	FinishedPDU  DirectiveCode = 0x5
	ACKPDU       DirectiveCode = 0x6
	MetadataPDU  DirectiveCode = 0x7
	NAKPDU       DirectiveCode = 0x8
	PromptPDU    DirectiveCode = 0x9
	KeepAlivePDU DirectiveCode = 0xC
)

type ConditionCode byte

// 0b1100 - 0b1101 are reserved for future use
const (
	NoError                 ConditionCode = 0b0000
	PositiveAckLimit        ConditionCode = 0b0001
	KeepAliveLimit          ConditionCode = 0b0010
	InvalidTransmissionMode ConditionCode = 0b0011
	FilestoreRejection      ConditionCode = 0b0100
	FileChecksumFailure     ConditionCode = 0b0101
	FileSizeError           ConditionCode = 0b0110
	NakLimit                ConditionCode = 0b0111
	Inactivity              ConditionCode = 0b1000
	InvalidFileStructure    ConditionCode = 0b1001
	CheckLimit              ConditionCode = 0b1010
	UnsupportedChecksum     ConditionCode = 0b1011
	SuspendRequestReceived  ConditionCode = 0b1100
	CancelRequestReceived   ConditionCode = 0b1111
)

func (c ConditionCode) IsError() bool {
	return c != NoError && c != SuspendRequestReceived && c != CancelRequestReceived
}

type Action uint8

const (
	Cancel Action = iota
	Suspend
	Ignore
	Abandon
)

type EOFPDUContents struct {
	ConditionCode ConditionCode
	FileChecksum  uint32
	FileSize      uint64 // takes 64 bits for files with largeFileFlag set, 32 bits otherwise
	FaultLocation uint   // omitted if condition code is 'No error', else it's the ID of the entity at which transaction cancellation was initiated.
}

func NewEOFPDU(largeFileFlag bool, srcEntityID, dstEntityID uint16, transactionID uint16, cc ConditionCode, fc uint32) (*FileDirectivePDU, error) {
	pduHeader := NewPDUHeader(largeFileFlag, srcEntityID, dstEntityID, transactionID, FileDirective)

	// TODO: implement proper EOFPDUContents initialization
	// For now, we will use a mock EOFPDUContents with no error and zero values
	// This should be replaced with actual values based on the file transfer status
	pduContents := EOFPDUContents{
		ConditionCode: NoError,
		FileChecksum:  fc,
		FileSize:      0,
		FaultLocation: 0,
	}
	d, err := pduContents.ToBytes(pduHeader)
	if err != nil {
		fmt.Println("Error converting MetadataPDUContents to bytes:", err)
		return nil, err
	}
	return &FileDirectivePDU{
		Header:  pduHeader,
		DirCode: EOFPDU,
		Data:    d,
	}, nil
}

func (e EOFPDUContents) ToBytes(h ProtocolDataUnitHeader) ([]byte, error) {
	bytes := new(bytes.Buffer)

	bytes.WriteByte(byte(e.ConditionCode))
	bytes.WriteByte(0b0000) // spare, 4 bits, always 0
	if err := binary.Write(bytes, binary.BigEndian, uint32(e.FileChecksum)); err != nil {
		return nil, fmt.Errorf("failed to write fileChecksum: %v", err)
	}

	if h.LargeFileFlag {
		if err := binary.Write(bytes, binary.BigEndian, e.FileSize); err != nil {
			return nil, fmt.Errorf("failed to write fileSize: %v", err)
		}
	} else {
		if err := binary.Write(bytes, binary.BigEndian, uint32(e.FileSize)); err != nil {
			return nil, fmt.Errorf("failed to write fileSize: %v", err)
		}
	}

	if e.ConditionCode.IsError() {
		// entity ID in the TLV is the ID of the entity at which transaction cancellation was initiated.
		tlv := TLVFormat{
			Type: EntityID,
			// TODO: implement proper faultLocation handling
			Value: []byte{123},
		}
		bytes.Write(tlv.ToBytes())
	}

	return bytes.Bytes(), nil
}

func (e *EOFPDUContents) FromBytes(data []byte, h ProtocolDataUnitHeader) error {
	buf := bytes.NewReader(data)

	// Read ConditionCode
	var conditionCode byte
	if err := binary.Read(buf, binary.BigEndian, &conditionCode); err != nil {
		return fmt.Errorf("failed to read condition code: %v", err)
	}
	e.ConditionCode = ConditionCode(conditionCode)

	// Read spare bits (4 bits, always 0)
	var spare byte
	if err := binary.Read(buf, binary.BigEndian, &spare); err != nil {
		return fmt.Errorf("failed to read spare bits: %v", err)
	}

	// Read FileChecksum
	if err := binary.Read(buf, binary.BigEndian, &e.FileChecksum); err != nil {
		return fmt.Errorf("failed to read file checksum: %v", err)
	}

	// Read FileSize
	if h.LargeFileFlag {
		if err := binary.Read(buf, binary.BigEndian, &e.FileSize); err != nil {
			return fmt.Errorf("failed to read file size: %v", err)
		}
	} else {
		var fs32 uint32
		if err := binary.Read(buf, binary.BigEndian, &fs32); err != nil {
			return fmt.Errorf("failed to read file size: %v", err)
		}
		e.FileSize = uint64(fs32)
	}

	if e.ConditionCode.IsError() {
		tlv := TLVFormat{}
		if err := tlv.FromBytes(buf); err != nil {
			return fmt.Errorf("failed to read TLV: %v", err)
		}
		if tlv.Type != EntityID {
			return fmt.Errorf("expected TLV type EntityID, got %d", tlv.Type)
		}
		if len(tlv.Value) != 1 {
			return fmt.Errorf("expected 1 byte for fault location, got %d bytes", len(tlv.Value))
		}
		e.FaultLocation = uint(tlv.Value[0]) // assuming the fault location is a single byte
	}

	return nil
}

type MetadataPDUContents struct {
	ClosureRequested    bool   // true if the file closure is requested, false otherwise; If transaction is in Acknowledged mode, set to ‘0’ and ignored.
	ChecksumType        byte   // Checksum algorithm identifier as registered in the SANA Checksum Types Registry. Zero indicates legacy modular checksum.
	FileSize            uint64 // If Large File flag is zero, the size of FSS data is 32 bits, else it is 64 bits.
	SourceFileName      string
	DestinationFileName string
	MessagesToUser      []Message
}

func NewMetadataPDU(largeFileFlag bool, srcEntityID, dstEntityID uint16, transactionID uint16, closureRequested bool, srcFileName, dstFileName string,
	fileSize uint64, checksumType ChecksumType, msgs []Message) (*FileDirectivePDU, error) {
	pduHeader := NewPDUHeader(largeFileFlag, srcEntityID, dstEntityID, transactionID, FileDirective)

	pduContents := MetadataPDUContents{
		ClosureRequested:    closureRequested,
		ChecksumType:        byte(checksumType),
		FileSize:            fileSize,
		SourceFileName:      srcFileName,
		DestinationFileName: dstFileName,
		MessagesToUser:      msgs,
	}
	d, err := pduContents.ToBytes(pduHeader)
	if err != nil {
		fmt.Println("Error converting MetadataPDUContents to bytes:", err)
		return nil, err
	}
	return &FileDirectivePDU{
		Header:  pduHeader,
		DirCode: MetadataPDU,
		Data:    d,
	}, nil
}

func (m MetadataPDUContents) ToBytes(h ProtocolDataUnitHeader) ([]byte, error) {
	// reserved bits left as 0
	bytes := new(bytes.Buffer)

	// first byte: reserved bits, ClosureRequested
	var flags byte
	if m.ClosureRequested { // TODO: if transaction is in Acknowledged mode, set to ‘0’ and ignored
		flags |= 0b00000100
	}
	bytes.WriteByte(flags)
	// second byte: ChecksumType
	bytes.WriteByte(m.ChecksumType)

	// FileSize
	if h.LargeFileFlag {
		binary.Write(bytes, binary.LittleEndian, m.FileSize)
	} else {
		binary.Write(bytes, binary.LittleEndian, uint32(m.FileSize))
	}

	// SourceFileName
	bytes.WriteByte(byte(len(m.SourceFileName)))
	bytes.WriteString(m.SourceFileName)

	// DestinationFileName
	bytes.WriteByte(byte(len(m.DestinationFileName)))
	bytes.WriteString(m.DestinationFileName)

	for _, option := range m.MessagesToUser {
		s, err := option.ToBytes()
		if err != nil {
			return nil, fmt.Errorf("failed to serialize DirectoryListingRequest: %w", err)
		}
		tlv := TLVFormat{
			Type:  MessagesToUser,
			Value: s,
		}
		bytes.Write(tlv.ToBytes())
	}
	return bytes.Bytes(), nil
}

func (m *MetadataPDUContents) FromBytes(data []byte, h ProtocolDataUnitHeader) error {
	buf := bytes.NewReader(data)

	// Flags
	var flags byte
	if err := binary.Read(buf, binary.LittleEndian, &flags); err != nil {
		return err
	}
	m.ClosureRequested = (flags & 0b00000100) != 0

	// ChecksumType
	if err := binary.Read(buf, binary.LittleEndian, &m.ChecksumType); err != nil {
		return err
	}

	// FileSize
	if h.LargeFileFlag {
		if err := binary.Read(buf, binary.LittleEndian, &m.FileSize); err != nil {
			return err
		}
	} else {
		var fs32 uint32
		if err := binary.Read(buf, binary.LittleEndian, &fs32); err != nil {
			return err
		}
		m.FileSize = uint64(fs32)
	}

	// SourceFileName
	var srcLen byte
	if err := binary.Read(buf, binary.LittleEndian, &srcLen); err != nil {
		return err
	}
	src := make([]byte, srcLen)
	if _, err := buf.Read(src); err != nil {
		return err
	}
	m.SourceFileName = string(src)

	// DestinationFileName
	var dstLen byte
	if err := binary.Read(buf, binary.LittleEndian, &dstLen); err != nil {
		return err
	}
	dst := make([]byte, dstLen)
	if _, err := buf.Read(dst); err != nil {
		return err
	}
	m.DestinationFileName = string(dst)

	// Options
	for buf.Len() > 0 {
		tlv := TLVFormat{}
		err := tlv.FromBytes(buf)
		if err != nil {
			return err
		}
		switch tlv.Type {
		case MessagesToUser:
			mt, err := DeserializeMessageType(tlv.Value)
			if err != nil {
				return fmt.Errorf("failed to deserialize message type: %v", err)
			}
			var msg Message
			switch mt {
			case MessageTypeDirectoryRequest:
				msg = &DirectoryListingRequest{}
			case MessageTypeDirectoryResponse:
				msg = &DirectoryListingResponse{}
			default:
				return fmt.Errorf("unknown message type: %d", mt)
			}

			if err := msg.FromBytes(tlv.Value); err != nil {
				return fmt.Errorf("failed to deserialize message: %v", err)
			}
			m.MessagesToUser = append(m.MessagesToUser, msg)

		default:
			return fmt.Errorf("unknown TLV type: %d", tlv.Type)
		}
	}

	return nil
}
