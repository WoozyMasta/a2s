package keywords

import (
	"fmt"
	"testing"

	"github.com/woozymasta/a2s/pkg/keywords/types"
)

func TestDayzKeywords(t *testing.T) {
	kw := []string{
		"unknown", "battleye", "no3rd", "shard001", "lqs0", "port777",
		"etm2.300000", "entm6.800000", "isDLC", "13:38",
	}

	data := ParseDayZ(kw)
	fmt.Println(data)

	if len(data.Unknowns) != 1 {
		t.Error("Wrong unknown keywords count")
	}

	if !data.BattlEye {
		t.Error("Battleye must be enabled")
	}

	if data.GamePort != 777 {
		t.Error("Wrong game port")
	}

	if data.Shard != "001" {
		t.Error("Wrong shard")
	}

	if data.TimeNightAccel != 6.8 {
		t.Error("Wrong night accel")
	}

	if float64(data.Time) != 4.908e+13 {
		t.Error("Wrong night accel")
	}
}

func TestArmaKeywords(t *testing.T) {
	kw := []string{
		"bt", "r218", "n150779", "s3", "i1", "mf", "lf", "vt", "dt", "tzeus", "g65541",
		"h285fa806", "oDE", "f0", "c-25--25", "pw", "e15", "j0", "k0", "x1", "z1",
	}

	data := ParseArma3(kw)
	fmt.Println(data)

	if len(data.Unknowns) != 2 {
		t.Error("Wrong unknown keywords count")
	}

	if !data.BattlEye {
		t.Error("Battleye must be enabled")
	}

	if data.Language.String() != "Czech" {
		t.Error("Wrong language")
	}

	if data.Country != "DE" {
		t.Error("Wrong country")
	}

	if data.GameType != types.GameTZeus {
		t.Error("Wrong game type")
	}

	if float64(data.TimeLeft) != 9e+11 {
		t.Error("Wrong night accel")
	}
}

func TestAnyKeywords(t *testing.T) {
	kwA := []string{
		"bt", "r218", "n150779", "s3", "i1", "mf", "lf", "vt", "dt", "tzeus", "g65541",
		"h285fa806", "f0", "c-2147483648--2147483648", "pw", "e15", "j0", "k0",
	}

	dataA, err := Parse(107410, kwA)
	if err != nil {
		t.Errorf("Cant get data for arma: %v", err)
	}

	switch dataA.(type) {
	case *Arma3:
		break
	case *DayZ:
		t.Error("Return dayz, but expect arma")
	default:
		t.Error("Return unknown, but expect arma")
	}

	kwD := []string{
		"unknown", "battleye", "no3rd", "shard001", "lqs0", "port777",
		"etm2.300000", "entm6.800000", "isDLC", "13:38",
	}

	dataD, err := Parse(1024020, kwD)
	if err != nil {
		t.Errorf("Cant get data for dayz: %v", err)
	}
	if _, ok := dataD.(*DayZ); !ok {
		t.Error("Return unexpected type, but expect dayz")
	}

	kwX := []string{"some"}
	_, err = Parse(1337, kwX)
	if err == nil {
		t.Error("Expect error, but found response")
	}
}

