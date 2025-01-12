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
