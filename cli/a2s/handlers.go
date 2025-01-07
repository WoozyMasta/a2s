package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/woozymasta/a2s/pkg/a2s"
)

// processing A2S_INFO request
func handleInfo(client *a2s.Client, c *cli.Context) error {
	info, err := client.GetInfo()
	if err != nil {
		return fmt.Errorf("failed to get server info: %w", err)
	}

	printInfo(info, fmt.Sprintf("%s:%d", host, port), c.Bool("json"))
	return nil
}

// processing A2S_RULES request
func handleRules(client *a2s.Client, c *cli.Context) error {
	if c.Bool("raw") {
		rules, err := client.GetRules()
		if err != nil {
			return fmt.Errorf("failed to get server rules: %w", err)
		}

		printRules(rules, fmt.Sprintf("%s:%d", host, port), c.Bool("json"))
		return nil
	}

	rules, err := client.GetParsedRules()
	if err != nil {
		return fmt.Errorf("failed to get server rules: %w", err)
	}

	printParsedRules(rules, fmt.Sprintf("%s:%d", host, port), c.Bool("json"))

	return nil
}

// processing A2S_PLAYERS request
func handlePlayers(client *a2s.Client, c *cli.Context) error {
	players, err := client.GetPlayers()
	if err != nil {
		return fmt.Errorf("failed to get players: %w", err)
	}

	printPlayers(players, fmt.Sprintf("%s:%d", host, port), c.Bool("json"))
	return nil
}

// processing A2S_INFO request as a source to create a cyclic ping to the server
func handlePing(client *a2s.Client, c *cli.Context) error {
	count := c.Int("ping-count")
	period := time.Duration(c.Int("ping-period"))
	var errorCount int

	// Create a ring buffer
	buffer := newPingRingBuff()

	// Channel for receiving the completion signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool)

	if count != 0 {
		fmt.Printf("Start %d times ping %s:%d with %ds period\n\n", count, host, port, c.Int("ping-period"))
	} else {
		fmt.Printf("Start infinity ping %s:%d with %ds period\n\n", host, port, c.Int("ping-period"))
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

			fmt.Printf("A2S_INFO response server=%s:%d folder=\"%s\" name=\"%s\" time=%s\n", host, port, info.Folder, info.Name, pingDuration)
			time.Sleep(period * time.Second)
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
	stats := calculateStats(buffer)

	// Display statistics
	fmt.Printf("\nTransmitted %d request, received %d response, failed %d\n", buffer.count+errorCount, buffer.count, errorCount)
	if buffer.count >= pingBuffSize {
		fmt.Printf("Requests counter truncated to %d\n", pingBuffSize)
	}
	fmt.Printf("Min=%s Max=%s Avg=%s\n", stats.Min, stats.Max, stats.Avg)

	return nil
}
