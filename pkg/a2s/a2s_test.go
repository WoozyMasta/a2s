package a2s

import (
	"os"
	"strconv"
	"testing"
)

func TestSimple(t *testing.T) {
	host, ok1 := os.LookupEnv("SERVER_HOST")
	port, ok2 := os.LookupEnv("SERVER_PORT")
	if !ok1 && !ok2 {
		t.Error("Server address must be passed in 'SERVER_HOST' and 'SERVER_PORT' environment variables")
	}

	portInt, err := strconv.ParseInt(port, 10, 32)
	if err != nil {
		t.Errorf("Cant get port %v", err)
	}

	client, err := New(host, int(portInt))
	if err != nil {
		t.Error(err)
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