func TestCoordinates(t *testing.T) {
	if lon, lat := parseCoordinates("-1-1"); lon != -1 || lat != 1 {
		t.Errorf("Unexpected coordinates, want [-1:1] but return [%d:%d]", lon, lat)
	}

	if lon, lat := parseCoordinates("-1--1"); lon != -1 || lat != -1 {
		t.Errorf("Unexpected coordinates, want [-1:-1] but return [%d:%d]", lon, lat)
	}

	if lon, lat := parseCoordinates("1-1"); lon != 1 || lat != 1 {
		t.Errorf("Unexpected coordinates, want [1:-1] but return [%d:%d]", lon, lat)
	}

	if lon, lat := parseCoordinates("1--1"); lon != 1 || lat != -1 {
		t.Errorf("Unexpected coordinates, want [1:-1] but return [%d:%d]", lon, lat)
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"true value", "t", true},
		{"false value empty", "", false},
		{"false value f", "f", false},
		{"false value false", "false", false},
		{"false value 0", "0", false},
		{"false value 1", "1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseBool(tt.input)
			if result != tt.expected {
				t.Errorf("parseBool(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseUint8(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint8
	}{
		{"zero", "0", 0},
		{"one", "1", 1},
		{"max value", "255", 255},
		{"overflow", "256", 0},
		{"large overflow", "1000", 0},
		{"negative", "-1", 0},
		{"invalid", "abc", 0},
		{"empty", "", 0},
		{"float", "12.5", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseUint8(tt.input)
			if result != tt.expected {
				t.Errorf("ParseUint8(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseUint16(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint16
	}{
		{"zero", "0", 0},
		{"one", "1", 1},
		{"max value", "65535", 65535},
		{"overflow", "65536", 0},
		{"large overflow", "100000", 0},
		{"negative", "-1", 0},
		{"invalid", "abc", 0},
		{"empty", "", 0},
		{"float", "12.5", 0},
		{"port 777", "777", 777},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseUint16(tt.input)
			if result != tt.expected {
				t.Errorf("ParseUint16(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseUint32(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint32
	}{
		{"zero", "0", 0},
		{"one", "1", 1},
		{"max value", "4294967295", 4294967295},
		{"overflow", "4294967296", 0},
		{"large overflow", "9999999999", 0},
		{"negative", "-1", 0},
		{"invalid", "abc", 0},
		{"empty", "", 0},
		{"float", "12.5", 0},
		{"version 218", "218", 218},
		{"build 150779", "150779", 150779},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseUint32(tt.input)
			if result != tt.expected {
				t.Errorf("parseUint32(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"zero", "0", 0.0},
		{"one", "1", 1.0},
		{"negative", "-1", -1.0},
		{"decimal", "2.3", 2.3},
		{"negative decimal", "-2.3", -2.3},
		{"large decimal", "2.300000", 2.3},
		{"night accel", "6.800000", 6.8},
		{"invalid", "abc", 0.0},
		{"empty", "", 0.0},
		{"scientific", "1e5", 100000.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFloat64(tt.input)
			if result != tt.expected {
				t.Errorf("parseFloat64(%q) = %f, want %f", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseCoordinatesTable(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedLon int32
		expectedLat int32
	}{
		{"both positive", "1-1", 1, 1},
		{"lon negative", "-1-1", -1, 1},
		{"lat negative", "1--1", 1, -1},
		{"both negative", "-1--1", -1, -1},
		{"large values", "2147483647--2147483648", 2147483647, -2147483648},
		{"zero", "0-0", 0, 0},
		{"invalid format", "abc", 0, 0},
		{"empty", "", 0, 0},
		{"single number", "123", 0, 0},
		{"no dash", "123456", 0, 0},
		{"multiple dashes", "1-2-3", 1, 2}, // fmt.Sscanf reads only first two numbers
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lon, lat := parseCoordinates(tt.input)
			if lon != tt.expectedLon || lat != tt.expectedLat {
				t.Errorf("parseCoordinates(%q) = [%d:%d], want [%d:%d]", tt.input, lon, lat, tt.expectedLon, tt.expectedLat)
			}
		})
	}
}

func TestParseTable(t *testing.T) {
	tests := []struct {
		name        string
		appID       uint64
		keywords    []string
		expectError bool
		expectType  string
	}{
		{
			name:        "Arma3 valid",
			appID:       107410,
			keywords:    []string{"bt", "r218"},
			expectError: false,
			expectType:  "*keywords.Arma3",
		},
		{
			name:        "DayZ valid",
			appID:       1024020,
			keywords:    []string{"battleye", "shard001"},
			expectError: false,
			expectType:  "*keywords.DayZ",
		},
		{
			name:        "unsupported appID",
			appID:       1337,
			keywords:    []string{"some"},
			expectError: true,
			expectType:  "",
		},
		{
			name:        "empty keywords",
			appID:       107410,
			keywords:    []string{},
			expectError: false,
			expectType:  "*keywords.Arma3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.appID, tt.keywords)
			if tt.expectError {
				if err == nil {
					t.Errorf("Parse() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("Parse() returned nil result")
				return
			}

			gotType := fmt.Sprintf("%T", result)
			if gotType != tt.expectType {
				t.Errorf("Parse() returned type %s, want %s", gotType, tt.expectType)
			}
		})
	}
}

func BenchmarkParseDayZ(b *testing.B) {
	kw := []string{
		"unknown", "battleye", "no3rd", "shard001", "lqs0", "port777",
		"etm2.300000", "entm6.800000", "isDLC", "13:38",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseDayZ(kw)
	}
}

func BenchmarkParseArma3(b *testing.B) {
	kw := []string{
		"bt", "r218", "n150779", "s3", "i1", "mf", "lf", "vt", "dt", "tzeus", "g65541",
		"h285fa806", "oDE", "f0", "c-25--25", "pw", "e15", "j0", "k0", "x1", "z1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseArma3(kw)
	}
}

func BenchmarkParse(b *testing.B) {
	kwA := []string{
		"bt", "r218", "n150779", "s3", "i1", "mf", "lf", "vt", "dt", "tzeus", "g65541",
		"h285fa806", "f0", "c-2147483648--2147483648", "pw", "e15", "j0", "k0",
	}

	kwD := []string{
		"unknown", "battleye", "no3rd", "shard001", "lqs0", "port777",
		"etm2.300000", "entm6.800000", "isDLC", "13:38",
	}

	b.Run("Arma3", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Parse(107410, kwA)
		}
	})

	b.Run("DayZ", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Parse(1024020, kwD)
		}
	})
}

func BenchmarkParseCoordinates(b *testing.B) {
	coords := []string{
		"1-1", "-1-1", "1--1", "-1--1",
		"2147483647--2147483648", "0-0", "1-2-3",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, coord := range coords {
			_, _ = parseCoordinates(coord)
		}
	}
}

func BenchmarkParseUint8(b *testing.B) {
	values := []string{"0", "1", "255", "256", "777", "abc"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, val := range values {
			_ = ParseUint8(val)
		}
	}
}

func BenchmarkParseUint16(b *testing.B) {
	values := []string{"0", "1", "65535", "777", "100000", "abc"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, val := range values {
			_ = ParseUint16(val)
		}
	}
}

func BenchmarkParseUint32(b *testing.B) {
	values := []string{"0", "1", "218", "150779", "4294967295", "4294967296"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, val := range values {
			_ = parseUint32(val)
		}
	}
}

func BenchmarkParseFloat64(b *testing.B) {
	values := []string{"0", "1", "2.3", "6.800000", "-1.5", "1e5", "abc"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, val := range values {
			_ = parseFloat64(val)
		}
	}
}
