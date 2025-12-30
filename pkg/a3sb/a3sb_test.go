package a3sb

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/steam/utils/appid"
)

// testServersConfig represents the structure of test_servers.json
type testServersConfig struct {
	Arma3 []string `json:"arma3"`
	DayZ  []string `json:"dayz"`
}

// readTestServersJSON reads server configuration from test_servers.json file
func readTestServersJSON() (*testServersConfig, error) {
	// Try to find file relative to the test file location
	// First try current directory, then try pkg/a3sb/ relative to project root
	var file *os.File
	var err error

	// Try current directory (when running tests from pkg/a3sb/)
	file, err = os.Open("test_servers.json")
	if err != nil {
		// Try relative to project root (when running tests from project root)
		file, err = os.Open(filepath.Join("pkg", "a3sb", "test_servers.json"))
		if err != nil {
			return nil, err
		}
	}
	defer file.Close()

	var config testServersConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// readTestServers reads all server addresses from test_servers.json
// Returns combined slice of all servers (Arma 3 + DayZ)
func readTestServers() ([]string, error) {
	config, err := readTestServersJSON()
	if err != nil {
		return nil, err
	}

	// Combine all servers
	var servers []string
	servers = append(servers, config.Arma3...)
	servers = append(servers, config.DayZ...)

	return servers, nil
}

// readTestServersArma3 reads Arma 3 server addresses from test_servers.json
func readTestServersArma3() ([]string, error) {
	config, err := readTestServersJSON()
	if err != nil {
		return nil, err
	}
	return config.Arma3, nil
}

// readTestServersDayZ reads DayZ server addresses from test_servers.json
func readTestServersDayZ() ([]string, error) {
	config, err := readTestServersJSON()
	if err != nil {
		return nil, err
	}
	return config.DayZ, nil
}

// getFirstTestServer returns the first server from test_servers.json (all servers combined)
func getFirstTestServer(t testing.TB) string {
	servers, err := readTestServers()
	if err != nil {
		t.Skipf("Cannot read test servers file: %v", err)
	}
	if len(servers) == 0 {
		t.Skip("No test servers found in test_servers.json")
	}
	return servers[0]
}

// createA3SBClient creates a3sb client from server address
func createA3SBClient(serverAddr string) (*Client, error) {
	a2sClient, err := a2s.NewWithString(serverAddr)
	if err != nil {
		return nil, err
	}
	return &Client{Client: a2sClient}, nil
}

// TestRulesSingle tests A2S_RULES query on first server from test_servers.json
func TestRulesSingle(t *testing.T) {
	serverAddr := getFirstTestServer(t)

	client, err := createA3SBClient(serverAddr)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Try to get rules with Arma 3 AppID first, then DayZ
	rules, err := client.GetRulesArma3()
	if err != nil {
		// If Arma 3 fails, try DayZ
		rules, err = client.GetRulesDayZ()
		if err != nil {
			t.Fatalf("GetRules failed for both Arma 3 and DayZ: %v", err)
		}
		t.Logf("Successfully retrieved DayZ rules")
	} else {
		t.Logf("Successfully retrieved Arma 3 rules")
	}

	if rules == nil {
		t.Fatal("GetRules returned nil")
	}

	// Basic validation
	if rules.Version == 0 {
		t.Error("Protocol version is 0")
	}

	t.Logf("Protocol version: %d | AppID: %d", rules.Version, rules.GetAppID())
	if rules.Description != "" {
		t.Logf("Description: %s", rules.Description)
	}
	if len(rules.Mods) > 0 {
		t.Logf("Mods count: %d", len(rules.Mods))
	}
	if len(rules.DLC) > 0 {
		t.Logf("DLC count: %d", len(rules.DLC))
	}
}

// TestRulesArma3Single tests A2S_RULES for Arma 3 on first server
func TestRulesArma3Single(t *testing.T) {
	serverAddr := getFirstTestServerArma3(t)

	client, err := createA3SBClient(serverAddr)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	rules, err := client.GetRulesArma3()
	if err != nil {
		t.Skipf("GetRulesArma3 failed (server might not be Arma 3): %v", err)
	}

	if rules == nil {
		t.Fatal("GetRulesArma3 returned nil")
	}

	// Basic validation
	if rules.Version == 0 {
		t.Error("Protocol version is 0")
	}

	t.Logf("Arma 3 Rules - Version: %d | AppID: %d", rules.Version, rules.GetAppID())
	if rules.Difficulty != nil {
		t.Logf("Difficulty: %+v", rules.Difficulty)
	}
	if len(rules.Mods) > 0 {
		t.Logf("Mods: %d", len(rules.Mods))
	}
	if len(rules.DLC) > 0 {
		t.Logf("DLC: %d", len(rules.DLC))
	}
	if len(rules.CreatorDLC) > 0 {
		t.Logf("Creator DLC: %d", len(rules.CreatorDLC))
	}
}

// TestRulesDayZSingle tests A2S_RULES for DayZ on first server
func TestRulesDayZSingle(t *testing.T) {
	serverAddr := getFirstTestServerDayZ(t)

	client, err := createA3SBClient(serverAddr)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	rules, err := client.GetRulesDayZ()
	if err != nil {
		t.Skipf("GetRulesDayZ failed (server might not be DayZ): %v", err)
	}

	if rules == nil {
		t.Fatal("GetRulesDayZ returned nil")
	}

	// Basic validation
	if rules.Version == 0 {
		t.Error("Protocol version is 0")
	}

	t.Logf("DayZ Rules - Version: %d | AppID: %d", rules.Version, rules.GetAppID())
	if rules.Description != "" {
		t.Logf("Description: %s", rules.Description)
	}
	if rules.Island != "" {
		t.Logf("Island: %s", rules.Island)
	}
	if rules.Platform != "" {
		t.Logf("Platform: %s", rules.Platform)
	}
	if len(rules.Mods) > 0 {
		t.Logf("Mods: %d", len(rules.Mods))
	}
	if len(rules.DLC) > 0 {
		t.Logf("DLC: %d", len(rules.DLC))
	}
}

// TestRulesMultiple tests A2S_RULES query on all servers from test_servers.json
func TestRulesMultiple(t *testing.T) {
	servers, err := readTestServers()
	if err != nil {
		t.Skipf("Cannot read test servers file: %v", err)
	}
	if len(servers) == 0 {
		t.Skip("No test servers found in test_servers.json")
	}

	successCount := 0
	arma3Count := 0
	dayzCount := 0

	for _, serverAddr := range servers {
		client, err := createA3SBClient(serverAddr)
		if err != nil {
			t.Logf("Failed to create client for %s: %v", serverAddr, err)
			continue
		}

		// Try Arma 3 first
		rules, err := client.GetRulesArma3()
		if err == nil && rules != nil {
			successCount++
			arma3Count++
			t.Logf("✓ %s (Arma 3): Version %d | Mods: %d | DLC: %d",
				serverAddr, rules.Version, len(rules.Mods), len(rules.DLC))
			client.Close()
			continue
		}

		// Try DayZ
		rules, err = client.GetRulesDayZ()
		client.Close()

		if err == nil && rules != nil {
			successCount++
			dayzCount++
			t.Logf("✓ %s (DayZ): Version %d | Island: %s | Mods: %d",
				serverAddr, rules.Version, rules.Island, len(rules.Mods))
		} else {
			t.Logf("✗ %s: Failed for both Arma 3 and DayZ", serverAddr)
		}
	}

	t.Logf("Successfully queried %d/%d servers (Arma 3: %d, DayZ: %d)",
		successCount, len(servers), arma3Count, dayzCount)
}

// TestRulesArma3Multiple tests A2S_RULES for Arma 3 on all servers
func TestRulesArma3Multiple(t *testing.T) {
	servers, err := readTestServersArma3()
	if err != nil {
		t.Skipf("Cannot read Arma 3 test servers file: %v", err)
	}
	if len(servers) == 0 {
		t.Skip("No Arma 3 test servers found in test_servers.json")
	}

	successCount := 0
	for _, serverAddr := range servers {
		client, err := createA3SBClient(serverAddr)
		if err != nil {
			t.Logf("Failed to create client for %s: %v", serverAddr, err)
			continue
		}

		rules, err := client.GetRulesArma3()
		client.Close()

		if err != nil {
			t.Logf("GetRulesArma3 failed for %s: %v", serverAddr, err)
			continue
		}

		if rules != nil {
			successCount++
			t.Logf("✓ %s: Version %d | Mods: %d | DLC: %d | Creator DLC: %d",
				serverAddr, rules.Version, len(rules.Mods), len(rules.DLC), len(rules.CreatorDLC))
		}
	}

	t.Logf("Successfully queried %d/%d Arma 3 servers", successCount, len(servers))
}

// TestRulesDayZMultiple tests A2S_RULES for DayZ on all servers
func TestRulesDayZMultiple(t *testing.T) {
	servers, err := readTestServersDayZ()
	if err != nil {
		t.Skipf("Cannot read DayZ test servers file: %v", err)
	}
	if len(servers) == 0 {
		t.Skip("No DayZ test servers found in test_servers.json")
	}

	successCount := 0
	for _, serverAddr := range servers {
		client, err := createA3SBClient(serverAddr)
		if err != nil {
			t.Logf("Failed to create client for %s: %v", serverAddr, err)
			continue
		}

		rules, err := client.GetRulesDayZ()
		client.Close()

		if err != nil {
			t.Logf("GetRulesDayZ failed for %s: %v", serverAddr, err)
			continue
		}

		if rules != nil {
			successCount++
			t.Logf("✓ %s: Version %d | Island: %s | Platform: %s | Mods: %d | DLC: %d",
				serverAddr, rules.Version, rules.Island, rules.Platform, len(rules.Mods), len(rules.DLC))
		}
	}

	t.Logf("Successfully queried %d/%d DayZ servers", successCount, len(servers))
}

// BenchmarkRules benchmarks A2S_RULES query (auto-detect game)
func BenchmarkRules(b *testing.B) {
	serverAddr := getFirstTestServer(b)
	if serverAddr == "" {
		b.Skip("No test server available")
	}

	client, err := createA3SBClient(serverAddr)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Try Arma 3 first, then DayZ
	var gameID uint64
	_, err = client.GetRulesArma3()
	if err != nil {
		// If Arma 3 fails, use DayZ
		gameID = 221100 // DayZ AppID
	} else {
		gameID = 107410 // Arma 3 AppID
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetRules(gameID)
		if err != nil {
			b.Fatalf("GetRules failed: %v", err)
		}
	}
}

// getFirstTestServerArma3 returns the first Arma 3 server from test_servers.json
func getFirstTestServerArma3(t testing.TB) string {
	servers, err := readTestServersArma3()
	if err != nil {
		t.Skipf("Cannot read Arma 3 test servers file: %v", err)
	}
	if len(servers) == 0 {
		t.Skip("No Arma 3 test servers found in test_servers.json")
	}
	return servers[0]
}

// getFirstTestServerDayZ returns the first DayZ server from test_servers.json
func getFirstTestServerDayZ(t testing.TB) string {
	servers, err := readTestServersDayZ()
	if err != nil {
		t.Skipf("Cannot read DayZ test servers file: %v", err)
	}
	if len(servers) == 0 {
		t.Skip("No DayZ test servers found in test_servers.json")
	}
	return servers[0]
}

// BenchmarkRulesArma3 benchmarks A2S_RULES for Arma 3
func BenchmarkRulesArma3(b *testing.B) {
	serverAddr := getFirstTestServerArma3(b)
	if serverAddr == "" {
		b.Skip("No Arma 3 test server available")
	}

	client, err := createA3SBClient(serverAddr)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetRules(appid.Arma3.Uint64())
		if err != nil {
			b.Fatalf("GetRules failed for Arma 3: %v", err)
		}
	}
}

// BenchmarkRulesDayZ benchmarks A2S_RULES for DayZ
func BenchmarkRulesDayZ(b *testing.B) {
	serverAddr := getFirstTestServerDayZ(b)
	if serverAddr == "" {
		b.Skip("No DayZ test server available")
	}

	client, err := createA3SBClient(serverAddr)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetRules(appid.DayZ.Uint64())
		if err != nil {
			b.Fatalf("GetRules failed for DayZ: %v", err)
		}
	}
}
