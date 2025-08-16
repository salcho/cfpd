package messages

import (
	"testing"
)

func TestMetadataPDUContents_ToBytesAndFromBytes(t *testing.T) {
	// Setup a sample header and contents
	header := ProtocolDataUnitHeader{
		largeFileFlag: true,
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
	decoded = decoded.FromBytes(data, header)

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
		largeFileFlag: false,
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
	decoded = decoded.FromBytes(data, header)

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
