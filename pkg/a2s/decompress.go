package a2s

import (
	"bytes"
	"compress/bzip2"
	"fmt"
	"hash/crc32"
	"io"
)

// maxDecompressedSize bounds memory allocation for a compressed response.
const maxDecompressedSize = 16 * 1024 * 1024

// decompressBzip2 decompresses a response and verifies its size and CRC.
func decompressBzip2(compressed []byte, size uint32, crc uint32) ([]byte, error) {
	if size > maxDecompressedSize {
		return nil, fmt.Errorf("%w: %d > %d", ErrDecompressSize, size, maxDecompressedSize)
	}

	reader := bzip2.NewReader(bytes.NewReader(compressed))
	decompressed := make([]byte, int(size))

	readBytes, err := io.ReadFull(reader, decompressed)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, ErrDecompressFailed
	}

	if readBytes != int(size) {
		return nil, ErrDecompressSizeMismatch
	}

	if crc32.ChecksumIEEE(decompressed) != crc {
		return nil, ErrDecompressCRC
	}

	return decompressed, nil
}
