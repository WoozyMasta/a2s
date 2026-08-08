package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const proxyCacheAuto = "auto"

// validateProxyCommand validates and normalizes parsed proxy options.
// It performs no DNS resolution, socket creation, or upstream request.
func validateProxyCommand(command *ProxyCommand) error {
	if command == nil {
		return fmt.Errorf("proxy command must not be nil")
	}

	upstream, err := normalizeEndpoint(command.Args.Host, command.Args.Port)
	if err != nil {
		return fmt.Errorf("invalid upstream endpoint: %w", err)
	}
	if err := validateProxyUDPEndpoint(upstream, "upstream"); err != nil {
		return err
	}

	command.Listen = strings.TrimSpace(command.Listen)
	if err := validateProxyUDPEndpoint(command.Listen, "listen"); err != nil {
		return err
	}

	cache, err := normalizeProxyCache(command.Cache)
	if err != nil {
		return err
	}

	command.Cache = cache
	return nil
}

// normalizeProxyCache validates cache selectors and removes duplicate values.
func normalizeProxyCache(values []string) ([]string, error) {
	if len(values) == 0 {
		return []string{proxyCacheAuto}, nil
	}

	seen := make(map[string]struct{}, len(values))
	cache := make([]string, 0, len(values))

	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			return nil, fmt.Errorf("cache selector must not be empty")
		}

		switch value {
		case proxyCacheAuto, "info", "players", "rules":
		default:
			return nil, fmt.Errorf("unsupported cache selector %q", value)
		}

		if _, ok := seen[value]; ok {
			continue
		}

		seen[value] = struct{}{}
		cache = append(cache, value)
	}

	_, hasAuto := seen[proxyCacheAuto]
	if len(cache) > 1 && hasAuto {
		return nil, fmt.Errorf(
			"cache selector %q cannot be combined with other values",
			proxyCacheAuto,
		)
	}

	return cache, nil
}

// validateProxyUDPEndpoint validates host:port syntax without resolving hosts.
func validateProxyUDPEndpoint(endpoint, name string) error {
	_, portText, err := net.SplitHostPort(endpoint)
	if err != nil {
		return fmt.Errorf("invalid %s endpoint %q: %w", name, endpoint, err)
	}

	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid %s port %q: want 1..65535", name, portText)
	}

	return nil
}
