// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package server

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"

	"github.com/woozymasta/a2s/pkg/a2s"
)

// Snapshot contains the typed responses served by State.
//
// Info is optional. Players and Rules are always stored as responses;
// nil or empty slices produce valid empty A2S_PLAYER and A2S_RULES responses.
type Snapshot struct {
	// Info is the A2S_INFO response, or nil when INFO is unavailable.
	Info *a2s.Info
	// Players is the current A2S_PLAYER response.
	Players []a2s.Player
	// Rules is the current A2S_RULES response in wire order.
	Rules a2s.Rules
}

// State is a concurrency-safe handler backed by an immutable encoded snapshot.
//
// The zero value is ready to use. Store publishes a complete replacement;
// readers never observe a partially updated snapshot and do not take a lock.
type State struct {
	// State must not be copied after its first use.
	snapshot atomic.Pointer[stateSnapshot]
}

type stateSnapshot struct {
	info    a2s.Packet // Pre-encoded INFO response.
	players a2s.Packet // Pre-encoded PLAYER response.
	rules   a2s.Packet // Pre-encoded RULES response.
	hasInfo bool       // Whether an INFO response was stored.
}

// Ensure State remains a Handler implementation.
var _ Handler = (*State)(nil)

// NewState returns an empty state handler.
func NewState() *State {
	return &State{}
}

// Store validates, deep-copies, encodes, and atomically publishes snapshot.
// If encoding fails, the previously published snapshot remains active.
func (s *State) Store(snapshot Snapshot) error {
	if s == nil {
		return fmt.Errorf("%w: state is nil", ErrState)
	}

	copySnapshot := cloneSnapshot(snapshot)
	next := &stateSnapshot{}
	if copySnapshot.Info != nil {
		packet, err := normalizeInfoResponse(a2s.InfoRequest, *copySnapshot.Info)
		if err != nil {
			return fmt.Errorf("%w: info: %w", ErrState, err)
		}

		next.info = packet
		next.hasInfo = true
	}

	players, err := normalizePlayersResponse(a2s.PlayerRequest, copySnapshot.Players)
	if err != nil {
		return fmt.Errorf("%w: players: %w", ErrState, err)
	}
	next.players = players

	rules, err := normalizeRulesResponse(a2s.RulesRequest, copySnapshot.Rules)
	if err != nil {
		return fmt.Errorf("%w: rules: %w", ErrState, err)
	}
	next.rules = rules

	s.snapshot.Store(next)
	return nil
}

// Handle returns a copy of the pre-encoded response for the request type.
func (s *State) Handle(_ context.Context, request *Request) (Response, error) {
	if s == nil {
		return nil, fmt.Errorf("%w: state is nil", ErrState)
	}
	if request == nil {
		return nil, fmt.Errorf("%w: request is nil", ErrState)
	}

	snapshot := s.snapshot.Load()
	if snapshot == nil {
		return nil, fmt.Errorf("%w: state has not been initialized", ErrStateUnavailable)
	}

	switch request.Query.Type {
	case a2s.InfoRequest:
		if !snapshot.hasInfo {
			return nil, fmt.Errorf("%w: info", ErrStateUnavailable)
		}
		return PacketResponse{Packet: clonePacket(snapshot.info)}, nil

	case a2s.PlayerRequest:
		return PacketResponse{Packet: clonePacket(snapshot.players)}, nil

	case a2s.RulesRequest:
		return PacketResponse{Packet: clonePacket(snapshot.rules)}, nil

	default:
		return nil, fmt.Errorf("%w: unsupported query 0x%X", ErrState, request.Query.Type)
	}
}

// cloneSnapshot copies all mutable model data before Store encodes it.
func cloneSnapshot(snapshot Snapshot) Snapshot {
	copySnapshot := Snapshot{
		Info:    cloneInfo(snapshot.Info),
		Players: clonePlayers(snapshot.Players),
		Rules:   cloneRules(snapshot.Rules),
	}

	return copySnapshot
}

// cloneInfo copies the pointer and slice fields used by a2s.Info.
func cloneInfo(info *a2s.Info) *a2s.Info {
	if info == nil {
		return nil
	}

	copyInfo := *info
	if info.TheShip != nil {
		theShip := *info.TheShip
		copyInfo.TheShip = &theShip
	}
	if info.Mod != nil {
		mod := *info.Mod
		copyInfo.Mod = &mod
	}
	if info.GameID != nil {
		gameID := *info.GameID
		copyInfo.GameID = &gameID
	}
	if info.Keywords != nil {
		copyInfo.Keywords = make([]string, len(info.Keywords))
		copy(copyInfo.Keywords, info.Keywords)
	}

	return &copyInfo
}

// clonePlayers copies the player slice before encoding it.
func clonePlayers(players []a2s.Player) []a2s.Player {
	if players == nil {
		return nil
	}

	copyPlayers := make([]a2s.Player, len(players))
	copy(copyPlayers, players)
	return copyPlayers
}

// cloneRules copies the ordered rule entries before encoding them.
func cloneRules(rules a2s.Rules) a2s.Rules {
	if rules == nil {
		return nil
	}

	copyRules := make(a2s.Rules, len(rules))
	copy(copyRules, rules)
	return copyRules
}

// clonePacket protects the immutable snapshot from PacketResponse callers.
func clonePacket(packet a2s.Packet) a2s.Packet {
	packet.Payload = bytes.Clone(packet.Payload)
	return packet
}
