// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

//go:build integration

package a2s

import (
	"context"
	"testing"
)

// BenchmarkIntegrationInfo measures a complete A2S_INFO query against a live server.
func BenchmarkIntegrationInfo(b *testing.B) {
	serverAddr := getFirstTestServer(b)
	client, err := NewWithString(serverAddr)
	if err != nil {
		b.Fatalf("create client: %v", err)
	}
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := client.GetInfo(context.Background()); err != nil {
			b.Fatalf("GetInfo failed: %v", err)
		}
	}
}

// BenchmarkIntegrationRules measures a complete A2S_RULES query against a live server.
func BenchmarkIntegrationRules(b *testing.B) {
	serverAddr := getFirstTestServer(b)
	client, err := NewWithString(serverAddr)
	if err != nil {
		b.Fatalf("create client: %v", err)
	}
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := client.GetRules(context.Background()); err != nil {
			b.Fatalf("GetRules failed: %v", err)
		}
	}
}

// BenchmarkIntegrationParsedRules measures parsed A2S_RULES values against a live server.
func BenchmarkIntegrationParsedRules(b *testing.B) {
	serverAddr := getFirstTestServer(b)
	client, err := NewWithString(serverAddr)
	if err != nil {
		b.Fatalf("create client: %v", err)
	}
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := client.GetParsedRules(context.Background()); err != nil {
			b.Fatalf("GetParsedRules failed: %v", err)
		}
	}
}

// BenchmarkIntegrationPlayers measures a complete A2S_PLAYER query against a live server.
func BenchmarkIntegrationPlayers(b *testing.B) {
	serverAddr := getFirstTestServer(b)
	client, err := NewWithString(serverAddr)
	if err != nil {
		b.Fatalf("create client: %v", err)
	}
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := client.GetPlayers(context.Background()); err != nil {
			b.Fatalf("GetPlayers failed: %v", err)
		}
	}
}
