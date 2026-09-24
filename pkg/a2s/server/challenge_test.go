// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package server

import (
	"bytes"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
)

var _ ChallengeProvider = (*statelessChallengeProvider)(nil)

func TestStatelessChallengeProviderValidatesCurrentAndPreviousBucket(t *testing.T) {
	secret := bytes.Repeat([]byte{0x42}, challengeSecretSize)
	now := time.Unix(120, 0)
	provider, err := newStatelessChallengeProvider(bytes.NewReader(secret), func() time.Time {
		return now
	})
	if err != nil {
		t.Fatalf("newStatelessChallengeProvider() error = %v", err)
	}

	remote := netip.MustParseAddrPort("192.0.2.10:27015")
	challenge := provider.Issue(remote)
	if challenge == a2s.InitialChallenge {
		t.Fatal("Issue() returned reserved initial challenge")
	}
	if !provider.Validate(remote, challenge) {
		t.Fatal("Validate() rejected current-bucket challenge")
	}

	now = time.Unix(180, 0)
	if !provider.Validate(remote, challenge) {
		t.Fatal("Validate() rejected previous-bucket challenge")
	}

	now = time.Unix(240, 0)
	if provider.Validate(remote, challenge) {
		t.Fatal("Validate() accepted expired challenge")
	}
}

func TestStatelessChallengeProviderBindsRemoteEndpoint(t *testing.T) {
	secret := bytes.Repeat([]byte{0x24}, challengeSecretSize)
	provider, err := newStatelessChallengeProvider(bytes.NewReader(secret), func() time.Time {
		return time.Unix(600, 0)
	})
	if err != nil {
		t.Fatalf("newStatelessChallengeProvider() error = %v", err)
	}

	remote := netip.MustParseAddrPort("198.51.100.7:27015")
	challenge := provider.Issue(remote)
	for _, other := range []string{
		"198.51.100.8:27015",
		"198.51.100.7:27016",
		"[2001:db8::7]:27015",
	} {
		if provider.Validate(netip.MustParseAddrPort(other), challenge) {
			t.Errorf("Validate() accepted challenge for unrelated endpoint %s", other)
		}
	}
}

func TestStatelessChallengeProviderCanonicalizesMappedIPv4(t *testing.T) {
	secret := bytes.Repeat([]byte{0x11}, challengeSecretSize)
	provider, err := newStatelessChallengeProvider(bytes.NewReader(secret), func() time.Time {
		return time.Unix(600, 0)
	})
	if err != nil {
		t.Fatalf("newStatelessChallengeProvider() error = %v", err)
	}

	ipv4 := netip.MustParseAddrPort("192.0.2.7:27015")
	mapped := netip.MustParseAddrPort("[::ffff:192.0.2.7]:27015")
	challenge := provider.Issue(ipv4)
	if !provider.Validate(mapped, challenge) {
		t.Fatal("Validate() did not canonicalize IPv4-mapped IPv6 endpoint")
	}
}

func TestStatelessChallengeProviderRejectsSecretFailure(t *testing.T) {
	_, err := newStatelessChallengeProvider(bytes.NewReader(nil), time.Now)
	if !errors.Is(err, ErrChallengeProvider) {
		t.Fatalf("newStatelessChallengeProvider() error = %v, want ErrChallengeProvider", err)
	}
}

func TestStatelessChallengeProviderNeverIssuesInitialChallenge(t *testing.T) {
	secret := bytes.Repeat([]byte{0xA5}, challengeSecretSize)
	provider, err := newStatelessChallengeProvider(bytes.NewReader(secret), func() time.Time {
		return time.Unix(0, 0)
	})
	if err != nil {
		t.Fatalf("newStatelessChallengeProvider() error = %v", err)
	}

	for port := 1; port <= 1000; port++ {
		remote := netip.AddrPortFrom(netip.MustParseAddr("192.0.2.1"), uint16(port))
		if got := provider.Issue(remote); got == a2s.InitialChallenge {
			t.Fatalf("Issue(%s) returned reserved initial challenge", remote)
		}
	}
}
