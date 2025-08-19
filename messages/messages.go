package messages

import (
	buf "bytes"
	"encoding/binary"
	"fmt"
	"reflect"
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

// ============= Originating Transaction ID

type OriginatingTransactionID struct {
	SourceEntityID            uint16
	TransactionSequenceNumber uint16
}

func (o OriginatingTransactionID) GetMessageType() MessageType {
	return MessageTypeOriginatingTransactionID
}

func (o OriginatingTransactionID) ToBytes() ([]byte, error) {
	bytes := new(buf.Buffer)

	bytes.WriteByte(byte(reflect.TypeOf(o.SourceEntityID).Size() - 1))
	bytes.WriteByte(byte(reflect.TypeOf(o.TransactionSequenceNumber).Size() - 1))

	if err := binary.Write(bytes, binary.LittleEndian, o.SourceEntityID); err != nil {
		return nil, fmt.Errorf("failed to write sourceEntityID: %v", err)
	}
	if err := binary.Write(bytes, binary.LittleEndian, o.TransactionSequenceNumber); err != nil {
		return nil, fmt.Errorf("failed to write transactionSequenceNumber: %v", err)
	}

	return bytes.Bytes(), nil
}

func (o *OriginatingTransactionID) FromBytes(data []byte) error {
	buf := buf.NewReader(data)

	// Read and discard the length prefix for SourceEntityID
	if _, err := buf.ReadByte(); err != nil {
		return fmt.Errorf("failed to read sourceEntityID size: %v", err)
	}

	// Read and discard the length prefix for TransactionSequenceNumber
	if _, err := buf.ReadByte(); err != nil {
		return fmt.Errorf("failed to read transactionSequenceNumber size: %v", err)
	}

	// Read the actual values, which matches the ToBytes implementation
	if err := binary.Read(buf, binary.LittleEndian, &o.SourceEntityID); err != nil {
		return fmt.Errorf("failed to read sourceEntityID: %v", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &o.TransactionSequenceNumber); err != nil {
		return fmt.Errorf("failed to read transactionSequenceNumber: %v", err)
	}

	return nil
}

func NewOriginatingTransactionID(sourceEntityID uint16, transactionSequenceNumber uint16) OriginatingTransactionID {
	return OriginatingTransactionID{
		SourceEntityID:            sourceEntityID,
		TransactionSequenceNumber: transactionSequenceNumber,
	}
}

// ============= Directory Listing

type DirectoryListingRequest struct {
	DirToList     string // local to the remote entity, the directory to list
	PathToRespond string // full file path local to the caller, where the listing will be saved
}

func (d *DirectoryListingRequest) GetMessageType() MessageType {
	return MessageTypeDirectoryRequest
}

func (d *DirectoryListingRequest) ToBytes() ([]byte, error) {
	bytes := new(buf.Buffer)

	bytes.WriteString("cfpd")
	bytes.WriteByte(byte(d.GetMessageType()))

	bytes.WriteByte(byte(len(d.DirToList)))
	bytes.WriteString(d.DirToList)

	bytes.WriteByte(byte(len(d.PathToRespond)))
	bytes.WriteString(d.PathToRespond)

	return bytes.Bytes(), nil
}

func (d *DirectoryListingRequest) FromBytes(data []byte) error {
	buf := buf.NewReader(data)

	var magic [4]byte
	if _, err := buf.Read(magic[:]); err != nil {
		return fmt.Errorf("failed to read magic: %v", err)
	}

	var messageType MessageType
	if err := binary.Read(buf, binary.LittleEndian, &messageType); err != nil {
		return fmt.Errorf("failed to read message type: %v", err)
	}

	if messageType != MessageTypeDirectoryRequest {
		return fmt.Errorf("invalid message type: %d", messageType)
	}

	var dirToListLen byte
	if err := binary.Read(buf, binary.LittleEndian, &dirToListLen); err != nil {
		return fmt.Errorf("failed to read directory length: %v", err)
	}

	dirToList := make([]byte, dirToListLen)
	if _, err := buf.Read(dirToList); err != nil {
		return fmt.Errorf("failed to read directory name: %v", err)
	}
	var pathToRespondLen byte
	if err := binary.Read(buf, binary.LittleEndian, &pathToRespondLen); err != nil {
		return fmt.Errorf("failed to read path length: %v", err)
	}

	pathToRespond := make([]byte, pathToRespondLen)
	if _, err := buf.Read(pathToRespond); err != nil {
		return fmt.Errorf("failed to read path name: %v", err)
	}

	d.DirToList = string(dirToList)
	d.PathToRespond = string(pathToRespond)

	return nil
}

type DirectoryListingResponse struct {
	WasSuccessful bool   // true if the directory listing was successful, false otherwise
	DirToList     string // the directory that was listed, taken from the listing request
	PathToRespond string // full file path local to the caller, taken from the listing request
}

func (d *DirectoryListingResponse) GetMessageType() MessageType {
	return MessageTypeDirectoryResponse
}

func (d *DirectoryListingResponse) ToBytes() ([]byte, error) {
	bytes := new(buf.Buffer)

	var flags byte
	if d.WasSuccessful {
		flags |= 0b10000000
	}
	bytes.WriteByte(flags)

	// dirToList LV pair
	bytes.WriteByte(byte(len(d.DirToList)))
	bytes.WriteString(d.DirToList)

	// pathToRespond LV pair
	bytes.WriteByte(byte(len(d.PathToRespond)))
	bytes.WriteString(d.PathToRespond)

	return bytes.Bytes(), nil
}

func (d *DirectoryListingResponse) FromBytes(data []byte) error {
	buf := buf.NewReader(data)

	var flags byte
	if err := binary.Read(buf, binary.LittleEndian, &flags); err != nil {
		return fmt.Errorf("failed to read flags: %v", err)
	}
	d.WasSuccessful = (flags & 0b10000000) != 0
	var dirToListLen byte
	if err := binary.Read(buf, binary.LittleEndian, &dirToListLen); err != nil {
		return fmt.Errorf("failed to read directory length: %v", err)
	}
	dirToList := make([]byte, dirToListLen)
	if _, err := buf.Read(dirToList); err != nil {
		return fmt.Errorf("failed to read directory name: %v", err)
	}
	var pathToRespondLen byte
	if err := binary.Read(buf, binary.LittleEndian, &pathToRespondLen); err != nil {
		return fmt.Errorf("failed to read path length: %v", err)
	}
	pathToRespond := make([]byte, pathToRespondLen)
	if _, err := buf.Read(pathToRespond); err != nil {
		return fmt.Errorf("failed to read path name: %v", err)
	}
	d.DirToList = string(dirToList)
	d.PathToRespond = string(pathToRespond)
	return nil
}

// ============= Protocol Data Units
type PduType byte

const (
	FileDirective PduType = 0x0
	FileData      PduType = 0x1
)

type ProtocolDataUnitHeader struct {
	version                        uint8 // "001"
	PduType                        PduType
	direction                      bool // PDU forwarding: false toward file receiver, true toward file sender
	transmissionMode               TransmissionMode
	crcFlag                        bool   // true if CRC is present
	LargeFileFlag                  bool   // files whose size can’t be represented in an unsigned 32-bit integer shall be flagged large
	pduDataFieldLength             int16  // in octets
	segmentationControl            bool   // whether record boundaries are preserved in file data segmentation, always false for file directives
	lengthEntityID                 byte   // number of octets in the entity ID minus one; 0 means 1 octet
	segmentMetadataFlag            bool   // whether the PDU contains segment metadata, always false for file directives
	lenghTransactionSequenceNumber uint8  // number of octets in the transaction sequence number minus one; 0 means 1 octet
	SourceEntityID                 uint16 // identifies the entity that originated the transaction.
	transactionSequenceNumber      uint16 // uniquely identifies the transaction within the source entity
	destinationEntityID            uint16 // identifies the final destination of the transaction’s metadata and file data
}

func (h ProtocolDataUnitHeader) ToBytes(dataFieldLength int16) ([]byte, error) {
	if dataFieldLength < 0 {
		return nil, fmt.Errorf("data field length must be non-negative")
	}
	bytes := new(buf.Buffer)

	// version, pduType, direction, transmissionMode, crcFlag, LargeFileFlag
	var flags byte
	flags |= h.version << 5 // 3 bits for version

	if h.PduType == FileData {
		flags |= 0b00010000 // 1 bit for pduType
	}

	if h.direction {
		flags |= 0b00001000 // 1 bit for direction
	}
	if h.transmissionMode == Unacknowledged {
		flags |= 0b00000100 // 1 bit for transmissionMode (0 for unacknowledged)
	}
	if h.crcFlag {
		flags |= 0b00000010 // 1 bit for crcFlag
	}
	if h.LargeFileFlag {
		flags |= 0b00000001 // 1 bit for LargeFileFlag
	}

	bytes.WriteByte(flags)

	if err := binary.Write(bytes, binary.LittleEndian, dataFieldLength); err != nil {
		return []byte{}, fmt.Errorf("failed to write pduDataFieldLength: %v", err)
	}

	// segmentationControl, lengthEntityID, segmentMetadataFlag, lenghTransactionSequenceNumber
	flags = 0
	if h.PduType == FileData && h.segmentationControl {
		flags |= 0b10000000 // 1 bit for segmentationControl
	}
	flags |= (h.lengthEntityID - 1) << 4 // 3 bits for lengthEntityID
	if h.segmentMetadataFlag {
		flags |= 0b00001000 // 1 bit for segmentMetadataFlag
	}
	flags |= (h.lenghTransactionSequenceNumber - 1) & 0b00000111 // 3 bits for lenghTransactionSequenceNumber

	bytes.WriteByte(flags)

	// sourceEntityID, transactionSequenceNumber, destinationEntityID
	if err := binary.Write(bytes, binary.LittleEndian, h.SourceEntityID); err != nil {
		return []byte{}, fmt.Errorf("failed to write sourceEntityID: %v", err)
	}
	if err := binary.Write(bytes, binary.LittleEndian, h.transactionSequenceNumber); err != nil {
		return []byte{}, fmt.Errorf("failed to write transactionSequenceNumber: %v", err)
	}
	if err := binary.Write(bytes, binary.LittleEndian, h.destinationEntityID); err != nil {
		return []byte{}, fmt.Errorf("failed to write destinationEntityID: %v", err)
	}

	return bytes.Bytes(), nil
}

func (h *ProtocolDataUnitHeader) FromBytes(data []byte) (int, error) {
	buf := buf.NewReader(data)

	// version, pduType, direction, transmissionMode, crcFlag, LargeFileFlag
	var flags byte
	if err := binary.Read(buf, binary.LittleEndian, &flags); err != nil {
		return 0, fmt.Errorf("failed to read flags: %v", err)
	}
	h.version = flags >> 5
	if (flags & 0b00010000) != 0 {
		h.PduType = FileData
	} else {
		h.PduType = FileDirective
	}
	h.direction = (flags & 0b00001000) != 0
	h.transmissionMode = TransmissionMode((flags & 0b00000100) >> 2)
	h.crcFlag = (flags & 0b00000010) != 0
	h.LargeFileFlag = (flags & 0b00000001) != 0

	if err := binary.Read(buf, binary.LittleEndian, &h.pduDataFieldLength); err != nil {
		return 0, fmt.Errorf("failed to read pduDataFieldLength: %v", err)
	}

	// segmentationControl, lengthEntityID, segmentMetadataFlag, lenghTransactionSequenceNumber
	flags = 0
	if err := binary.Read(buf, binary.LittleEndian, &flags); err != nil {
		return 0, fmt.Errorf("failed to read flags: %v", err)
	}
	h.segmentationControl = (flags & 0b10000000) != 0
	h.lengthEntityID = ((flags & 0b01110000) >> 4) + 1 // +1 because it's minus one in the spec
	h.segmentMetadataFlag = (flags & 0b00001000) != 0
	h.lenghTransactionSequenceNumber = flags&0b00000111 + 1 // +1 because it's minus one in the spec

	if err := binary.Read(buf, binary.LittleEndian, &h.SourceEntityID); err != nil {
		return 0, fmt.Errorf("failed to read sourceEntityID: %v", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &h.transactionSequenceNumber); err != nil {
		return 0, fmt.Errorf("failed to read transactionSequenceNumber: %v", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &h.destinationEntityID); err != nil {
		return 0, fmt.Errorf("failed to read destinationEntityID: %v", err)
	}

	// Return the number of bytes read
	bytesRead := int(buf.Size()) - buf.Len()
	return bytesRead, nil
}

func NewPDUHeader(largeFileFlag bool, srcEntityID uint16, dstEntityID uint16, transactionID uint16, pduType PduType) ProtocolDataUnitHeader {
	return ProtocolDataUnitHeader{
		version:                        1,
		PduType:                        pduType,
		direction:                      false,
		transmissionMode:               Unacknowledged,
		crcFlag:                        false,
		LargeFileFlag:                  largeFileFlag,
		pduDataFieldLength:             0, // TODO: set to actual length when data is added
		segmentationControl:            false,
		lengthEntityID:                 4, // 4 octets for uint32
		segmentMetadataFlag:            false,
		lenghTransactionSequenceNumber: 4, // 4 octets for uint32
		SourceEntityID:                 srcEntityID,
		transactionSequenceNumber:      transactionID,
		destinationEntityID:            dstEntityID,
	}
}

type FileDataPDU struct {
	Header   ProtocolDataUnitHeader
	Offset   uint16 // offset in the file data, used for segmentation
	FileData []byte
}

func (pdu *FileDataPDU) ToBytes(dataFieldLength int16) []byte {
	bytes := new(buf.Buffer)

	Header, error := pdu.Header.ToBytes(dataFieldLength)
	if error != nil {
		fmt.Println("Error converting Header to bytes:", error)
		return nil
	}
	bytes.Write(Header)

	// record continuation state, 2 bits - Present if and only if the value of the segment metadata flag in the PDU Header is 1.
	// 00 - this PDU is neither the start nor end of any record,
	//      if `segmentationControl`, then this is a continuation PDU, otherwise this indicates file has no records
	// 01 - this PDU is the first octet of a record and the end of the record is not within this PDU
	// 10 - this PDU is the last octet of a record and the start of the record is not within this PDU
	// 11 - this PDU is both the first and last octet of a record, i.e. the record is contained within this PDU

	// TODO: implement this properly
	if pdu.Header.segmentMetadataFlag {
		bytes.WriteByte(0b01)
		// segment metadata length, 6 bits - Present if and only if the value of the segment metadata flag in the PDU Header is 1.
		bytes.WriteByte(0b00000000) // 0 for now, no segment metadata
		// No segment metadata bytes are written if length is 0
	}

	bytes.WriteByte(byte(pdu.Offset >> 8)) // 2 bytes for offset, left-shifted by 8 bits to fit in 2 bytes
	bytes.WriteByte(byte(pdu.Offset))

	bytes.Write(pdu.FileData) // write the file data

	return bytes.Bytes()
}

func (pdu *FileDataPDU) FromBytes(data []byte) error {
	buf := buf.NewReader(data)

	// Read the Header
	hLen, err := pdu.Header.FromBytes(data)
	if err != nil {
		return fmt.Errorf("failed to read Header: %v", err)
	}
	if hLen > len(data) || hLen < 0 {
		return fmt.Errorf("invalid Header length: %d", hLen)
	}
	buf.Seek(int64(hLen), 0) // Move the reader to the end of the Header

	if pdu.Header.segmentMetadataFlag {
		// Read the record continuation state
		var recordContinuationState byte
		if err := binary.Read(buf, binary.LittleEndian, &recordContinuationState); err != nil {
			return fmt.Errorf("failed to read record continuation state: %v", err)
		}
		// Read the segment metadata length
		var segmentMetadataLength byte
		if err := binary.Read(buf, binary.LittleEndian, &segmentMetadataLength); err != nil {
			return fmt.Errorf("failed to read segment metadata length: %v", err)
		}
		// Read the segment metadata if length > 0
		if segmentMetadataLength > 0 {
			segmentMetadata := make([]byte, segmentMetadataLength)
			if _, err := buf.Read(segmentMetadata); err != nil {
				return fmt.Errorf("failed to read segment metadata: %v", err)
			}
		}
	}

	// Read the offset
	if err := binary.Read(buf, binary.BigEndian, &pdu.Offset); err != nil {
		return fmt.Errorf("failed to read offset: %v", err)
	}

	// Read the file data
	pdu.FileData = make([]byte, buf.Len())
	if _, err := buf.Read(pdu.FileData); err != nil {
		return fmt.Errorf("failed to read file data: %v", err)
	}

	return nil
}

type FileDirectivePDU struct {
	Header  ProtocolDataUnitHeader
	DirCode DirectiveCode
	Data    []byte
}

func (pdu FileDirectivePDU) ToBytes(dataFieldLength int16) []byte {
	bytes := new(buf.Buffer)

	header, error := pdu.Header.ToBytes(dataFieldLength)
	if error != nil {
		fmt.Println("Error converting header to bytes:", error)
		return nil
	}
	bytes.Write(header)

	bytes.WriteByte(byte(pdu.DirCode)) // 1 byte for directive code
	bytes.Write(pdu.Data)              // write the data field

	return bytes.Bytes()
}

func (pdu *FileDirectivePDU) FromBytes(data []byte) error {
	buf := buf.NewReader(data)

	// Read the header
	hLen, err := pdu.Header.FromBytes(data)
	if err != nil {
		return fmt.Errorf("failed to read header: %v", err)
	}
	if hLen > len(data) || hLen < 0 {
		return fmt.Errorf("invalid header length: %d", hLen)
	}
	buf.Seek(int64(hLen), 0) // Move the reader to the end of the header

	// Read the directive code
	var dirCode byte
	if err := binary.Read(buf, binary.LittleEndian, &dirCode); err != nil {
		return fmt.Errorf("failed to read directive code: %v", err)
	}
	pdu.DirCode = DirectiveCode(dirCode)

	// Read the data field
	pdu.Data = make([]byte, pdu.Header.pduDataFieldLength-1) // -1 for the directive code byte
	if _, err := buf.Read(pdu.Data); err != nil {
		return fmt.Errorf("failed to read data field: %v", err)
	}

	return nil
}

func NewFileDirectivePDU(largeFileFlag bool, srcEntityID, dstEntityID uint16, transactionID uint16, pduType PduType, closureRequested bool, srcFileName, dstFileName string, msgs []Message) (*FileDirectivePDU, error) {
	pduHeader := NewPDUHeader(largeFileFlag, srcEntityID, dstEntityID, transactionID, pduType)

	pduContents := MetadataPDUContents{
		ClosureRequested:    closureRequested,
		ChecksumType:        0xff, // Mock checksum for simplicity
		FileSize:            0,
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
	MessagesToUser      []Message
}

func (m MetadataPDUContents) ToBytes(h ProtocolDataUnitHeader) ([]byte, error) {
	// reserved bits left as 0
	bytes := new(buf.Buffer)

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
	buf := buf.NewReader(data)

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
			r := DirectoryListingRequest{}
			err := r.FromBytes(tlv.Value)
			if err != nil {
				return fmt.Errorf("failed to deserialize MessagesToUser: %v", err)
			}
			m.MessagesToUser = append(m.MessagesToUser, &r)

		default:
			return fmt.Errorf("unknown TLV type: %d", tlv.Type)
		}
	}

	return nil
}

type TLVType byte

const (
	FilestoreRequest     TLVType = 0x0
	FilestoreResponse    TLVType = 0x1
	MessagesToUser       TLVType = 0x2
	FaultHandlerOverride TLVType = 0x4
	FlowLabel            TLVType = 0x5
	EntityID             TLVType = 0x6
)

func TLVTypeFromString(b byte) (TLVType, error) {
	switch b {
	case 0x0:
		return FilestoreRequest, nil
	case 0x1:
		return FilestoreResponse, nil
	case 0x2:
		return MessagesToUser, nil
	case 0x4:
		return FaultHandlerOverride, nil
	case 0x5:
		return FlowLabel, nil
	case 0x6:
		return EntityID, nil
	default:
		return 0, fmt.Errorf("unknown TLV type: %d", b)
	}
}

type TLVFormat struct {
	Type  TLVType
	Value []byte
}

func (t TLVFormat) ToBytes() []byte {
	bytes := new(buf.Buffer)
	bytes.WriteByte(byte(t.Type))
	bytes.WriteByte(byte(len(t.Value)))
	bytes.Write(t.Value)
	return bytes.Bytes()
}

func (t *TLVFormat) FromBytes(data *buf.Reader) error {
	b, err := data.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read TLV type: %v", err)
	}
	t.Type, err = TLVTypeFromString(b)
	if err != nil {
		return fmt.Errorf("failed to parse TLV type: %v", err)
	}

	length, err := data.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read TLV length: %v", err)
	}

	if data.Len() < int(length)-1 {
		return fmt.Errorf("data too short for TLV value, expected %d bytes, got %d", length, data.Len())
	}
	// Read the value
	t.Value = make([]byte, length)
	if _, err := data.Read(t.Value); err != nil {
		return fmt.Errorf("failed to read TLV value: %v", err)
	}

	return nil
}
