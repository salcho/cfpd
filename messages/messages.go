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
	GetHeader() MessageHeader
	GetMessageType() MessageType
}

type MessageImpl struct {
	Header      MessageHeader
	MessageType MessageType
}

// func (m MessageImpl) ToBytes() ([]byte, error) {
// 	bytes := []byte{}
// 	// Header
// 	bytes = append(bytes, "cfpd"...)
// 	bytes = append(bytes, byte(m.MessageType))

// 	// bytes = append(bytes, 0x1)
// 	return bytes, nil
// }

func (m MessageImpl) GetHeader() MessageHeader {
	return m.Header
}

func (m MessageImpl) GetMessageType() MessageType {
	return m.MessageType
}

func (m MessageImpl) String() string {
	return fmt.Sprintf("Message{header: %s, messageType: %s}", m.Header.GetMagic(), m.MessageType)
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

type MessageHeader interface {
	GetMagic() string
}

type MessageHeaderImpl struct {
	Magic string
}

func (h MessageHeaderImpl) GetMagic() string {
	return h.Magic
}

func NewHeader() MessageHeader {
	return &MessageHeaderImpl{Magic: "cfpd"}
}

// ============= Originating Transaction ID

type OriginatingTransactionID struct {
	MessageImpl
	reserved                       bool
	lengthEntityID                 uint8 // number of octets in the entity ID minus one; 0 means 1 octet
	reservedTwo                    bool
	lenghTransactionSequenceNumber uint8 // number of octets in the transaction sequence number minus one
	SourceEntityID                 uint  // unsigned binary integer of length lengthEntityID
	TransactionSequenceNumber      uint  // uniquely identifies the transaction within the source entity
}

func NewOriginatingTransactionID(sourceEntityID uint, transactionSequenceNumber uint) OriginatingTransactionID {
	return OriginatingTransactionID{
		MessageImpl: MessageImpl{
			Header:      NewHeader(),
			MessageType: MessageTypeOriginatingTransactionID,
		},
		reserved:                       false,
		lengthEntityID:                 0, // 1 octet
		reservedTwo:                    false,
		lenghTransactionSequenceNumber: 0, // 1 octet
		SourceEntityID:                 sourceEntityID,
		TransactionSequenceNumber:      transactionSequenceNumber,
	}
}

// ============= Directory Listing

type DirectoryListingRequest struct {
	MessageImpl
	DirToList     string // local to the remote entity, the directory to list
	PathToRespond string // full file path local to the caller, where the listing will be saved
}

func NewDirectoryListingRequest(dir string, file string) DirectoryListingRequest {
	return DirectoryListingRequest{
		DirToList:     dir,
		PathToRespond: file,
		MessageImpl: MessageImpl{
			Header:      NewHeader(),
			MessageType: MessageTypeDirectoryRequest,
		},
	}
}

type DirectoryListingResponse struct {
	MessageImpl
	ResponseCode  bool   // true if the directory listing was successful, false otherwise
	Spare         uint8  // all zeros, 7 bits
	DirToList     string // the directory that was listed, taken from the listing request
	PathToRespond string // full file path local to the caller, taken from the listing request
}

func NewDirectoryListingResponse(dirToList string, pathToRespond string) DirectoryListingResponse {
	return DirectoryListingResponse{
		ResponseCode:  true, // assuming success for this example
		Spare:         0,
		DirToList:     dirToList,
		PathToRespond: pathToRespond,
		MessageImpl: MessageImpl{
			Header:      NewHeader(),
			MessageType: MessageTypeDirectoryResponse,
		},
	}
}

// ============= Protocol Data Units

type ProtocolDataUnit interface {
	GetHeader() ProtocolDataUnitHeader
}

type ProtocolDataUnitHeader struct {
	version                        uint8 // "001"
	pduType                        bool  // false for file directive, true for file data
	direction                      bool  // PDU forwarding: false toward file receiver, true toward file sender
	transmissionMode               TransmissionMode
	crcFlag                        bool   // true if CRC is present
	LargeFileFlag                  bool   // files whose size can’t be represented in an unsigned 32-bit integer shall be flagged large
	pduDataFieldLength             uint16 // in octets
	segmentationControl            bool   // whether record boundaries are preserved in file data segmentation, always false for file directives
	lengthEntityID                 uint8  // number of octets in the entity ID minus one; 0 means 1 octet
	segmentMetadataFlag            bool   // whether the PDU contains segment metadata, always false for file directives
	lenghTransactionSequenceNumber uint8  // number of octets in the transaction sequence number minus one; 0 means 1 octet
	sourceEntityID                 uint   // identifies the entity that originated the transaction.
	transactionSequenceNumber      uint   // uniquely identifies the transaction within the source entity
	destinationEntityID            uint   // identifies the final destination of the transaction’s metadata and file data
}

func NewPDUHeader(largeFileFlag bool, srcEntityID uint, dstEntityID uint, transactionID uint) ProtocolDataUnitHeader {
	return ProtocolDataUnitHeader{
		version:                        1,
		pduType:                        false,
		direction:                      false,
		transmissionMode:               Unacknowledged,
		crcFlag:                        false,
		LargeFileFlag:                  largeFileFlag,
		pduDataFieldLength:             0, // TODO: set to actual length when data is added
		segmentationControl:            false,
		lengthEntityID:                 4, // 4 octets for uint32
		segmentMetadataFlag:            false,
		lenghTransactionSequenceNumber: 4, // 4 octets for uint32
		sourceEntityID:                 srcEntityID,
		transactionSequenceNumber:      transactionID,
		destinationEntityID:            dstEntityID,
	}
}

// type FileDataPDU struct {
// 	header ProtocolDataUnitHeader
// 	data   []byte
// }

type FileDirectivePDU struct {
	Header ProtocolDataUnitHeader
	Dc     DirectiveCode
}

func (pdu FileDirectivePDU) GetHeader() ProtocolDataUnitHeader {
	return pdu.Header
}

// ============= File Data PDUs
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

type ConditionCode uint8

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

type EOFPDUData struct {
	conditionCode ConditionCode
	fileChecksum  int32
	fileSize      uint64 // takes 64 bits for files with largeFileFlag set, 32 bits otherwise
	faultLocation uint   // omitted if condition code is 'No error', else it's the ID of the entity at which transaction cancellation was initiated.
}

func NewEOFPDUData(conditionCode ConditionCode, fileChecksum int32, fileSize uint64, faultLocation uint) EOFPDUData {
	return EOFPDUData{
		conditionCode: conditionCode,
		fileChecksum:  fileChecksum,
		fileSize:      fileSize,
		faultLocation: faultLocation,
	}
}

func (e EOFPDUData) GetDirectiveCode() DirectiveCode {
	return EOFPDU
}

func (e EOFPDUData) GetDirectiveParameter() []byte {
	return []byte{
		byte(e.conditionCode),
		0b0000, // spare, 4 bits, always 0
		byte(0b11111111 & e.fileChecksum >> 24),
		byte(e.fileSize),
		byte(e.faultLocation),
	}
}

type MetadataPDUContents struct {
	ClosureRequested    bool   // true if the file closure is requested, false otherwise; If transaction is in Acknowledged mode, set to ‘0’ and ignored.
	ChecksumType        byte   // Checksum algorithm identifier as registered in the SANA Checksum Types Registry. Zero indicates legacy modular checksum.
	FileSize            uint64 // If Large File flag is zero, the size of FSS data is 32 bits, else it is 64 bits.
	SourceFileName      string
	DestinationFileName string
}

func (m MetadataPDUContents) ToBytes(h ProtocolDataUnitHeader) []byte {
	// reserved bits left as 0
	// bytes := make([]byte, 1024)
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

	// TODO: Add Options field, see section 5.2.5 METADATA PDU of the spec
	return bytes.Bytes()
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

	// TODO: Parse Options field if present

	return nil
}
