// Package ping runs cyclic A2S_INFO requests,
// accumulates response-time statistics,
// and prints a report when the run completes.
package ping

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
)

// Start sends A2S_INFO requests until count is reached or a termination signal is received,
// then prints statistics for successful responses.
func Start(client *a2s.Client, count, period int) {
	var errorCount int
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create a ring buffer
	buffer := NewBuffer()

	if count != 0 {
		fmt.Printf("Start %d times ping %s with %ds period\n\n", count, client.Addr(), period)
	} else {
		fmt.Printf("Start infinity ping %s with %ds period\n\n", client.Addr(), period)
	}

	interrupted := false
	for i := 0; count == 0 || i < count; i++ {
		info, meta, err := client.GetInfoWithMeta(ctx)
		if err != nil {
			if ctx.Err() != nil {
				interrupted = true
				break
			}

			log.Printf("Failed to get ping: %s", formatPingError(err))
			errorCount++
			continue
		}

		// Write the ping result to the ring buffer.
		pingDuration := meta.Duration
		buffer.Add(pingDuration)

		fmt.Printf(
			"A2S_INFO response server=%s folder=\"%s\" name=\"%s\" time=%s\n",
			client.Addr(), info.Folder, info.Name, pingDuration)

		if !waitPeriod(ctx, period) {
			interrupted = true
			break
		}
	}

	if interrupted {
		fmt.Println("Received signal, stopping...")
	}

	// Calculating statistics from the ring buffer
	stats := CalculateStats(buffer)

	// Display statistics
	fmt.Printf(
		"\nTransmitted %d request, received %d response, failed %d\n",
		buffer.count+errorCount, buffer.count, errorCount)

	if buffer.count >= pingBuffSize {
		fmt.Printf("Requests counter truncated to %d\n", pingBuffSize)
	}

	fmt.Printf("Min=%s Max=%s Avg=%s\n", stats.Min, stats.Max, stats.Avg)
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
