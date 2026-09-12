// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package proxy

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/woozymasta/a2s/pkg/a2s"
)

// cacheableQueryCount is the number of query types supported by Cache.
const cacheableQueryCount = 3

// Cache stores independently valid logical responses for selected A2S queries.
//
// Cache must not be copied after first use.
// Store and Invalidate publish immutable snapshots;
// Load returns a packet with an independently owned payload.
type Cache struct {
	state   atomic.Pointer[cacheSnapshot] // Current immutable snapshot.
	writeMu sync.Mutex                    // Serializes snapshot publication.
	enabled [cacheableQueryCount]bool     // Enabled cache entries by query index.
}

// cacheSnapshot is immutable after publication through Cache.state.
type cacheSnapshot struct {
	entries [cacheableQueryCount]cacheEntry // Independent state for each cacheable query.
}

// cacheEntry contains one response and its validity state.
type cacheEntry struct {
	packet a2s.Packet // Cached logical response packet.
	valid  bool       // Whether packet may be served.
}

// NewCache creates a cache with the supplied INFO, PLAYER, and RULES entries enabled.
// Duplicate query types are harmless.
// An empty list creates a cache with no enabled entries.
func NewCache(queries []a2s.QueryType) (*Cache, error) {
	cache := &Cache{}
	for _, query := range queries {
		index, ok := cacheIndex(query)
		if !ok {
			return nil, fmt.Errorf("%w: 0x%X", ErrCacheQuery, query)
		}

		cache.enabled[index] = true
	}

	return cache, nil
}

// Enabled reports whether query is configured for caching.
func (c *Cache) Enabled(query a2s.QueryType) bool {
	if c == nil {
		return false
	}

	index, ok := cacheIndex(query)
	return ok && c.enabled[index]
}

// Store publishes a valid response for an enabled query.
//
// The packet payload is cloned before publication.
// The response type must match the query type
// so a cache entry cannot be served for the wrong query.
func (c *Cache) Store(query a2s.QueryType, packet a2s.Packet) error {
	index, err := c.enabledIndex(query)
	if err != nil {
		return err
	}
	if !matchesQuery(query, packet.Type) {
		return fmt.Errorf("%w: query 0x%X with response 0x%X", ErrCachePacket, query, packet.Type)
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	next := cloneSnapshot(c.state.Load())
	next.entries[index] = cacheEntry{
		packet: clonePacket(packet),
		valid:  true,
	}
	c.state.Store(next)

	return nil
}

// Load returns a valid response for an enabled query.
//
// The returned packet owns its payload
// and can be modified by the caller without changing the cache.
func (c *Cache) Load(query a2s.QueryType) (a2s.Packet, bool) {
	if c == nil || !c.Enabled(query) {
		return a2s.Packet{}, false
	}

	index, _ := cacheIndex(query)
	snapshot := c.state.Load()
	if snapshot == nil || !snapshot.entries[index].valid {
		return a2s.Packet{}, false
	}

	return clonePacket(snapshot.entries[index].packet), true
}

// Invalidate marks an enabled query unavailable without affecting other entries.
// The previous packet may remain in an internal snapshot but is never returned.
func (c *Cache) Invalidate(query a2s.QueryType) error {
	index, err := c.enabledIndex(query)
	if err != nil {
		return err
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	next := cloneSnapshot(c.state.Load())
	next.entries[index].valid = false
	c.state.Store(next)

	return nil
}

// enabledIndex validates query support and returns its snapshot index.
func (c *Cache) enabledIndex(query a2s.QueryType) (int, error) {
	if c == nil {
		return 0, fmt.Errorf("%w: cache is nil", ErrCache)
	}

	index, ok := cacheIndex(query)
	if !ok {
		return 0, fmt.Errorf("%w: 0x%X", ErrCacheQuery, query)
	}
	if !c.enabled[index] {
		return 0, fmt.Errorf("%w: 0x%X", ErrCacheDisabled, query)
	}

	return index, nil
}

// cacheIndex maps a supported query type to its snapshot index.
func cacheIndex(query a2s.QueryType) (int, bool) {
	switch query {
	case a2s.InfoRequest:
		return 0, true

	case a2s.PlayerRequest:
		return 1, true

	case a2s.RulesRequest:
		return 2, true

	default:
		return 0, false
	}
}

// matchesQuery reports whether response can answer query.
func matchesQuery(query a2s.QueryType, response a2s.ResponseType) bool {
	switch query {
	case a2s.InfoRequest:
		return response == a2s.ResponseInfo || response == a2s.ResponseInfoGoldSource

	case a2s.PlayerRequest:
		return response == a2s.ResponsePlayers

	case a2s.RulesRequest:
		return response == a2s.ResponseRules

	default:
		return false
	}
}

// cloneSnapshot copies snapshot metadata without sharing mutable packet payloads.
func cloneSnapshot(snapshot *cacheSnapshot) *cacheSnapshot {
	if snapshot == nil {
		return &cacheSnapshot{}
	}

	clone := *snapshot
	return &clone
}

// clonePacket returns a packet with independently owned payload bytes.
func clonePacket(packet a2s.Packet) a2s.Packet {
	packet.Payload = bytes.Clone(packet.Payload)
	return packet
}
