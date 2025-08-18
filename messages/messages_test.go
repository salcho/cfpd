package messages

import (
	"testing"
)

func TestMetadataPDUContents_ToBytesAndFromBytes(t *testing.T) {
	// Setup a sample header and contents
	header := ProtocolDataUnitHeader{
		LargeFileFlag: true,
	}
	original := MetadataPDUContents{
		ClosureRequested:    true,
		ChecksumType:        0x2,
		FileSize:            0x123456789ABCDEF0,
		SourceFileName:      "source.txt",
		DestinationFileName: "dest.txt",
	}

	// Encode to bytes
	data := original.ToBytes(header)

	// Decode from bytes
	var decoded MetadataPDUContents
	decoded.FromBytes(data, header)

	// Check fields
	if decoded.ClosureRequested != original.ClosureRequested {
		t.Errorf("ClosureRequested mismatch: got %v, want %v", decoded.ClosureRequested, original.ClosureRequested)
	}
	if decoded.ChecksumType != original.ChecksumType {
		t.Errorf("ChecksumType mismatch: got %v, want %v", decoded.ChecksumType, original.ChecksumType)
	}
	if decoded.FileSize != original.FileSize {
		t.Errorf("FileSize mismatch: got %v, want %v", decoded.FileSize, original.FileSize)
	}
	if decoded.SourceFileName != original.SourceFileName {
		t.Errorf("SourceFileName mismatch: got %v, want %v", decoded.SourceFileName, original.SourceFileName)
	}
	if decoded.DestinationFileName != original.DestinationFileName {
		t.Errorf("DestinationFileName mismatch: got %v, want %v", decoded.DestinationFileName, original.DestinationFileName)
	}
}

func TestMetadataPDUContents_SmallFileFlag(t *testing.T) {
	header := ProtocolDataUnitHeader{
		LargeFileFlag: false,
	}
	original := MetadataPDUContents{
		ClosureRequested:    false,
		ChecksumType:        0x1,
		FileSize:            0x12345678,
		SourceFileName:      "foo",
		DestinationFileName: "bar",
	}

	data := original.ToBytes(header)
	var decoded MetadataPDUContents
	decoded.FromBytes(data, header)

	if decoded.FileSize != original.FileSize {
		t.Errorf("FileSize mismatch for small file: got %v, want %v", decoded.FileSize, original.FileSize)
	}
	if decoded.SourceFileName != original.SourceFileName {
		t.Errorf("SourceFileName mismatch: got %v, want %v", decoded.SourceFileName, original.SourceFileName)
	}
	if decoded.DestinationFileName != original.DestinationFileName {
		t.Errorf("DestinationFileName mismatch: got %v, want %v", decoded.DestinationFileName, original.DestinationFileName)
	}
}

