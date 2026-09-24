// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

// Package ping runs cyclic A2S query requests,
// accumulates response-time statistics,
// and prints a report when the run completes.
package ping

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
)

// Output configures destinations and localization for a ping run.
type Output struct {
	// Out receives normal progress and statistics output.
	Out io.Writer

	// Err receives timestamped request errors.
	Err io.Writer

	// Localize resolves a message and applies printf-style arguments.
	Localize func(key, fallback string, args ...any) string

	// LocalizeLabel resolves a label without applying printf formatting.
	LocalizeLabel func(key, fallback string) string

	// FormatGameID formats the effective game ID for human-readable output.
	FormatGameID func(uint64) string

	// Compact prints only successful response times in milliseconds.
	Compact bool

	// NoSummary suppresses final request and latency statistics.
	NoSummary bool
}

// writeLine writes best-effort progress output to the configured destination.
func writeLine(out io.Writer, values ...any) {
	_, _ = fmt.Fprintln(out, values...)
}

// Start sends selected A2S requests until count is reached
// or a termination signal is received,
// then prints statistics for successful responses.
func Start(client *a2s.Client, count int, period time.Duration, requestType a2s.QueryType, output Output) {
	var errorCount int
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if output.Out == nil {
		output.Out = os.Stdout
	}
	if output.Err == nil {
		output.Err = os.Stderr
	}
	if output.Localize == nil {
		output.Localize = func(_ string, fallback string, args ...any) string {
			return fmt.Sprintf(fallback, args...)
		}
	}
	logger := log.New(output.Err, "", log.LstdFlags)

	// Create a ring buffer
	buffer := NewBuffer()

	if !output.Compact && count != 0 {
		writeLine(
			output.Out,
			output.Localize(
				"ping.start_count",
				"Start %d ping requests with interval: %s",
				count,
				period,
			),
		)
	} else if !output.Compact {
		writeLine(
			output.Out,
			output.Localize(
				"ping.start_infinite",
				"Start infinite ping with interval: %s",
				period,
			),
		)
	}
	if !output.Compact {
		writeLine(output.Out)
	}

	interrupted := false
	metadataPrinted := false
	for i := 0; count == 0 || i < count; i++ {
		info, meta, err := query(ctx, client, requestType)
		if err != nil {
			if ctx.Err() != nil {
				interrupted = true
				break
			}

			logger.Printf(
				"%s",
				output.Localize("ping.error", "Failed to get ping: %s", formatLocalizedPingError(err, output.Localize)),
			)
			errorCount++
			continue
		}

		// Write the ping result to the ring buffer.
		pingDuration := meta.Duration
		buffer.Add(pingDuration)

		if output.Compact {
			writeLine(output.Out, formatMilliseconds(pingDuration))
		} else {
			if requestType == a2s.InfoRequest && !metadataPrinted {
				writeServerMetadata(output, client, info)
				writeLine(output.Out)
				metadataPrinted = true
			}

			writeLine(
				output.Out,
				output.Localize(
					"ping.response",
					"Response #%d: %s",
					buffer.count,
					pingDuration,
				),
			)
		}

		if !waitPeriod(ctx, period) {
			interrupted = true
			break
		}
	}

	if interrupted {
		writeLine(output.Out)
		writeLine(output.Out, output.Localize("ping.interrupted", "Received signal, stopping..."))
	}
	if output.NoSummary {
		return
	}

	// Calculating statistics from the ring buffer
	stats := CalculateStats(buffer)
	successRate := successPercentage(buffer.count, errorCount)
	writeLine(output.Out)

	// Display statistics
	writeLine(
		output.Out,
		output.Localize(
			"ping.summary",
			"Transmitted %d requests, received %d responses, failed %d",
			buffer.count+errorCount,
			buffer.count,
			errorCount,
		),
	)

	if buffer.count >= pingBuffSize {
		writeLine(
			output.Out,
			output.Localize("ping.truncated", "Request counter truncated to %d", pingBuffSize),
		)
	}

	writeLine(
		output.Out,
		output.Localize(
			"ping.stats",
			"Min=%s Max=%s Avg=%s Success: %d%%",
			stats.Min,
			stats.Max,
			stats.Avg,
			successRate,
		),
	)
}

// query executes and validates one selected A2S response.
func query(ctx context.Context, client *a2s.Client, requestType a2s.QueryType) (*a2s.Info, a2s.QueryMeta, error) {
	packet, meta, err := client.Query(ctx, requestType)
	if err != nil {
		return nil, a2s.QueryMeta{}, err
	}

	switch requestType {
	case a2s.InfoRequest:
		info, err := a2s.DecodeInfo(packet)
		return info, meta, err

	case a2s.PlayerRequest:
		_, err := a2s.DecodePlayers(packet)
		return nil, meta, err

	case a2s.RulesRequest:
		_, err := a2s.DecodeRules(packet)
		return nil, meta, err

	default:
		return nil, a2s.QueryMeta{}, fmt.Errorf("unsupported ping query type: 0x%02x", byte(requestType))
	}
}

// writeServerMetadata prints stable fields from the first successful response.
func writeServerMetadata(output Output, client *a2s.Client, info *a2s.Info) {
	writeInfoField(output, "ping.server", "Server:", client.Addr().String())

	gameID := strconv.FormatUint(info.EffectiveID(), 10)
	if output.FormatGameID != nil {
		gameID = output.FormatGameID(info.EffectiveID())
	}

	writeInfoField(output, "info.game_id", "Game ID:", gameID)
	writeInfoField(output, "info.game_folder", "Game folder:", info.Folder)
	writeInfoField(output, "info.server_name", "Server name:", info.Name)
	writeInfoField(output, "info.map", "Map on server:", info.Map)
}

// writeInfoField prints one localized label and its value.
func writeInfoField(output Output, key, fallback, value string) {
	label := fallback
	if output.LocalizeLabel != nil {
		label = output.LocalizeLabel(key, fallback)
	}

	writeLine(output.Out, label+" "+value)
}

// formatMilliseconds formats a duration as a unitless millisecond value.
func formatMilliseconds(value time.Duration) string {
	return strconv.FormatFloat(float64(value)/float64(time.Millisecond), 'f', -1, 64)
}

// successPercentage calculates the integer percentage of successful requests.
func successPercentage(received, failed int) int {
	total := received + failed
	if total == 0 {
		return 0
	}

	return received * 100 / total
}

// waitPeriod waits between ping requests while remaining interruptible.
func waitPeriod(ctx context.Context, period time.Duration) bool {
	timer := time.NewTimer(period)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// formatPingError shortens expected timeout diagnostics while preserving other errors.
func formatPingError(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}

	return err.Error()
}

// formatLocalizedPingError translates timeout diagnostics while preserving other causes.
func formatLocalizedPingError(err error, localize func(key, fallback string, args ...any) string) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return localize("error.timeout_short", "timeout")
	}

	return formatPingError(err)
}
