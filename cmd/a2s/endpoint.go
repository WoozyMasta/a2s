package main

import (
	"fmt"
	"net"
	"strings"
)

// normalizeEndpoint combines a host and optional query port into a UDP endpoint.
// It accepts hostnames, IPv4, bracketed IPv6 and bare IPv6 literals.
func normalizeEndpoint(host, port string) (string, error) {
	host = strings.TrimSpace(host)
	port = strings.TrimSpace(port)
	if host == "" {
		return "", fmt.Errorf("host must not be empty")
	}

	if strings.HasPrefix(host, "[") {
		if endpointHost, endpointPort, err := net.SplitHostPort(host); err == nil {
			if endpointPort == "" {
				return "", fmt.Errorf("host %q has an empty port", host)
			}
			if port != "" {
				return "", fmt.Errorf("host %q already includes port %q", host, endpointPort)
			}
			return net.JoinHostPort(endpointHost, endpointPort), nil
		}

		if !strings.HasSuffix(host, "]") {
			return "", fmt.Errorf("invalid bracketed host %q", host)
		}

		host = strings.TrimSuffix(strings.TrimPrefix(host, "["), "]")
		if host == "" {
			return "", fmt.Errorf("invalid bracketed host")
		}
	} else if strings.HasSuffix(host, "]") {
		return "", fmt.Errorf("invalid bracketed host %q", host)
	}

	if endpointHost, endpointPort, err := net.SplitHostPort(host); err == nil {
		if endpointPort == "" {
			return "", fmt.Errorf("host %q has an empty port", host)
		}
		if port != "" {
			return "", fmt.Errorf("host %q already includes port %q", host, endpointPort)
		}
		return net.JoinHostPort(endpointHost, endpointPort), nil
	}

	if port == "" {
		return "", fmt.Errorf("query port must be provided when host has no port")
	}

	return net.JoinHostPort(host, port), nil
}
