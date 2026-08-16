// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"context"
	"testing"
)

func TestQueryReturnsLogicalPacket(t *testing.T) {
	fixture := newUDPPacketFixture(t, singlePacketFixture(ResponseRules, []byte("rules")))
	client, err := NewWithAddr(fixture.Addr())
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	packet, meta, err := client.Query(context.Background(), RulesRequest)
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	if packet.Type != ResponseRules {
		t.Fatalf("packet type = 0x%X, want 0x%X", packet.Type, ResponseRules)
	}
	if string(packet.Payload) != "rules" {
		t.Fatalf("packet payload = %q, want %q", packet.Payload, "rules")
	}
	if meta.Duration < 0 {
		t.Fatalf("query duration = %s, want non-negative duration", meta.Duration)
	}
	if meta.UsedChallenge {
		t.Fatal("query metadata reports an unexpected challenge exchange")
	}
}

func TestQueryMetadataReportsChallengeExchange(t *testing.T) {
	fixture := newUDPPacketFixture(
		t,
		singlePacketFixture(ResponseChallenge, []byte{1, 2, 3, 4}),
		singlePacketFixture(ResponseRules, []byte("rules")),
	)
	client, err := NewWithAddr(fixture.Addr())
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	_, meta, err := client.Query(context.Background(), RulesRequest)
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	if !meta.UsedChallenge {
		t.Fatal("query metadata does not report the challenge exchange")
	}
}
