//go:build integration

package a3sb

import (
	"context"
	"testing"

	"github.com/woozymasta/a2s/pkg/appid"
)

// BenchmarkIntegrationRules measures automatic A3SB parsing against a live server.
func BenchmarkIntegrationRules(b *testing.B) {
	serverAddr := getFirstTestServer(b)
	client, err := createA3SBClient(serverAddr)
	if err != nil {
		b.Fatalf("create client: %v", err)
	}
	defer client.Close()

	var game uint64 = appid.Arma3
	if _, err := client.GetRulesArma3(context.Background()); err != nil {
		game = appid.DayZ
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := client.GetRules(context.Background(), game); err != nil {
			b.Fatalf("GetRules failed: %v", err)
		}
	}
}

// BenchmarkIntegrationRulesArma3 measures explicit Arma 3 parsing against a live server.
func BenchmarkIntegrationRulesArma3(b *testing.B) {
	serverAddr := getFirstTestServerArma3(b)
	client, err := createA3SBClient(serverAddr)
	if err != nil {
		b.Fatalf("create client: %v", err)
	}
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := client.GetRules(context.Background(), appid.Arma3); err != nil {
			b.Fatalf("GetRules failed for Arma 3: %v", err)
		}
	}
}

// BenchmarkIntegrationRulesDayZ measures explicit DayZ parsing against a live server.
func BenchmarkIntegrationRulesDayZ(b *testing.B) {
	serverAddr := getFirstTestServerDayZ(b)
	client, err := createA3SBClient(serverAddr)
	if err != nil {
		b.Fatalf("create client: %v", err)
	}
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := client.GetRules(context.Background(), appid.DayZ); err != nil {
			b.Fatalf("GetRules failed for DayZ: %v", err)
		}
	}
}
