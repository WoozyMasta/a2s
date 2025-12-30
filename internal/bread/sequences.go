package bread

// EscapeSequences decodes A3SBP escape sequences:
//
//	{0x01, 0x01} -> 0x01
//	{0x01, 0x02} -> 0x00
//	{0x01, 0x03} -> 0xFF
func EscapeSequences(data []byte) []byte {
	buf := make([]byte, 0, len(data))

	for i := 0; i < len(data); i++ {
		if data[i] == 0x01 && i+1 < len(data) {
			switch data[i+1] {
			case 0x01:
				buf = append(buf, 0x01)
				i++
			case 0x02:
				buf = append(buf, 0x00)
				i++
			case 0x03:
				buf = append(buf, 0xFF)
				i++
			default:
				buf = append(buf, data[i])
			}
		} else {
			buf = append(buf, data[i])
		}
	}

	return buf
}

// AppendEscapeSequences appends decoded escape sequences to dst.
//
//	{0x01, 0x01} -> 0x01
//	{0x01, 0x02} -> 0x00
//	{0x01, 0x03} -> 0xFF
func AppendEscapeSequences(dst []byte, data []byte) []byte {
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
