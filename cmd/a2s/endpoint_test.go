package main

import "testing"

func TestNormalizeEndpoint(t *testing.T) {
	tests := []struct {
		name string
		host string
		port string
		want string
	}{
		{
			name: "hostname with port",
			host: "example.org",
			port: "2303",
			want: "example.org:2303",
		},
		{
			name: "ipv4 with port",
			host: "127.0.0.1",
			port: "27015",
			want: "127.0.0.1:27015",
		},
		{
			name: "bracketed ipv6 endpoint",
			host: "[2001:db8::1]:27015",
			want: "[2001:db8::1]:27015",
		},
		{
			name: "bracketed ipv6 with separate port",
			host: "[2001:db8::1]",
			port: "27015",
			want: "[2001:db8::1]:27015",
		},
		{
			name: "bare ipv6 with separate port",
			host: "2001:db8::1",
			port: "27015",
			want: "[2001:db8::1]:27015",
		},
		{
			name: "service port",
			host: "example.org",
			port: "steam",
			want: "example.org:steam",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeEndpoint(tt.host, tt.port)
			if err != nil {
				t.Fatalf("normalizeEndpoint() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("normalizeEndpoint() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeEndpointRejectsInvalidCombinations(t *testing.T) {
	tests := []struct {
		name string
		host string
		port string
	}{
		{
			name: "empty host",
			host: "",
			port: "27015",
		},
		{
			name: "missing port",
			host: "example.org",
		},
		{
			name: "port duplicated",
			host: "example.org:2303",
			port: "27015",
		},
		{
			name: "empty embedded port",
			host: "example.org:",
		},
		{
			name: "empty bracketed port",
			host: "[2001:db8::1]:",
		},
		{
			name: "unmatched brackets",
			host: "[2001:db8::1",
			port: "27015",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := normalizeEndpoint(tt.host, tt.port); err == nil {
				t.Fatal("normalizeEndpoint() error = nil, want validation error")
			}
		})
	}
}
