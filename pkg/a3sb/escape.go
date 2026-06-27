package a3sb

// AppendEscapeSequences appends A3SB-escaped data to dst.
//
//	0x01 -> {0x01, 0x01}
//	0x00 -> {0x01, 0x02}
//	0xFF -> {0x01, 0x03}
//
// This encoder is the inverse of appendDecodedEscapeSequences.
func AppendEscapeSequences(dst []byte, data []byte) []byte {
	for _, value := range data {
		switch value {
		case 0x01:
			dst = append(dst, 0x01, 0x01)

		case 0x00:
			dst = append(dst, 0x01, 0x02)

		case 0xFF:
			dst = append(dst, 0x01, 0x03)

		default:
			dst = append(dst, value)
		}
	}

	return dst
}

// appendDecodedEscapeSequences appends decoded A3SB escape sequences to dst.
//
// Unknown and incomplete escape sequences are preserved byte-for-byte.
// When dst is nil, data is decoded in place and must be owned by the caller.
func appendDecodedEscapeSequences(dst []byte, data []byte) []byte {
	if dst == nil {
		return decodeEscapeSequencesInPlace(data)
	}

	for i := 0; i < len(data); i++ {
		if data[i] == 0x01 && i+1 < len(data) {
			switch data[i+1] {
			case 0x01:
				dst = append(dst, 0x01)
				i++

			case 0x02:
				dst = append(dst, 0x00)
				i++

			case 0x03:
				dst = append(dst, 0xFF)
				i++

			default:
				dst = append(dst, data[i])
			}
		} else {
			dst = append(dst, data[i])
		}
	}

	return dst
}

// decodeEscapeSequencesInPlace decodes data in its existing backing buffer.
// The caller must own data and must not need the encoded representation after this call.
// The returned slice may be shorter than data.
func decodeEscapeSequencesInPlace(data []byte) []byte {
	write := 0

	for read := 0; read < len(data); read++ {
		value := data[read]
		if value == 0x01 && read+1 < len(data) {
			switch data[read+1] {
			case 0x01:
				value = 0x01
				read++

			case 0x02:
				value = 0x00
				read++

			case 0x03:
				value = 0xFF
				read++
			}
		}

		data[write] = value
		write++
	}

	return data[:write]
}
