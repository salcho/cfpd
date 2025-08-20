package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

// ============= Originating Transaction ID

type OriginatingTransactionID struct {
	SourceEntityID            uint16
	TransactionSequenceNumber uint16
}

func (o OriginatingTransactionID) GetMessageType() MessageType {
	return MessageTypeOriginatingTransactionID
}

func (o OriginatingTransactionID) ToBytes() ([]byte, error) {
	bytes := new(bytes.Buffer)

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
	buf := bytes.NewReader(data)

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
	bytes := new(bytes.Buffer)

	bytes.WriteString("cfpd")
	bytes.WriteByte(byte(d.GetMessageType()))

	bytes.WriteByte(byte(len(d.DirToList)))
	bytes.WriteString(d.DirToList)

	bytes.WriteByte(byte(len(d.PathToRespond)))
	bytes.WriteString(d.PathToRespond)

	return bytes.Bytes(), nil
}

func (d *DirectoryListingRequest) FromBytes(data []byte) error {
	buf := bytes.NewReader(data)

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
	bytes := new(bytes.Buffer)

	bytes.WriteString("cfpd")
	bytes.WriteByte(byte(d.GetMessageType()))

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
	buf := bytes.NewReader(data)

	var magic [4]byte
	if _, err := buf.Read(magic[:]); err != nil {
		return fmt.Errorf("failed to read magic: %v", err)
	}

	var messageType MessageType
	if err := binary.Read(buf, binary.LittleEndian, &messageType); err != nil {
		return fmt.Errorf("failed to read message type: %v", err)
	}
	if messageType != MessageTypeDirectoryResponse {
		return fmt.Errorf("invalid message type: %d", messageType)
	}

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