func TestProtocolDataUnitHeader_ToBytesAndFromBytes(t *testing.T) {
	original := ProtocolDataUnitHeader{
		version:                        1,
		pduType:                        true,
		direction:                      true,
		transmissionMode:               Acknowledged,
		crcFlag:                        true,
		LargeFileFlag:                  true,
		pduDataFieldLength:             1234,
		segmentationControl:            true,
		lengthEntityID:                 4,
		segmentMetadataFlag:            true,
		lenghTransactionSequenceNumber: 4,
		sourceEntityID:                 0x1234,
		transactionSequenceNumber:      0x5678,
		destinationEntityID:            0x9ABC,
	}

	// Serialize
	b, err := original.ToBytes(original.pduDataFieldLength)
	if err != nil {
		t.Fatalf("ToBytes failed: %v", err)
	}

	// Deserialize
	var decoded ProtocolDataUnitHeader
	if err := decoded.FromBytes(b); err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}

	// Compare fields
	if decoded.version != original.version {
		t.Errorf("version mismatch: got %v, want %v", decoded.version, original.version)
	}
	if decoded.pduType != original.pduType {
		t.Errorf("pduType mismatch: got %v, want %v", decoded.pduType, original.pduType)
	}
	if decoded.direction != original.direction {
		t.Errorf("direction mismatch: got %v, want %v", decoded.direction, original.direction)
	}
	if decoded.transmissionMode != original.transmissionMode {
		t.Errorf("transmissionMode mismatch: got %v, want %v", decoded.transmissionMode, original.transmissionMode)
	}
	if decoded.crcFlag != original.crcFlag {
		t.Errorf("crcFlag mismatch: got %v, want %v", decoded.crcFlag, original.crcFlag)
	}
	if decoded.LargeFileFlag != original.LargeFileFlag {
		t.Errorf("LargeFileFlag mismatch: got %v, want %v", decoded.LargeFileFlag, original.LargeFileFlag)
	}
	if decoded.pduDataFieldLength != original.pduDataFieldLength {
		t.Errorf("pduDataFieldLength mismatch: got %v, want %v", decoded.pduDataFieldLength, original.pduDataFieldLength)
	}
	if decoded.segmentationControl != original.segmentationControl {
		t.Errorf("segmentationControl mismatch: got %v, want %v", decoded.segmentationControl, original.segmentationControl)
	}
	if decoded.lengthEntityID != original.lengthEntityID {
		t.Errorf("lengthEntityID mismatch: got %v, want %v", decoded.lengthEntityID, original.lengthEntityID)
	}
	if decoded.segmentMetadataFlag != original.segmentMetadataFlag {
		t.Errorf("segmentMetadataFlag mismatch: got %v, want %v", decoded.segmentMetadataFlag, original.segmentMetadataFlag)
	}
	if decoded.lenghTransactionSequenceNumber != original.lenghTransactionSequenceNumber {
		t.Errorf("lenghTransactionSequenceNumber mismatch: got %v, want %v", decoded.lenghTransactionSequenceNumber, original.lenghTransactionSequenceNumber)
	}
	if decoded.sourceEntityID != original.sourceEntityID {
		t.Errorf("sourceEntityID mismatch: got %v, want %v", decoded.sourceEntityID, original.sourceEntityID)
	}
	if decoded.transactionSequenceNumber != original.transactionSequenceNumber {
		t.Errorf("transactionSequenceNumber mismatch: got %v, want %v", decoded.transactionSequenceNumber, original.transactionSequenceNumber)
	}
	if decoded.destinationEntityID != original.destinationEntityID {
		t.Errorf("destinationEntityID mismatch: got %v, want %v", decoded.destinationEntityID, original.destinationEntityID)
	}
}

func TestMetadataPDUContents_ToBytesAndFromBytes_LargeFile(t *testing.T) {
	header := ProtocolDataUnitHeader{
		LargeFileFlag: true,
	}
	original := MetadataPDUContents{
		ClosureRequested:    true,
		ChecksumType:        0x2,
		FileSize:            0x123456789ABCDEF0,
		SourceFileName:      "source.txt",
		DestinationFileName: "dest.txt",
	}

	// Serialize
	b := original.ToBytes(header)

	// Deserialize
	var decoded MetadataPDUContents
	if err := decoded.FromBytes(b, header); err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}

	// Compare fields
	if decoded.ClosureRequested != original.ClosureRequested {
		t.Errorf("ClosureRequested mismatch: got %v, want %v", decoded.ClosureRequested, original.ClosureRequested)
	}
	if decoded.ChecksumType != original.ChecksumType {
		t.Errorf("ChecksumType mismatch: got %v, want %v", decoded.ChecksumType, original.ChecksumType)
	}
	if decoded.FileSize != original.FileSize {
		t.Errorf("FileSize mismatch: got %v, want %v", decoded.FileSize, original.FileSize)
	}
	if decoded.SourceFileName != original.SourceFileName {
		t.Errorf("SourceFileName mismatch: got %v, want %v", decoded.SourceFileName, original.SourceFileName)
	}
	if decoded.DestinationFileName != original.DestinationFileName {
		t.Errorf("DestinationFileName mismatch: got %v, want %v", decoded.DestinationFileName, original.DestinationFileName)
	}
}

func TestMetadataPDUContents_ToBytesAndFromBytes_SmallFile(t *testing.T) {
	header := ProtocolDataUnitHeader{
		LargeFileFlag: false,
	}
	original := MetadataPDUContents{
		ClosureRequested:    false,
		ChecksumType:        0x1,
		FileSize:            0x12345678,
		SourceFileName:      "foo",
		DestinationFileName: "bar",
	}

	// Serialize
	b := original.ToBytes(header)

	// Deserialize
	var decoded MetadataPDUContents
	if err := decoded.FromBytes(b, header); err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}

	// Compare fields
	if decoded.FileSize != original.FileSize {
		t.Errorf("FileSize mismatch for small file: got %v, want %v", decoded.FileSize, original.FileSize)
	}
	if decoded.SourceFileName != original.SourceFileName {
		t.Errorf("SourceFileName mismatch: got %v, want %v", decoded.SourceFileName, original.SourceFileName)
	}
	if decoded.DestinationFileName != original.DestinationFileName {
		t.Errorf("DestinationFileName mismatch: got %v, want %v", decoded.DestinationFileName, original.DestinationFileName)
	}
}
