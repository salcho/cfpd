package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

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
	bytes := new(bytes.Buffer)

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
	buf := bytes.NewReader(data)

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
		crcFlag:                        true,
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

type PDU interface {
	ToBytes() []byte
	FromBytes(data []byte) error
	GetType() PduType
}

type FileDataPDU struct {
	Header   ProtocolDataUnitHeader
	Offset   uint16 // offset in the file data, used for segmentation
	FileData []byte
}

func (pdu *FileDataPDU) GetType() PduType {
	return pdu.Header.PduType
}

func (pdu *FileDataPDU) ToBytes() []byte {
	bytes := new(bytes.Buffer)

	Header, error := pdu.Header.ToBytes(int16(len(pdu.FileData)))
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
	buf := bytes.NewReader(data)

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

func (pdu *FileDirectivePDU) GetType() PduType {
	return pdu.Header.PduType
}

func (pdu *FileDirectivePDU) ToBytes() []byte {
	bytes := new(bytes.Buffer)

	header, error := pdu.Header.ToBytes(int16(len(pdu.Data)))
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
	buf := bytes.NewReader(data)

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
	pdu.Data = make([]byte, pdu.Header.pduDataFieldLength)
	if _, err := buf.Read(pdu.Data); err != nil {
		return fmt.Errorf("failed to read data field: %v", err)
	}

	return nil
}
