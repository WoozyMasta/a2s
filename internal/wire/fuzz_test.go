package wire

import "testing"

func FuzzDecoderCursorInvariants(f *testing.F) {
	operations := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	f.Add([]byte{}, operations)
	f.Add([]byte{0}, operations)
	f.Add([]byte("unterminated"), operations)
	f.Add(fuzzInfoLikeData(), operations)
	f.Add([]byte{1, 0, 'm', 'o', 'd', 'e', 0, 'c', 'o', 'o', 'p', 0}, operations)
	f.Add([]byte{2, 0, 0, 0, 0, 0, 1, 0xFF}, operations)

	f.Fuzz(func(t *testing.T, data, operations []byte) {
		decoder := NewDecoder(data)
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("decoder panicked: %v", recovered)
			}
		}()

		assertDecoderState(t, &decoder, len(data))

		for step, operation := range operations {
			before := decoder.Offset()
			var err error

			switch operation % 11 {
			case 0:
				_, err = decoder.Byte()
			case 1:
				_, err = decoder.Bytes(fuzzLength(operation, len(data)))
			case 2:
				_, err = decoder.Uint16()
			case 3:
				_, err = decoder.Uint32()
			case 4:
				_, err = decoder.Uint64()
			case 5:
				_, err = decoder.Int32()
			case 6:
				_, err = decoder.Float32()
			case 7:
				_, err = decoder.CStringBytes()
			case 8:
				_, err = decoder.CString()
			case 9:
				_, err = decoder.FixedString(fuzzLength(operation, len(data)))
			case 10:
				tail := decoder.Tail()
				if len(tail) != decoder.Remaining() {
					t.Fatalf("step %d: Tail() length = %d, Remaining() = %d", step, len(tail), decoder.Remaining())
				}
			}

			if err != nil && decoder.Offset() != before {
				t.Fatalf("step %d: failed operation moved offset from %d to %d", step, before, decoder.Offset())
			}
			assertDecoderState(t, &decoder, len(data))
		}
	})
}

func assertDecoderState(t *testing.T, decoder *Decoder, dataLength int) {
	t.Helper()

	offset := decoder.Offset()
	if offset < 0 || offset > dataLength {
		t.Fatalf("offset = %d, want range [0, %d]", offset, dataLength)
	}

	wantRemaining := dataLength - offset
	if got := decoder.Remaining(); got != wantRemaining {
		t.Fatalf("remaining = %d, want %d", got, wantRemaining)
	}
	if got := len(decoder.Tail()); got != wantRemaining {
		t.Fatalf("tail length = %d, want %d", got, wantRemaining)
	}
	if decoder.Empty() != (wantRemaining == 0) {
		t.Fatalf("empty = %v, want %v", decoder.Empty(), wantRemaining == 0)
	}
}

func fuzzLength(value byte, dataLength int) int {
	switch value & 3 {
	case 0:
		return int(int8(value))
	case 1:
		return dataLength
	case 2:
		return dataLength + 1 + int(value>>2)
	default:
		return int(value)
	}
}

func fuzzInfoLikeData() []byte {
	data := []byte{17}
	for _, value := range []string{"server", "map", "folder", "game"} {
		data = append(data, value...)
		data = append(data, 0)
	}
	data = append(data, 0xD2, 0x04, 1, 16, 2, 'd', 'w', 1, 0)
	data = append(data, "1.0"...)
	data = append(data, 0)
	return data
}
