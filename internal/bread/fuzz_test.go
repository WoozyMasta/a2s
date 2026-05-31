package bread

import "testing"

func FuzzReader(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8})
	f.Fuzz(func(t *testing.T, data []byte) {
		reader := NewReader(data)
		_, _ = reader.Byte()

		reader.Reset(data)
		_, _ = reader.Bool()
		reader.Reset(data)
		_, _ = reader.Uint16()
		reader.Reset(data)
		_, _ = reader.Uint32()
		reader.Reset(data)
		_, _ = reader.Int32()
		reader.Reset(data)
		_, _ = reader.Uint64()
		reader.Reset(data)
		_, _ = reader.Float32()
		reader.Reset(data)
		_, _ = reader.Float64()
		reader.Reset(data)
		_, _ = reader.String()
		reader.Reset(data)
		_, _ = reader.BytesPage()
		reader.Reset(data)
		_, _ = reader.StringLen(0)
		reader.Reset(data)
		_, _ = reader.StringLen(len(data))
		reader.Reset(data)
		_, _ = reader.Duration32()
		reader.Reset(data)
		_, _ = reader.Duration64()

		reader.Reset(data)
		_ = reader.Pos()
		_ = reader.Len()
	})
}

func FuzzEscapeSequences(f *testing.F) {
	f.Add([]byte{0x01, 0x01, 0x01, 0x02, 0x01, 0x03})
	f.Fuzz(func(t *testing.T, data []byte) {
		decoded := EscapeSequences(data)
		appended := AppendEscapeSequences(nil, data)
		if string(decoded) != string(appended) {
			t.Fatalf("escape implementations differ: %x != %x", decoded, appended)
		}
	})
}
