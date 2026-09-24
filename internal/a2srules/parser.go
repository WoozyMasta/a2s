// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

// Package a2srules parses the common A2S_RULES key/value envelope.
//
// It deliberately does not interpret A3SB pages or any game-specific fields.
package a2srules

import (
	"errors"

	"github.com/woozymasta/a2s/internal/wire"
)

var (
	// ErrCount indicates that the rule count could not be read.
	ErrCount = errors.New("A2S_RULES: count read failed")

	// ErrInsufficientData indicates that a rule entry cannot contain its required key and value terminators.
	ErrInsufficientData = errors.New("A2S_RULES: insufficient rule data")

	// ErrKey indicates that a rule key is not NUL-terminated.
	ErrKey = errors.New("A2S_RULES: key read failed")

	// ErrValue indicates that a rule value is not NUL-terminated.
	ErrValue = errors.New("A2S_RULES: value read failed")
)

// Entry is one ordered A2S_RULES key/value pair.
//
// Key and Value refer to the input buffer and are valid only while that buffer remains valid.
// Callers that retain either field must copy it.
type Entry struct {
	Key   []byte
	Value []byte
}

// Result contains parsed entries and bytes left after the declared entries.
type Result struct {
	Entries   []Entry
	Remaining []byte
}

// Parse reads the A2S_RULES envelope without interpreting its entries.
// Duplicate keys and entry order are preserved.
// Trailing bytes are returned in Result.Remaining so each protocol package can apply its own policy.
func Parse(data []byte) (Result, error) {
	decoder := wire.NewDecoder(data)
	count, err := decoder.Uint16()
	if err != nil {
		return Result{}, errors.Join(ErrCount, err)
	}

	if count == 0 {
		return Result{Remaining: decoder.Tail()}, nil
	}

	entries := make([]Entry, 0, int(count))
	for i := 0; i < int(count); i++ {
		key, err := decoder.CStringBytes()
		if err != nil {
			return Result{}, errors.Join(ErrKey, err)
		}

		value, err := decoder.CStringBytes()
		if err != nil {
			return Result{}, errors.Join(ErrValue, err)
		}

		entries = append(entries, Entry{Key: key, Value: value})
	}

	return Result{
		Entries:   entries,
		Remaining: decoder.Tail(),
	}, nil
}

// Map converts entries to the public string map representation.
// Duplicate keys use the last value, matching the existing A2S behavior.
func Map(entries []Entry) map[string]string {
	if len(entries) == 0 {
		return nil
	}

	rules := make(map[string]string, len(entries))
	for _, entry := range entries {
		rules[string(entry.Key)] = string(entry.Value)
	}

	return rules
}
