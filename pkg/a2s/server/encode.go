// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package server

import (
	"fmt"

	"github.com/woozymasta/a2s/pkg/a2s"
)

// NormalizeResponse converts a handler response into one logical A2S packet.
//
// Typed responses go through the shared package encoders.
// PacketResponse is already a logical packet
// and is returned without decoding or re-encoding its payload.
func NormalizeResponse(req *Request, response Response) (a2s.Packet, error) {
	if req == nil {
		return a2s.Packet{}, fmt.Errorf("%w: nil request", ErrResponse)
	}
	if response == nil {
		return a2s.Packet{}, fmt.Errorf("%w: nil response", ErrResponse)
	}

	switch response := response.(type) {
	case InfoResponse:
		return normalizeInfoResponse(req.Query.Type, response.Info)

	case *InfoResponse:
		if response == nil {
			return a2s.Packet{}, fmt.Errorf("%w: nil InfoResponse", ErrResponse)
		}
		return normalizeInfoResponse(req.Query.Type, response.Info)

	case PlayersResponse:
		return normalizePlayersResponse(req.Query.Type, response.Players)

	case *PlayersResponse:
		if response == nil {
			return a2s.Packet{}, fmt.Errorf("%w: nil PlayersResponse", ErrResponse)
		}
		return normalizePlayersResponse(req.Query.Type, response.Players)

	case RulesResponse:
		return normalizeRulesResponse(req.Query.Type, response.Rules)

	case *RulesResponse:
		if response == nil {
			return a2s.Packet{}, fmt.Errorf("%w: nil RulesResponse", ErrResponse)
		}
		return normalizeRulesResponse(req.Query.Type, response.Rules)

	case PacketResponse:
		return normalizePacketResponse(req.Query.Type, response.Packet)

	case *PacketResponse:
		if response == nil {
			return a2s.Packet{}, fmt.Errorf("%w: nil PacketResponse", ErrResponse)
		}
		return normalizePacketResponse(req.Query.Type, response.Packet)

	default:
		return a2s.Packet{}, fmt.Errorf("%w: unsupported response %T", ErrResponse, response)
	}
}

// normalizeInfoResponse encodes an InfoResponse after checking its query kind.
func normalizeInfoResponse(query a2s.QueryType, info a2s.Info) (a2s.Packet, error) {
	if err := validateResponseQuery(query, a2s.ResponseInfo); err != nil {
		return a2s.Packet{}, err
	}

	data, err := a2s.AppendInfo(nil, info)
	if err != nil {
		return a2s.Packet{}, fmt.Errorf("%w: encode InfoResponse: %w", ErrResponse, err)
	}

	return decodeEncodedResponse(data)
}

// normalizePlayersResponse encodes a PlayersResponse after checking its query kind.
func normalizePlayersResponse(query a2s.QueryType, players []a2s.Player) (a2s.Packet, error) {
	if err := validateResponseQuery(query, a2s.ResponsePlayers); err != nil {
		return a2s.Packet{}, err
	}

	data, err := a2s.AppendPlayers(nil, players)
	if err != nil {
		return a2s.Packet{}, fmt.Errorf("%w: encode PlayersResponse: %w", ErrResponse, err)
	}

	return decodeEncodedResponse(data)
}

// normalizeRulesResponse encodes a RulesResponse after checking its query kind.
func normalizeRulesResponse(query a2s.QueryType, rules a2s.Rules) (a2s.Packet, error) {
	if err := validateResponseQuery(query, a2s.ResponseRules); err != nil {
		return a2s.Packet{}, err
	}

	data, err := a2s.AppendRules(nil, rules)
	if err != nil {
		return a2s.Packet{}, fmt.Errorf("%w: encode RulesResponse: %w", ErrResponse, err)
	}

	return decodeEncodedResponse(data)
}

// normalizePacketResponse validates a logical packet without changing it.
func normalizePacketResponse(query a2s.QueryType, packet a2s.Packet) (a2s.Packet, error) {
	if err := validateResponseQuery(query, packet.Type); err != nil {
		return a2s.Packet{}, err
	}

	return packet, nil
}

// decodeEncodedResponse converts the complete bytes produced by a shared encoder
// into the server package's logical packet contract.
func decodeEncodedResponse(data []byte) (a2s.Packet, error) {
	packet, err := a2s.DecodePacket(data)
	if err != nil {
		return a2s.Packet{}, fmt.Errorf("%w: decode encoded response: %w", ErrResponse, err)
	}

	return packet, nil
}

// validateResponseQuery rejects response types that cannot answer query.
func validateResponseQuery(query a2s.QueryType, response a2s.ResponseType) error {
	valid := false
	switch query {
	case a2s.InfoRequest:
		valid = response == a2s.ResponseInfo ||
			response == a2s.ResponseInfoGoldSource ||
			response == a2s.ResponseChallenge

	case a2s.PlayerRequest:
		valid = response == a2s.ResponsePlayers || response == a2s.ResponseChallenge

	case a2s.RulesRequest:
		valid = response == a2s.ResponseRules || response == a2s.ResponseChallenge

	case a2s.ChallengeRequest:
		valid = response == a2s.ResponseChallenge

	case a2s.PingRequest:
		valid = response == a2s.ResponsePing

	default:
		return fmt.Errorf("%w: unsupported query 0x%X", ErrResponseQuery, query)
	}

	if !valid {
		return fmt.Errorf(
			"%w: query 0x%X cannot use response 0x%X",
			ErrResponseQuery,
			query,
			response,
		)
	}

	return nil
}
