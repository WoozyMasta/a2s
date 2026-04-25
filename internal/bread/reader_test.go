package bread

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
	"time"
)

func TestReaderState(t *testing.T) {
	reader := NewReader([]byte{1, 2, 3})
	if got, want := reader.Pos(), 0; got != want {
		t.Fatalf("Pos() = %d, want %d", got, want)
	}
	if got, want := reader.Len(), 3; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}

	if _, err := reader.Byte(); err != nil {
		t.Fatalf("Byte() error = %v", err)
	}
	if got, want := reader.Pos(), 1; got != want {
		t.Fatalf("Pos() after Byte() = %d, want %d", got, want)
	}
	if got, want := reader.Len(), 2; got != want {
		t.Fatalf("Len() after Byte() = %d, want %d", got, want)
	}

	reader.Reset([]byte{4, 5})
	if got, want := reader.Pos(), 0; got != want {
		t.Fatalf("Pos() after Reset() = %d, want %d", got, want)
	}
	if got, want := reader.Len(), 2; got != want {
		t.Fatalf("Len() after Reset() = %d, want %d", got, want)
	}
}

func TestReaderByte(t *testing.T) {
	reader := NewReader([]byte{0xAB})
	if got, err := reader.Byte(); err != nil || got != 0xAB {
		t.Fatalf("Byte() = 0x%X, %v; want 0xAB, nil", got, err)
	}
	if _, err := reader.Byte(); err != ErrUnderflow {
		t.Fatalf("Byte() error = %v, want %v", err, ErrUnderflow)
	}
}

func TestReaderBool(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
		err  error
	}{
		{name: "true", data: []byte{1}, want: true},
		{name: "false", data: []byte{0}},
		{name: "invalid", data: []byte{2}, err: ErrBool},
		{name: "underflow", err: ErrUnderflow},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewReader(test.data).Bool()
			if got != test.want {
				t.Errorf("Bool() = %v, want %v", got, test.want)
			}
			if err != test.err {
				t.Errorf("Bool() error = %v, want %v", err, test.err)
			}
		})
	}
}

