package a2s

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readTestServers reads server addresses from test_servers.conf file
// Returns slice of "IP:PORT" strings, ignoring comments and empty lines
func readTestServers() ([]string, error) {
	// Try to find test_servers.conf relative to the test file location
	// First try current directory, then try pkg/a2s/ relative to project root
	var file *os.File
	var err error

	// Try current directory (when running tests from pkg/a2s/)
	file, err = os.Open("test_servers.conf")
	if err != nil {
		// Try relative to project root (when running tests from project root)
		file, err = os.Open(filepath.Join("pkg", "a2s", "test_servers.conf"))
		if err != nil {
			return nil, err
		}
	}
	defer file.Close()

	var servers []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		servers = append(servers, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return servers, nil
}

// getFirstTestServer returns the first server from test_servers.conf
func getFirstTestServer(t testing.TB) string {
	servers, err := readTestServers()
	if err != nil {
		t.Skipf("Cannot read test servers file: %v", err)
	}
	if len(servers) == 0 {
		t.Skip("No test servers found in test_servers.conf")
	}
	return servers[0]
}

// TestSimple tests all queries on first server from test_servers.conf
func TestSimple(t *testing.T) {
	serverAddr := getFirstTestServer(t)

	client, err := NewWithString(serverAddr)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	if _, err := client.GetInfo(); err != nil {
		t.Error(err)
	}

	if _, err := client.GetRules(); err != nil {
		t.Error(err)
	}

	if _, err := client.GetParsedRules(); err != nil {
		t.Error(err)
	}

	if _, err := client.GetPlayers(); err != nil {
		t.Error(err)
	}
}

// TestInfoSingle tests A2S_INFO query on first server from test_servers.conf
func TestInfoSingle(t *testing.T) {
	serverAddr := getFirstTestServer(t)

	client, err := NewWithString(serverAddr)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	info, err := client.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo failed: %v", err)
	}

	if info == nil {
		t.Fatal("GetInfo returned nil")
	}

	// Basic validation
	if info.Name == "" {
		t.Error("Server name is empty")
	}
	if info.Map == "" {
		t.Error("Server map is empty")
	}
	if info.Protocol == 0 {
		t.Error("Protocol version is 0")
	}

	t.Logf("Server: %s | Map: %s | Players: %d/%d | Ping: %v",
		info.Name, info.Map, info.Players, info.MaxPlayers, info.Ping)
}

// TestRulesSingle tests A2S_RULES query on first server from test_servers.conf
func TestRulesSingle(t *testing.T) {
	serverAddr := getFirstTestServer(t)

	client, err := NewWithString(serverAddr)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	rules, err := client.GetRules()
	if err != nil {
		t.Fatalf("GetRules failed: %v", err)
	}

	if rules == nil {
		t.Fatal("GetRules returned nil")
	}

	t.Logf("Retrieved %d rules", len(rules))
	for key, value := range rules {
		t.Logf("  %s = %s", key, value)
	}
}

// TestRulesParsedSingle tests A2S_RULES with parsing on first server
func TestRulesParsedSingle(t *testing.T) {
	serverAddr := getFirstTestServer(t)

	client, err := NewWithString(serverAddr)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	rules, err := client.GetParsedRules()
	if err != nil {
		t.Fatalf("GetParsedRules failed: %v", err)
	}

	if rules == nil {
		t.Fatal("GetParsedRules returned nil")
	}

	t.Logf("Retrieved %d parsed rules", len(rules))
}

// TestPlayersSingle tests A2S_PLAYER query on first server from test_servers.conf
func TestPlayersSingle(t *testing.T) {
	serverAddr := getFirstTestServer(t)

	client, err := NewWithString(serverAddr)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	players, err := client.GetPlayers()
	if err != nil {
		t.Fatalf("GetPlayers failed: %v", err)
	}

	if players == nil {
		t.Fatal("GetPlayers returned nil")
	}

	t.Logf("Retrieved %d players", len(*players))
	for i, player := range *players {
		t.Logf("  Player %d: %s (Score: %d, Duration: %v)",
			i+1, player.Name, player.Score, player.Duration)
	}
}

