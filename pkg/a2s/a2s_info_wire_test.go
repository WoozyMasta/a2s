// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"encoding/binary"
	"errors"
	"io"
	"testing"

	"github.com/woozymasta/a2s/internal/wire"
)

func TestReadInfoBool(t *testing.T) {
	for _, test := range []struct {
		name  string
		value byte
		want  bool
		err   error
	}{
		{name: "false", value: 0},
		{name: "true", value: 1, want: true},
		{name: "invalid", value: 2, err: errInfoInvalidBoolean},
	} {
		t.Run(test.name, func(t *testing.T) {
			decoder := wire.NewDecoder([]byte{test.value})
			got, err := readInfoBool(&decoder)
			if got != test.want {
				t.Errorf("readInfoBool() = %v, want %v", got, test.want)
			}
			if !errors.Is(err, test.err) {
				t.Errorf("readInfoBool() error = %v, want %v", err, test.err)
			}
		})
	}
}

func TestDecodeInfoRejectsUnsupportedPacketType(t *testing.T) {
	_, err := DecodeInfo(Packet{Type: ResponseRules})
	if !errors.Is(err, ErrInfoUnsupportedFormat) {
		t.Fatalf("DecodeInfo() error = %v, want ErrInfoUnsupportedFormat", err)
	}
}

func TestParseInfoSourceOptionalTailFailures(t *testing.T) {
	base := benchmarkSourceInfo()
	base = base[:len(base)-1]

	withEDF := func(edf EDF, tail []byte) []byte {
		data := append([]byte(nil), base...)
		data = append(data, byte(edf))
		return append(data, tail...)
	}

	port := make([]byte, 2)
	binary.LittleEndian.PutUint16(port, 27015)

	tests := []struct {
		name string
		data []byte
		want error
	}{
		{
			name: "missing optional EDF",
			data: base,
		},
		{
			name: "missing port",
			data: withEDF(edfPort, nil),
			want: ErrInfoEDFPort,
		},
		{
			name: "missing SteamID",
			data: withEDF(edfSteamID, nil),
			want: ErrInfoEDFSteamID,
		},
		{
			name: "missing SourceTV port",
			data: withEDF(edfSourceTV, nil),
			want: ErrInfoEDFSourceTVPort,
		},
		{
			name: "missing SourceTV name",
			data: withEDF(edfSourceTV, port),
			want: ErrInfoEDFSourceTVName,
		},
		{
			name: "missing keywords terminator",
			data: withEDF(edfKeywords, []byte("one")),
			want: ErrInfoEDFKeywords,
		},
		{
			name: "missing GameID",
			data: withEDF(edfGameID, nil),
			want: ErrInfoEDFGameID,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := DecodeInfo(Packet{Type: ResponseInfo, Payload: test.data})
			if test.want == nil {
				if err != nil {
					t.Fatalf("DecodeInfo() error = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, test.want) {
				t.Fatalf("DecodeInfo() error = %v, want %v", err, test.want)
			}
			if !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Fatalf("DecodeInfo() error = %v, want io.ErrUnexpectedEOF", err)
			}
		})
	}
}
