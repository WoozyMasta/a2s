// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestAppendRulesMatchesFixture(t *testing.T) {
	payload := readProtocolFixture(t, "rules_payload.hex")
	rules, err := DecodeRules(Packet{Type: ResponseRules, Payload: payload})
	if err != nil {
		t.Fatalf("DecodeRules() error = %v", err)
	}

	encoded, err := AppendRules(nil, rules)
	if err != nil {
		t.Fatalf("AppendRules() error = %v", err)
	}

	want := singlePacketFixture(ResponseRules, payload)
	if !bytes.Equal(encoded, want) {
		t.Fatalf("encoded rules response = %X, want %X", encoded, want)
	}
}

func TestAppendRulesPreservesOrderAndDuplicates(t *testing.T) {
	want := Rules{
		{Name: "mode", Value: "coop"},
		{Name: "mode", Value: "versus"},
		{Name: "empty", Value: ""},
	}

	encoded, err := AppendRules(nil, want)
	if err != nil {
		t.Fatalf("AppendRules() error = %v", err)
	}
	packet, err := DecodePacket(encoded)
	if err != nil {
		t.Fatalf("DecodePacket() error = %v", err)
	}
	got, err := DecodeRules(packet)
	if err != nil {
		t.Fatalf("DecodeRules() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip rules = %+v, want %+v", got, want)
	}

	repeated, err := AppendRules(nil, want)
	if err != nil {
		t.Fatalf("second AppendRules() error = %v", err)
	}
	if !bytes.Equal(repeated, encoded) {
		t.Fatal("AppendRules() output is not deterministic")
	}
}

func TestAppendRulesRejectsInvalidModels(t *testing.T) {
	tests := []struct {
		name  string
		rules Rules
	}{
		{name: "too many rules", rules: make(Rules, 65536)},
		{name: "NUL in name", rules: Rules{{Name: "bad\x00name"}}},
		{name: "NUL in value", rules: Rules{{Value: "bad\x00value"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := AppendRules(nil, test.rules); !errors.Is(err, ErrRuleEncode) {
				t.Fatalf("AppendRules() error = %v, want ErrRuleEncode", err)
			}
		})
	}
}

func TestDecodeRulesRejectsWrongResponseType(t *testing.T) {
	_, err := DecodeRules(Packet{Type: ResponsePlayers})
	if !errors.Is(err, ErrRuleRead) {
		t.Fatalf("DecodeRules() error = %v, want ErrRuleRead", err)
	}
}

func TestDecodeRulesIgnoresTrailingData(t *testing.T) {
	payload := []byte{
		1, 0,
		'h', 'o', 's', 't', 0,
		't', 'e', 's', 't', 0,
		0xFF, 0x00,
	}

	rules, err := DecodeRules(Packet{Type: ResponseRules, Payload: payload})
	if err != nil {
		t.Fatalf("DecodeRules() error = %v", err)
	}
	if len(rules) != 1 || rules[0].Name != "host" || rules[0].Value != "test" {
		t.Fatalf("rules = %#v, want one host=test rule", rules)
	}
}
