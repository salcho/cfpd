package messages

import (
	"hash/crc32"
)

type ChecksumType byte

// checksums 4-14 are reserved for future use, see https://sanaregistry.org/r/checksum_identifiers/?sort=checksumid
const (
	Modular    ChecksumType = 0
	Proximity1 ChecksumType = 1
	CRC32C     ChecksumType = 2
	CRC32      ChecksumType = 3
	Null       ChecksumType = 15
)

func GetChecksumAlgorithm(checksumType ChecksumType) func([]byte) uint32 {
	switch checksumType {
	case Modular:
		return ModularChecksum
	case Proximity1:
		return Proximity1Checksum
	case CRC32C:
		return Crc32c
	case CRC32:
		return Crc32
	case Null:
		return NullChecksum
	default:
		return nil
	}
}

func ModularChecksum(data []byte) uint32 {
	var sum byte = 0
	for _, b := range data {
		sum += b
	}
	return uint32(sum) % (2 ^ 32)
}

func Crc32(data []byte) uint32 {
	crc32q := crc32.MakeTable(crc32.IEEE)
	return crc32.Checksum(data, crc32q)
}

func Crc32c(data []byte) uint32 {
	crc32c := crc32.MakeTable(crc32.Castagnoli)
	return crc32.Checksum(data, crc32c)
}

func Proximity1Checksum(data []byte) uint32 {
	crc32p1 := crc32.MakeTable(0xA0100500) // reversed form of polynomial 0xA00805, see https://sanaregistry.org/r/checksum_identifiers/
	return crc32.Checksum(data, crc32p1)
}

func NullChecksum(data []byte) uint32 {
	return 0
}