// TestInfoMultiple tests A2S_INFO query on all servers from test_servers.conf
func TestInfoMultiple(t *testing.T) {
	servers, err := readTestServers()
	if err != nil {
		t.Skipf("Cannot read test servers file: %v", err)
	}
	if len(servers) == 0 {
		t.Skip("No test servers found in test_servers.conf")
	}

	successCount := 0
	for _, serverAddr := range servers {
		client, err := NewWithString(serverAddr)
		if err != nil {
			t.Logf("Failed to create client for %s: %v", serverAddr, err)
			continue
		}

		info, err := client.GetInfo()
		client.Close()

		if err != nil {
			t.Logf("GetInfo failed for %s: %v", serverAddr, err)
			continue
		}

		if info != nil {
			successCount++
			t.Logf("✓ %s: %s | Map: %s | Players: %d/%d | Ping: %v",
				serverAddr, info.Name, info.Map, info.Players, info.MaxPlayers, info.Ping)
		}
	}

	t.Logf("Successfully queried %d/%d servers", successCount, len(servers))
}

// TestRulesMultiple tests A2S_RULES query on all servers from test_servers.conf
func TestRulesMultiple(t *testing.T) {
	servers, err := readTestServers()
	if err != nil {
		t.Skipf("Cannot read test servers file: %v", err)
	}
	if len(servers) == 0 {
		t.Skip("No test servers found in test_servers.conf")
	}

	successCount := 0
	for _, serverAddr := range servers {
		client, err := NewWithString(serverAddr)
		if err != nil {
			t.Logf("Failed to create client for %s: %v", serverAddr, err)
			continue
		}

		rules, err := client.GetRules()
		client.Close()

		if err != nil {
			t.Logf("GetRules failed for %s: %v", serverAddr, err)
			continue
		}

		if rules != nil {
			successCount++
			t.Logf("✓ %s: Retrieved %d rules", serverAddr, len(rules))
		}
	}

	t.Logf("Successfully queried %d/%d servers", successCount, len(servers))
}

// TestPlayersMultiple tests A2S_PLAYER query on all servers from test_servers.conf
func TestPlayersMultiple(t *testing.T) {
	servers, err := readTestServers()
	if err != nil {
		t.Skipf("Cannot read test servers file: %v", err)
	}
	if len(servers) == 0 {
		t.Skip("No test servers found in test_servers.conf")
	}

	successCount := 0
	for _, serverAddr := range servers {
		client, err := NewWithString(serverAddr)
		if err != nil {
			t.Logf("Failed to create client for %s: %v", serverAddr, err)
			continue
		}

		players, err := client.GetPlayers()
		client.Close()

		if err != nil {
			t.Logf("GetPlayers failed for %s: %v", serverAddr, err)
			continue
		}

		if players != nil {
			successCount++
			t.Logf("✓ %s: Retrieved %d players", serverAddr, len(*players))
		}
	}

	t.Logf("Successfully queried %d/%d servers", successCount, len(servers))
}

// BenchmarkInfo benchmarks A2S_INFO query
func BenchmarkInfo(b *testing.B) {
	serverAddr := getFirstTestServer(b)
	if serverAddr == "" {
		b.Skip("No test server available")
	}

	client, err := NewWithString(serverAddr)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetInfo()
		if err != nil {
			b.Fatalf("GetInfo failed: %v", err)
		}
	}
}

// BenchmarkRules benchmarks A2S_RULES query
func BenchmarkRules(b *testing.B) {
	serverAddr := getFirstTestServer(b)
	if serverAddr == "" {
		b.Skip("No test server available")
	}

	client, err := NewWithString(serverAddr)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetRules()
		if err != nil {
			b.Fatalf("GetRules failed: %v", err)
		}
	}
}

// BenchmarkRulesParsed benchmarks A2S_RULES with parsing
func BenchmarkRulesParsed(b *testing.B) {
	serverAddr := getFirstTestServer(b)
	if serverAddr == "" {
		b.Skip("No test server available")
	}

	client, err := NewWithString(serverAddr)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetParsedRules()
		if err != nil {
			b.Fatalf("GetParsedRules failed: %v", err)
		}
	}
}

// BenchmarkPlayers benchmarks A2S_PLAYER query
func BenchmarkPlayers(b *testing.B) {
	serverAddr := getFirstTestServer(b)
	if serverAddr == "" {
		b.Skip("No test server available")
	}

	client, err := NewWithString(serverAddr)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.GetPlayers()
		if err != nil {
			b.Fatalf("GetPlayers failed: %v", err)
		}
	}
}