func TestReaderIntegers(t *testing.T) {
	t.Run("Uint16", func(t *testing.T) {
		data := []byte{0x34, 0x12}
		got, err := NewReader(data).Uint16()
		if err != nil || got != 0x1234 {
			t.Fatalf("Uint16() = 0x%X, %v; want 0x1234, nil", got, err)
		}
	})

	t.Run("Uint32", func(t *testing.T) {
		data := make([]byte, 4)
		binary.LittleEndian.PutUint32(data, 0x12345678)
		got, err := NewReader(data).Uint32()
		if err != nil || got != 0x12345678 {
			t.Fatalf("Uint32() = 0x%X, %v; want 0x12345678, nil", got, err)
		}
	})

	t.Run("Int32", func(t *testing.T) {
		data := make([]byte, 16)
		for index, value := range []int32{math.MinInt32, -1, 0, math.MaxInt32} {
			binary.LittleEndian.PutUint32(data[index*4:], uint32(value))
		}

		reader := NewReader(data)
		for _, want := range []int32{math.MinInt32, -1, 0, math.MaxInt32} {
			got, err := reader.Int32()
			if err != nil {
				t.Fatalf("Int32() error = %v", err)
			}
			if got != want {
				t.Fatalf("Int32() = %d, want %d", got, want)
			}
		}
	})

	t.Run("Uint64", func(t *testing.T) {
		data := make([]byte, 8)
		binary.LittleEndian.PutUint64(data, 0x1234567890ABCDEF)
		got, err := NewReader(data).Uint64()
		if err != nil || got != 0x1234567890ABCDEF {
			t.Fatalf("Uint64() = 0x%X, %v; want 0x1234567890ABCDEF, nil", got, err)
		}
	})

	tests := []struct {
		name string
		read func(*Reader) error
		size int
	}{
		{name: "Uint16", size: 1, read: func(reader *Reader) error {
			_, err := reader.Uint16()
			return err
		}},
		{name: "Uint32", size: 3, read: func(reader *Reader) error {
			_, err := reader.Uint32()
			return err
		}},
		{name: "Int32", size: 3, read: func(reader *Reader) error {
			_, err := reader.Int32()
			return err
		}},
		{name: "Uint64", size: 7, read: func(reader *Reader) error {
			_, err := reader.Uint64()
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name+" underflow", func(t *testing.T) {
			if err := test.read(NewReader(make([]byte, test.size))); err != ErrUnderflow {
				t.Fatalf("error = %v, want %v", err, ErrUnderflow)
			}
		})
	}
}

func TestReaderFloatsAndDurations(t *testing.T) {
	float32Data := make([]byte, 4)
	binary.LittleEndian.PutUint32(float32Data, math.Float32bits(1.25))
	if got, err := NewReader(float32Data).Float32(); err != nil || got != 1.25 {
		t.Fatalf("Float32() = %v, %v; want 1.25, nil", got, err)
	}

	float64Data := make([]byte, 8)
	binary.LittleEndian.PutUint64(float64Data, math.Float64bits(2.5))
	if got, err := NewReader(float64Data).Float64(); err != nil || got != 2.5 {
		t.Fatalf("Float64() = %v, %v; want 2.5, nil", got, err)
	}

	if got, err := NewReader(float32Data).Duration32(); err != nil || got != 1250*time.Millisecond {
		t.Fatalf("Duration32() = %v, %v; want 1.25s, nil", got, err)
	}
	if got, err := NewReader(float64Data).Duration64(); err != nil || got != 2500*time.Millisecond {
		t.Fatalf("Duration64() = %v, %v; want 2.5s, nil", got, err)
	}

	for _, test := range []struct {
		name string
		read func(*Reader) error
		size int
	}{
		{name: "Float32", size: 3, read: func(reader *Reader) error {
			_, err := reader.Float32()
			return err
		}},
		{name: "Float64", size: 7, read: func(reader *Reader) error {
			_, err := reader.Float64()
			return err
		}},
		{name: "Duration32", size: 3, read: func(reader *Reader) error {
			_, err := reader.Duration32()
			return err
		}},
		{name: "Duration64", size: 7, read: func(reader *Reader) error {
			_, err := reader.Duration64()
			return err
		}},
	} {
		t.Run(test.name+" underflow", func(t *testing.T) {
			if err := test.read(NewReader(make([]byte, test.size))); err != ErrUnderflow {
				t.Fatalf("error = %v, want %v", err, ErrUnderflow)
			}
		})
	}
}

func TestReaderStrings(t *testing.T) {
	reader := NewReader([]byte("hello\x00tail"))
	if got, err := reader.String(); err != nil || got != "hello" {
		t.Fatalf("String() = %q, %v; want hello, nil", got, err)
	}
	if got, err := reader.StringLen(4); err != nil || got != "tail" {
		t.Fatalf("StringLen() = %q, %v; want tail, nil", got, err)
	}

	pageData := []byte("page\x00tail")
	page := NewReader(pageData)
	got, err := page.BytesPage()
	if err != nil || !bytes.Equal(got, []byte("page")) {
		t.Fatalf("BytesPage() = %q, %v; want page, nil", got, err)
	}
	got[0] = 'P'
	if pageData[0] != 'P' {
		t.Fatal("BytesPage() did not return a slice backed by the input")
	}

	if got, err := NewReader([]byte("unterminated")).String(); err != ErrString || got != "" {
		t.Fatalf("String() = %q, %v; want empty string, %v", got, err, ErrString)
	}
	if got, err := NewReader([]byte("unterminated")).BytesPage(); err != ErrString || got != nil {
		t.Fatalf("BytesPage() = %q, %v; want nil, %v", got, err, ErrString)
	}

	if got, err := NewReader([]byte("abc")).StringLen(4); err != ErrUnderflow || got != "" {
		t.Fatalf("StringLen() = %q, %v; want empty string, %v", got, err, ErrUnderflow)
	}
	if got, err := NewReader([]byte("abc")).StringLen(0); err != nil || got != "" {
		t.Fatalf("StringLen(0) = %q, %v; want empty string, nil", got, err)
	}
}
