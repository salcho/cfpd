package messages

import (
	"bytes"
	"fmt"
)

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
	bytes := new(bytes.Buffer)
	bytes.WriteByte(byte(t.Type))
	bytes.WriteByte(byte(len(t.Value)))
	bytes.Write(t.Value)
	return bytes.Bytes()
}

func (t *TLVFormat) FromBytes(data *bytes.Reader) error {
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
