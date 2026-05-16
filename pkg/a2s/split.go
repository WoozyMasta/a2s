package a2s

import "encoding/binary"

const (
	// splitMin is the minimum size needed to inspect a split header.
	splitMin = 9

	// srcSplitHeader is the size of a standard Source split header.
	srcSplitHeader = 12

	// srcSplitNoSize is the size of a Source split header without split size.
	srcSplitNoSize = 10

	// goldSrcHeader is the size of a GoldSource split header.
	goldSrcHeader = 9

	// splitSizeOff is the offset of the Source split size field.
	splitSizeOff = 10

	// splitSizeMax limits the advertised Source split size.
	splitSizeMax = 4096

	// unpackProbeMax limits decompression probing for ambiguous headers.
	unpackProbeMax = 32 * 1024 * 1024

	// splitPacketCountMax is the largest count representable by split headers.
	splitPacketCountMax = 255

	// splitResponseSizeMax bounds raw and assembled split response data.
	splitResponseSizeMax = 16 * 1024 * 1024
)

// splitHeaderInfo contains metadata about a split packet.
type splitHeaderInfo struct {
	count        int    // Total number of packets in the response.
	index        int    // Current packet number.
	headerSize   int    // Size of the base header without compression metadata.
	dataOff      int    // Offset of the data in this packet.
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

	// Compression metadata belongs to fragment zero.
	// Other fragments carry only the base split header and must be retained from that offset.
	if (packetID & 0x80000000) != 0 {
		info.compressed = true
		if info.index == 0 {
			if len(data) < baseHeaderSize+8 {
				return splitHeaderInfo{}, ErrMultiPacket
			}

			info.unpackedSize = binary.LittleEndian.Uint32(data[baseHeaderSize : baseHeaderSize+4])
			info.crc = binary.LittleEndian.Uint32(data[baseHeaderSize+4 : baseHeaderSize+8])
			info.dataOff = baseHeaderSize + 8

			// Some servers omit the split-size field,
			// producing a 10-byte header where the standard Source header is 12 bytes.
			if !useGold &&
				baseHeaderSize == srcSplitHeader &&
				info.unpackedSize > unpackProbeMax &&
				len(data) >= 18 {
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
	}

	if info.count <= 0 || info.count > splitPacketCountMax || info.index < 0 || info.index >= info.count {
		return splitHeaderInfo{}, ErrMultiPacket
	}

	return info, nil
}

// validateSplitFragment checks metadata shared by every fragment in a response.
func validateSplitFragment(expected, actual splitHeaderInfo) error {
	if actual.id != expected.id ||
		actual.count != expected.count ||
		actual.compressed != expected.compressed ||
		actual.goldSrc != expected.goldSrc ||
		actual.index < 0 || actual.index >= expected.count {
		return ErrMultiPacketInconsistent
	}

	return nil
}
