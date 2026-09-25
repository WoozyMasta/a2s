// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package server

import (
	"fmt"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

// NormalizeResponse converts a handler response into one logical A2S packet.
//
// Typed responses go through the shared package encoders.
// PacketResponse is already a logical packet
// and is returned without decoding or re-encoding its payload.
func NormalizeResponse(req *Request, response Response) (transport.Packet, error) {
	if req == nil {
		return transport.Packet{}, fmt.Errorf("%w: nil request", ErrResponse)
	}
	if response == nil {
		return transport.Packet{}, fmt.Errorf("%w: nil response", ErrResponse)
	}

	switch response := response.(type) {
	case InfoResponse:
		return normalizeInfoResponse(req.Query.Type, response.Info)

	case *InfoResponse:
		if response == nil {
			return transport.Packet{}, fmt.Errorf("%w: nil InfoResponse", ErrResponse)
		}
		return normalizeInfoResponse(req.Query.Type, response.Info)

	case PlayersResponse:
		return normalizePlayersResponse(req.Query.Type, response.Players)

	case *PlayersResponse:
		if response == nil {
			return transport.Packet{}, fmt.Errorf("%w: nil PlayersResponse", ErrResponse)
		}
		return normalizePlayersResponse(req.Query.Type, response.Players)

	case RulesResponse:
		return normalizeRulesResponse(req.Query.Type, response.Rules)

	case *RulesResponse:
		if response == nil {
			return transport.Packet{}, fmt.Errorf("%w: nil RulesResponse", ErrResponse)
		}
		return normalizeRulesResponse(req.Query.Type, response.Rules)

	case PacketResponse:
		return normalizePacketResponse(req.Query.Type, response.Packet)

	case *PacketResponse:
		if response == nil {
			return transport.Packet{}, fmt.Errorf("%w: nil PacketResponse", ErrResponse)
		}
		return normalizePacketResponse(req.Query.Type, response.Packet)

	default:
		return transport.Packet{}, fmt.Errorf("%w: unsupported response %T", ErrResponse, response)
	}
}

// normalizeInfoResponse encodes an InfoResponse after checking its query kind.
func normalizeInfoResponse(query transport.QueryType, info a2s.Info) (transport.Packet, error) {
	if err := validateResponseQuery(query, transport.ResponseInfo); err != nil {
		return transport.Packet{}, err
	}

	data, err := a2s.AppendInfo(nil, info)
	if err != nil {
		return transport.Packet{}, fmt.Errorf("%w: encode InfoResponse: %w", ErrResponse, err)
	}

	return decodeEncodedResponse(data)
}

// normalizePlayersResponse encodes a PlayersResponse after checking its query kind.
func normalizePlayersResponse(query transport.QueryType, players []a2s.Player) (transport.Packet, error) {
	if err := validateResponseQuery(query, transport.ResponsePlayers); err != nil {
		return transport.Packet{}, err
	}

	data, err := a2s.AppendPlayers(nil, players)
	if err != nil {
		return transport.Packet{}, fmt.Errorf("%w: encode PlayersResponse: %w", ErrResponse, err)
	}

	return decodeEncodedResponse(data)
}

// normalizeRulesResponse encodes a RulesResponse after checking its query kind.
func normalizeRulesResponse(query transport.QueryType, rules a2s.Rules) (transport.Packet, error) {
	if err := validateResponseQuery(query, transport.ResponseRules); err != nil {
		return transport.Packet{}, err
	}

	data, err := a2s.AppendRules(nil, rules)
	if err != nil {
		return transport.Packet{}, fmt.Errorf("%w: encode RulesResponse: %w", ErrResponse, err)
	}

	return decodeEncodedResponse(data)
}

// normalizePacketResponse validates a logical packet without changing it.
func normalizePacketResponse(query transport.QueryType, packet transport.Packet) (transport.Packet, error) {
	if err := validateResponseQuery(query, packet.Type); err != nil {
		return transport.Packet{}, err
	}

	return packet, nil
}

// decodeEncodedResponse converts the complete bytes produced by a shared encoder
// into the server package's logical packet contract.
func decodeEncodedResponse(data []byte) (transport.Packet, error) {
	packet, err := transport.DecodePacket(data)
	if err != nil {
		return transport.Packet{}, fmt.Errorf("%w: decode encoded response: %w", ErrResponse, err)
	}

	return packet, nil
}

// validateResponseQuery rejects response types that cannot answer query.
func validateResponseQuery(query transport.QueryType, response transport.ResponseType) error {
	valid := false
	switch query {
	case transport.InfoRequest:
		valid = response == transport.ResponseInfo ||
			response == transport.ResponseInfoGoldSource ||
			response == transport.ResponseChallenge

	case transport.PlayerRequest:
		valid = response == transport.ResponsePlayers || response == transport.ResponseChallenge

	case transport.RulesRequest:
		valid = response == transport.ResponseRules || response == transport.ResponseChallenge

	case transport.ChallengeRequest:
		valid = response == transport.ResponseChallenge

	case transport.PingRequest:
		valid = response == transport.ResponsePing

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
