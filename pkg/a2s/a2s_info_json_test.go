// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"reflect"
	"testing"
)

func TestSourceInfoJSONContract(t *testing.T) {
	body := []byte{17}
	for _, value := range []string{"Test server", "test_map", "test_folder", "Test game"} {
		body = append(body, value...)
		body = append(body, 0)
	}
	body = binary.LittleEndian.AppendUint16(body, 1234)
	body = append(body, 1, 16, 2, 'd', 'w', 1, 0)
	body = append(body, "1.0"...)
	body = append(body, 0, byte(edfPort|edfSteamID|edfSourceTV|edfKeywords|edfGameID))
	body = binary.LittleEndian.AppendUint16(body, 27015)
	body = binary.LittleEndian.AppendUint64(body, 123456789)
	body = binary.LittleEndian.AppendUint16(body, 27020)
	body = append(body, "spectator"...)
	body = append(body, 0)
	body = append(body, "one,two"...)
	body = append(body, 0)
	body = binary.LittleEndian.AppendUint64(body, 107410)

	fixture := newUDPPacketFixture(t, singlePacketFixture(ResponseInfo, body))
	client, err := NewWithAddr(fixture.Addr())
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	info, err := client.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo returned error: %v", err)
	}
	if info.AppID != 1234 {
		t.Fatalf("parsed AppID = %d, want 1234", info.AppID)
	}
	if info.GameID == nil || *info.GameID != 107410 {
		t.Fatalf("parsed GameID = %v, want 107410", info.GameID)
	}
	if got := info.EffectiveID(); got != 107410 {
		t.Fatalf("effective ID = %d, want 107410", got)
	}

	jsonData, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal info: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(jsonData, &got); err != nil {
		t.Fatalf("unmarshal info JSON: %v", err)
	}
	want := map[string]any{
		"name":           "Test server",
		"map":            "test_map",
		"folder":         "test_folder",
		"game":           "Test game",
		"version":        "1.0",
		"keywords":       []any{"one", "two"},
		"app_id":         float64(1234),
		"game_id":        float64(107410),
		"steam_id":       float64(123456789),
		"port":           float64(27015),
		"source_tv_name": "spectator",
		"source_tv_port": float64(27020),
		"format":         "Source",
		"protocol":       float64(17),
		"players":        float64(1),
		"max_players":    float64(16),
		"bots":           float64(2),
		"server_type":    "Dedicated",
		"environment":    "Windows",
		"visibility":     true,
		"vac":            false,
		"EDF":            float64(241),
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Info JSON = %#v, want %#v", got, want)
	}
}
