package a2s

import "encoding/binary"

const (
	splitMin       = 9                // Minimum size of a split header.
	srcSplitHeader = 12               // Size of a Source split header.
	srcSplitNoSize = 10               // Size of a Source split header without split size.
	goldSrcHeader  = 9                // Size of a GoldSource split header.
	splitSizeOff   = 10               // Offset of the split size in a Source split header.
	splitSizeMax   = 4096             // Max allowed split size.
	unpackProbeMax = 32 * 1024 * 1024 // Max allowed decompressed size for probe.
)

// splitHeaderInfo contains metadata about a split packet.
type splitHeaderInfo struct {
	count        int    // Total number of packets in the response.
	index        int    // Current packet number.
	headerSize   int    // Size of the base header.
	dataOff      int    // Offset of the data in the packet.
	id           uint32 // ID of the packet.
	unpackedSize uint32 // Size of the decompressed data.
	crc          uint32 // CRC of the decompressed data.
	compressed   bool   // Whether the packet is compressed.
	goldSrc      bool   // Whether the packet is from the GoldSource engine.
}

// parseSplitHeader parses the split header from a packet.
// Returns the split header info and an error if the packet is not a split packet.
func parseSplitHeader(data []byte) (splitHeaderInfo, error) {
	if len(data) < splitMin {
		return splitHeaderInfo{}, ErrMultiPacket
	}

	// Get the packet ID.
	packetID := binary.LittleEndian.Uint32(data[4:8])
	countSrc := 0
	numberSrc := 0
	if len(data) >= 10 {
		countSrc = int(data[8])
		numberSrc = int(data[9])
	}

	countGold := int(data[8] & 0x0F)
	numberGold := int((data[8] & 0xF0) >> 4)

	// Check if the packet is from the Source engine.
	isSource := false
	if len(data) >= 13 && binary.LittleEndian.Uint32(data[9:13]) == singlePacket {
		isSource = false
	} else if len(data) >= 16 && binary.LittleEndian.Uint32(data[12:16]) == singlePacket {
		isSource = true
	} else if len(data) >= srcSplitHeader {
		splitSize := binary.LittleEndian.Uint16(data[splitSizeOff:srcSplitHeader])
		if splitSize > 0 && splitSize <= splitSizeMax {
			isSource = true
		}
	}

	// Set packet count and current packet number.
	packetCount := countGold
	currentPacket := numberGold
	baseHeaderSize := goldSrcHeader
	useGold := !isSource
	if isSource {
		packetCount = countSrc
		currentPacket = numberSrc
		baseHeaderSize = srcSplitHeader
	}

	// Create split header info.
	info := splitHeaderInfo{
		id:         packetID,
		count:      packetCount,
		index:      currentPacket,
		headerSize: baseHeaderSize,
		dataOff:    baseHeaderSize,
		goldSrc:    useGold,
	}

	// Check if packet is compressed and set decompressed size and CRC.
	if (packetID & 0x80000000) != 0 {
		if len(data) < baseHeaderSize+8 {
			return splitHeaderInfo{}, ErrMultiPacket
		}

		info.compressed = true
		info.unpackedSize = binary.LittleEndian.Uint32(data[baseHeaderSize : baseHeaderSize+4])
		info.crc = binary.LittleEndian.Uint32(data[baseHeaderSize+4 : baseHeaderSize+8])
		info.dataOff = baseHeaderSize + 8

		// Some servers omit the split-size field (base header 10 instead of 12).
		if !useGold && baseHeaderSize == srcSplitHeader && info.unpackedSize > unpackProbeMax && len(data) >= 18 {
			altSize := binary.LittleEndian.Uint32(data[splitSizeOff : splitSizeOff+4])
			altCRC := binary.LittleEndian.Uint32(data[splitSizeOff+4 : splitSizeOff+8])
			if altSize > 0 && altSize <= unpackProbeMax {
				info.unpackedSize = altSize
				info.crc = altCRC
				info.headerSize = srcSplitNoSize
				info.dataOff = splitSizeOff + 8
			}
		}
	}

	if info.count <= 0 {
		return splitHeaderInfo{}, ErrMultiPacket
	}

	return info, nil
}

// readPacketNumber reads the packet number from a packet.
// Returns the packet number.
func (s splitHeaderInfo) readPacketNumber(data []byte) int {
	if s.goldSrc {
		return int((data[8] & 0xF0) >> 4)
	}

	return int(data[9])
}
