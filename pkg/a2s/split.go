package a2s

import "encoding/binary"

// splitHeaderInfo contains metadata about a split packet.
type splitHeaderInfo struct {
	packetCount      int    // The total number of packets in the response.
	currentPacket    int    // The current packet number.
	baseHeaderSize   int    // The size of the base header.
	dataOffset       int    // The offset of the data in the packet.
	packetID         uint32 // The ID of the packet.
	decompressedSize uint32 // The size of the decompressed data.
	crc              uint32 // The CRC of the decompressed data.
	compressed       bool   // Whether the packet is compressed.
	useGoldSrc       bool   // Whether the packet is from the GoldSource engine.
}

// parseSplitHeader parses the split header from a packet.
// Returns the split header info and an error if the packet is not a split packet.
func parseSplitHeader(data []byte) (splitHeaderInfo, error) {
	if len(data) < 9 {
		return splitHeaderInfo{}, ErrMultiPacket
	}

	packetID := binary.LittleEndian.Uint32(data[4:8])
	countSrc := 0
	numberSrc := 0
	if len(data) >= 10 {
		countSrc = int(data[8])
		numberSrc = int(data[9])
	}

	countGold := int(data[8] & 0x0F)
	numberGold := int((data[8] & 0xF0) >> 4)

	isSource := false
	if len(data) >= 13 && binary.LittleEndian.Uint32(data[9:13]) == singlePacket {
		isSource = false
	} else if len(data) >= 16 && binary.LittleEndian.Uint32(data[12:16]) == singlePacket {
		isSource = true
	} else if len(data) >= 12 {
		splitSize := binary.LittleEndian.Uint16(data[10:12])
		if splitSize > 0 && splitSize <= 4096 {
			isSource = true
		}
	}

	packetCount := countGold
	currentPacket := numberGold
	baseHeaderSize := 9
	useGold := !isSource
	if isSource {
		packetCount = countSrc
		currentPacket = numberSrc
		baseHeaderSize = 12
	}

	info := splitHeaderInfo{
		packetID:       packetID,
		packetCount:    packetCount,
		currentPacket:  currentPacket,
		baseHeaderSize: baseHeaderSize,
		dataOffset:     baseHeaderSize,
		useGoldSrc:     useGold,
	}

	if (packetID & 0x80000000) != 0 {
		if len(data) < baseHeaderSize+8 {
			return splitHeaderInfo{}, ErrMultiPacket
		}
		info.compressed = true
		info.decompressedSize = binary.LittleEndian.Uint32(data[baseHeaderSize : baseHeaderSize+4])
		info.crc = binary.LittleEndian.Uint32(data[baseHeaderSize+4 : baseHeaderSize+8])
		info.dataOffset = baseHeaderSize + 8

		// Some servers omit the split-size field (base header 10 instead of 12).
		if !useGold && baseHeaderSize == 12 && info.decompressedSize > 32*1024*1024 && len(data) >= 18 {
			altSize := binary.LittleEndian.Uint32(data[10:14])
			altCRC := binary.LittleEndian.Uint32(data[14:18])
			if altSize > 0 && altSize <= 32*1024*1024 {
				info.decompressedSize = altSize
				info.crc = altCRC
				info.baseHeaderSize = 10
				info.dataOffset = 18
			}
		}
	}

	if info.packetCount <= 0 {
		return splitHeaderInfo{}, ErrMultiPacket
	}

	return info, nil
}

// readPacketNumber reads the packet number from a packet.
// Returns the packet number.
func (s splitHeaderInfo) readPacketNumber(data []byte) int {
	if s.useGoldSrc {
		return int((data[8] & 0xF0) >> 4)
	}

	return int(data[9])
}
