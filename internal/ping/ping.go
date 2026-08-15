// Package ping runs cyclic A2S_INFO requests,
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
}

// writeLine writes best-effort progress output to the configured destination.
func writeLine(out io.Writer, values ...any) {
	_, _ = fmt.Fprintln(out, values...)
}

// Start sends A2S_INFO requests until count is reached or a termination signal is received,
// then prints statistics for successful responses.
func Start(client *a2s.Client, count, period int, output Output) {
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

	if count != 0 {
		writeLine(
			output.Out,
			output.Localize(
				"ping.start_count",
				"Start %d ping requests to %s with %ds period",
				count,
				client.Addr(),
				period,
			),
		)
	} else {
		writeLine(
			output.Out,
			output.Localize(
				"ping.start_infinite",
				"Start infinite ping to %s with %ds period",
				client.Addr(),
				period,
			),
		)
	}
	writeLine(output.Out)

	interrupted := false
	for i := 0; count == 0 || i < count; i++ {
		info, meta, err := client.GetInfoWithMeta(ctx)
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

		writeLine(
			output.Out,
			output.Localize(
				"ping.response",
				"A2S_INFO response server=%s folder=\"%s\" name=\"%s\" time=%s",
				client.Addr(),
				info.Folder,
				info.Name,
				pingDuration,
			),
		)

		if !waitPeriod(ctx, period) {
			interrupted = true
			break
		}
	}

	if interrupted {
		writeLine(output.Out, output.Localize("ping.interrupted", "Received signal, stopping..."))
	}

	// Calculating statistics from the ring buffer
	stats := CalculateStats(buffer)

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
		output.Localize("ping.stats", "Min=%s Max=%s Avg=%s", stats.Min, stats.Max, stats.Avg),
	)
}

// waitPeriod waits between ping requests while remaining interruptible.
func waitPeriod(ctx context.Context, period int) bool {
	timer := time.NewTimer(time.Duration(period) * time.Second)
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
