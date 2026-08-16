// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package testfixtures

import "testing"

func TestReadProtocolFixtures(t *testing.T) {
	fixtures := []string{
		"request_info.hex",
		"request_challenge.hex",
		"request_players_challenge.hex",
		"request_rules_challenge.hex",
		"response_challenge.hex",
		"source_info_payload.hex",
		"goldsource_info_payload.hex",
		"players_payload.hex",
		"rules_payload.hex",
		"a3sb_arma3_payload.hex",
		"a3sb_dayz_payload.hex",
		"source_split_packet_0.hex",
		"source_split_packet_1.hex",
		"source_split_packet_2.hex",
		"source_split_packet_3.hex",
		"source_split_packet_4.hex",
	}

	for _, name := range fixtures {
		t.Run(name, func(t *testing.T) {
			data, err := Read(name)
			if err != nil {
				t.Fatal(err)
			}
			if len(data) == 0 {
				t.Fatal("fixture is empty")
			}
		})
	}
}
