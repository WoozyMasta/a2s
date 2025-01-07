package a2s

// TODO: notes for bzip2 decompression
/*
import (
	"bytes"
	"compress/bzip2"
	"encoding/binary"
	"fmt"
	"hash/crc32"
)

func decompressBzip2(resp []byte, n int) ([]byte, error) {
	size := binary.LittleEndian.Uint32(resp[12:16])
	crc := binary.LittleEndian.Uint32(resp[16:20])

	if size > 1024*1024 {
		return nil, fmt.Errorf("bz2 decompressed size exceeds 1 MB")
	}

	compressedData := resp[20:n]
	reader := bzip2.NewReader(bytes.NewReader(compressedData))

	decompressed := make([]byte, size)
	readBytes, err := reader.Read(decompressed)
	if err != nil {
		return nil, fmt.Errorf("bz2 decompression failed: %w", err)
	}

	if uint32(readBytes) != size {
		return nil, fmt.Errorf("bz2 decompressed size mismatch: expected %d, got %d", size, readBytes)
	}

	if crc32.ChecksumIEEE(decompressed) != crc {
		return nil, fmt.Errorf("bz2 CRC32 checksum mismatch")
	}

	return decompressed, nil
}
*/
