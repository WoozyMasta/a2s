package server

import (
	"context"
	"errors"
	"net/netip"
	"testing"

	"github.com/woozymasta/a2s/pkg/a2s"
)

type gateTestProvider struct {
	challenge a2s.Challenge
	valid     bool
	issued    int
}

func (p *gateTestProvider) Issue(netip.AddrPort) a2s.Challenge {
	p.issued++
	return p.challenge
}

func (p *gateTestProvider) Validate(netip.AddrPort, a2s.Challenge) bool {
	return p.valid
}

func TestChallengeGateRequiresInfoChallenge(t *testing.T) {
	provider := &gateTestProvider{challenge: a2s.Challenge{1, 2, 3, 4}}
	called := false
	next := HandlerFunc(func(context.Context, *Request) (Response, error) {
		called = true
		return InfoResponse{}, nil
	})
	gate, err := NewChallengeGate(next, SecureChallengePolicy(), provider)
	if err != nil {
		t.Fatalf("NewChallengeGate() error = %v", err)
	}

	response, err := gate.Handle(context.Background(), &Request{
		Remote: netip.MustParseAddrPort("192.0.2.1:27015"),
		Query:  a2s.Request{Type: a2s.InfoRequest},
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	packet, ok := response.(PacketResponse)
	if !ok {
		t.Fatalf("Handle() response = %T, want PacketResponse", response)
	}
	if packet.Packet.Type != a2s.ResponseChallenge {
		t.Fatalf("challenge response type = 0x%X, want 0x%X", packet.Packet.Type, a2s.ResponseChallenge)
	}
	if len(packet.Packet.Payload) != len(a2s.Challenge{}) {
		t.Fatalf("challenge payload length = %d, want 4", len(packet.Packet.Payload))
	}
	if called {
		t.Fatal("handler was called for a request without a challenge")
	}
	if provider.issued != 1 {
		t.Fatalf("Issue() calls = %d, want 1", provider.issued)
	}
}

func TestChallengeGateForwardsValidChallenge(t *testing.T) {
	provider := &gateTestProvider{valid: true}
	called := false
	want := RulesResponse{Rules: a2s.Rules{{Name: "hostname", Value: "server"}}}
	next := HandlerFunc(func(context.Context, *Request) (Response, error) {
		called = true
		return want, nil
	})
	gate, err := NewChallengeGate(next, SecureChallengePolicy(), provider)
	if err != nil {
		t.Fatalf("NewChallengeGate() error = %v", err)
	}

	response, err := gate.Handle(context.Background(), &Request{
		Remote: netip.MustParseAddrPort("192.0.2.1:27015"),
		Query: a2s.Request{
			Type:         a2s.RulesRequest,
			Challenge:    a2s.Challenge{4, 3, 2, 1},
			HasChallenge: true,
		},
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	got, ok := response.(RulesResponse)
	if !ok || len(got.Rules) != len(want.Rules) || got.Rules[0] != want.Rules[0] {
		t.Fatalf("Handle() response = %#v, want %#v", response, want)
	}
	if !called {
		t.Fatal("handler was not called for a valid challenge")
	}
	if provider.issued != 0 {
		t.Fatalf("Issue() calls = %d, want 0", provider.issued)
	}
}

func TestChallengeGateRejectsInvalidChallengeWithoutHandler(t *testing.T) {
	provider := &gateTestProvider{challenge: a2s.Challenge{9, 8, 7, 6}}
	called := false
	next := HandlerFunc(func(context.Context, *Request) (Response, error) {
		called = true
		return nil, errors.New("handler must not be called")
	})
	gate, err := NewChallengeGate(next, SecureChallengePolicy(), provider)
	if err != nil {
		t.Fatalf("NewChallengeGate() error = %v", err)
	}

	response, err := gate.Handle(context.Background(), &Request{
		Remote: netip.MustParseAddrPort("192.0.2.1:27015"),
		Query: a2s.Request{
			Type:         a2s.PlayerRequest,
			Challenge:    a2s.Challenge{1, 1, 1, 1},
			HasChallenge: true,
		},
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if _, ok := response.(PacketResponse); !ok {
		t.Fatalf("Handle() response = %T, want PacketResponse", response)
	}
	if called {
		t.Fatal("handler was called for an invalid challenge")
	}
}

func TestChallengeGateLegacyInfoBypassesChallenge(t *testing.T) {
	provider := &gateTestProvider{}
	called := false
	next := HandlerFunc(func(context.Context, *Request) (Response, error) {
		called = true
		return InfoResponse{}, nil
	})
	gate, err := NewChallengeGate(next, LegacyChallengePolicy(), provider)
	if err != nil {
		t.Fatalf("NewChallengeGate() error = %v", err)
	}

	if _, err := gate.Handle(context.Background(), &Request{
		Query: a2s.Request{Type: a2s.InfoRequest},
	}); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if !called {
		t.Fatal("handler was not called for legacy direct INFO")
	}
	if provider.issued != 0 {
		t.Fatalf("Issue() calls = %d, want 0", provider.issued)
	}
}

func TestNewChallengeGateRejectsMissingConfiguration(t *testing.T) {
	handler := HandlerFunc(func(context.Context, *Request) (Response, error) {
		return nil, nil
	})
	provider := &gateTestProvider{}
	policy := NoChallengePolicy()

	for name, args := range map[string]struct {
		next     Handler
		policy   ChallengePolicy
		provider ChallengeProvider
	}{
		"handler":  {next: nil, policy: policy, provider: provider},
		"policy":   {next: handler, policy: nil, provider: provider},
		"provider": {next: handler, policy: policy, provider: nil},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewChallengeGate(args.next, args.policy, args.provider); !errors.Is(err, ErrChallengeGate) {
				t.Fatalf("NewChallengeGate() error = %v, want ErrChallengeGate", err)
			}
		})
	}
}
