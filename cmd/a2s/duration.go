// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"errors"
	"math"
	"strconv"
	"time"
)

// Duration is a CLI duration that treats a bare number as seconds.
type Duration time.Duration

// UnmarshalFlag parses a standard duration or a bare number of seconds.
func (d *Duration) UnmarshalFlag(value string) error {
	parsed, err := time.ParseDuration(value)
	if err != nil {
		seconds, parseErr := strconv.ParseFloat(value, 64)
		if parseErr != nil || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
			return err
		}

		nanoseconds := seconds * float64(time.Second)
		if nanoseconds < math.MinInt64 || nanoseconds > math.MaxInt64 {
			return errors.New("duration is out of range")
		}

		parsed = time.Duration(nanoseconds)
	}

	*d = Duration(parsed)
	return nil
}

// MarshalFlag formats the duration for help and generated documentation.
//
//nolint:unparam // flags.Marshaler requires an error result.
func (d Duration) MarshalFlag() (string, error) {
	return time.Duration(d).String(), nil
}

// String formats the duration using the standard Go representation.
func (d Duration) String() string {
	return time.Duration(d).String()
}
