package ping

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
)

// processing A2S_INFO request as a source to create a cyclic ping to the server
func Start(client *a2s.Client, count, period int) {
	var errorCount int

	// Create a ring buffer
	buffer := NewBuffer()

	// Channel for receiving the completion signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool)

	if count != 0 {
		fmt.Printf("Start %d times ping %s with %ds period\n\n", count, client.Address, period)
	} else {
		fmt.Printf("Start infinity ping %s with %ds period\n\n", client.Address, period)
	}

	// Starting the main ping loop in a goroutine
	go func() {
		for i := 0; count == 0 || i < count; i++ {
			info, err := client.GetInfo()
			if err != nil {
				log.Printf("Failed to get ping: %v", err)
				errorCount++
				continue
			}

			// Write ping to the ring buffer
			pingDuration := info.Ping
			buffer.Add(pingDuration)

			fmt.Printf("A2S_INFO response server=%s folder=\"%s\" name=\"%s\" time=%s\n", client.Address, info.Folder, info.Name, pingDuration)
			time.Sleep(time.Duration(period) * time.Second)
		}
		done <- true
	}()

	// Waiting for the completion signal
	select {
	case <-signalChan:
		fmt.Println("Received signal, stopping...")
	case <-done:
	}

	// Calculating statistics from the ring buffer
	stats := CalculateStats(buffer)

	// Display statistics
	fmt.Printf("\nTransmitted %d request, received %d response, failed %d\n", buffer.count+errorCount, buffer.count, errorCount)
	if buffer.count >= pingBuffSize {
		fmt.Printf("Requests counter truncated to %d\n", pingBuffSize)
	}
	fmt.Printf("Min=%s Max=%s Avg=%s\n", stats.Min, stats.Max, stats.Avg)
}
