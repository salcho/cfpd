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
	data, err := original.ToBytes(header)
	if err != nil {
		t.Fatalf("ToBytes failed: %v", err)
	}

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

	data, err := original.ToBytes(header)
	if err != nil {
		t.Fatalf("ToBytes failed: %v", err)
	}

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
		PduType:                        FileDirective,
		direction:                      true,
		transmissionMode:               Acknowledged,
		crcFlag:                        true,
		LargeFileFlag:                  true,
		pduDataFieldLength:             1234,
		segmentationControl:            false,
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
	hLen, err := decoded.FromBytes(b)
	if err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}
	if hLen != len(b) {
		t.Fatalf("FromBytes returned length %d, expected %d", hLen, len(b))
	}

	// Compare fields
	if decoded.version != original.version {
		t.Errorf("version mismatch: got %v, want %v", decoded.version, original.version)
	}
	if decoded.PduType != original.PduType {
		t.Errorf("pduType mismatch: got %v, want %v", decoded.PduType, original.PduType)
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
	b, err := original.ToBytes(header)
	if err != nil {
		t.Fatalf("ToBytes failed: %v", err)
	}

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
	b, err := original.ToBytes(header)
	if err != nil {
		t.Fatalf("ToBytes failed: %v", err)
	}

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

func TestFileDirectivePDU_ToBytesAndFromBytes(t *testing.T) {
	// 1. Setup
	originalHeader := ProtocolDataUnitHeader{
		version:                        1,
		PduType:                        FileDirective,
		direction:                      false,
		transmissionMode:               Unacknowledged,
		crcFlag:                        false,
		LargeFileFlag:                  false,
		pduDataFieldLength:             0, // This will be set by ToBytes
		segmentationControl:            false,
		lengthEntityID:                 2,
		segmentMetadataFlag:            false,
		lenghTransactionSequenceNumber: 2,
		sourceEntityID:                 10,
		transactionSequenceNumber:      100,
		destinationEntityID:            20,
	}

	originalPDU := FileDirectivePDU{
		Header:  originalHeader,
		DirCode: MetadataPDU,
		Data:    []byte{0x01, 0x02, 0x03, 0x04},
	}

	// The data field length is 1 byte for the directive code plus the length of the data
	dataFieldLength := int16(1 + len(originalPDU.Data))

	// 2. Serialize
	pduBytes := originalPDU.ToBytes(dataFieldLength)
	if pduBytes == nil {
		t.Fatal("ToBytes returned nil")
	}

	// 3. Deserialize
	var decodedPDU FileDirectivePDU
	// We need to manually set the header's data field length before deserializing the rest
	// because FromBytes relies on it to know how much data to read.
	decodedPDU.Header.pduDataFieldLength = dataFieldLength
	if err := decodedPDU.FromBytes(pduBytes); err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}

	// 4. Compare
	// The header's pduDataFieldLength is set during serialization, so we update the original to match
	originalPDU.Header.pduDataFieldLength = dataFieldLength

	if decodedPDU.Header.version != originalPDU.Header.version {
		t.Errorf("Header.version mismatch: got %d, want %d", decodedPDU.Header.version, originalPDU.Header.version)
	}
	if decodedPDU.Header.pduDataFieldLength != originalPDU.Header.pduDataFieldLength {
		t.Errorf("Header.pduDataFieldLength mismatch: got %d, want %d", decodedPDU.Header.pduDataFieldLength, originalPDU.Header.pduDataFieldLength)
	}
	if decodedPDU.DirCode != originalPDU.DirCode {
		t.Errorf("DirCode mismatch: got %v, want %v", decodedPDU.DirCode, originalPDU.DirCode)
	}

	if len(decodedPDU.Data) != len(originalPDU.Data) {
		t.Fatalf("Data length mismatch: got %d, want %d", len(decodedPDU.Data), len(originalPDU.Data))
	}

	for i := range originalPDU.Data {
		if decodedPDU.Data[i] != originalPDU.Data[i] {
			t.Errorf("Data mismatch at index %d: got %x, want %x", i, decodedPDU.Data[i], originalPDU.Data[i])
		}
	}
}

func TestMetadataPDUContents_WithOptions(t *testing.T) {
	// 1. Setup
	header := ProtocolDataUnitHeader{
		LargeFileFlag: true,
	}
	original := MetadataPDUContents{
		ClosureRequested:    true,
		ChecksumType:        0x1,
		FileSize:            1024,
		SourceFileName:      "src.dat",
		DestinationFileName: "dest.dat",
		MessagesToUser: []Message{
			&DirectoryListingRequest{
				DirToList:     "foo",
				PathToRespond: "bar",
			},
			&DirectoryListingRequest{
				DirToList:     "caz",
				PathToRespond: "qwe",
			},
		},
	}

	// 2. Serialize
	data, err := original.ToBytes(header)
	if err != nil {
		t.Fatalf("ToBytes failed: %v", err)
	}

	// 3. Deserialize
	var decoded MetadataPDUContents
	if err := decoded.FromBytes(data, header); err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}

	// 4. Compare basic fields
	if decoded.SourceFileName != original.SourceFileName {
		t.Errorf("SourceFileName mismatch: got %v, want %v", decoded.SourceFileName, original.SourceFileName)
	}
	if decoded.FileSize != original.FileSize {
		t.Errorf("FileSize mismatch: got %v, want %v", decoded.FileSize, original.FileSize)
	}

	// 5. Compare Options
	if len(decoded.MessagesToUser) != len(original.MessagesToUser) {
		t.Fatalf("MessagesToUser length mismatch: got %d, want %d", len(decoded.MessagesToUser), len(original.MessagesToUser))
	}

	for i, originalOpt := range original.MessagesToUser {
		decodedOpt := decoded.MessagesToUser[i]
		if originalOpt.GetMessageType() != decodedOpt.GetMessageType() {
			t.Errorf("Option %d Type mismatch: got %v, want %v", i, decodedOpt.GetMessageType(), originalOpt.GetMessageType())
		}
	}
}
