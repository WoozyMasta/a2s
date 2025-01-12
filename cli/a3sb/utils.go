package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/keywords"
)

func createClient(host string, port int, c *cli.Context) (*a2s.Client, error) {
	client, err := a2s.New(host, port)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	if timeout := c.Int("timeout"); timeout > 0 {
		client.SetDeadlineTimeout(timeout)
	}

	if bufferSize := c.Int("buffer-size"); bufferSize > 0 {
		if bufferSize < 0 || bufferSize > 65535 {
			return nil, fmt.Errorf("failed to set buffer size: %d", bufferSize)
		}
		client.SetBufferSize(uint16(bufferSize))
	}

	return client, nil
}

func printJSON(data any) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal().Msgf("Failed to marshal JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}

func printJSONWithDayZ(info *a2s.Info) {
	jsonData, err := json.Marshal(info)
	if err != nil {
		log.Fatal().Msgf("Failed to marshal Info: %v", err)
	}

	// Unmarshal into a map to add custom fields
	var jsonMap map[string]any
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		log.Fatal().Msgf("Failed to unmarshal JSON: %v", err)
	}

	// Add the parsed DayZ structure to the JSON map
	delete(jsonMap, "keywords")
	dayZData := keywords.ParseDayZ(info.Keywords)
	jsonMap["keywords"] = dayZData

	// Marshal back to JSON for output
	updatedJSONData, err := json.MarshalIndent(jsonMap, "", "  ")
	if err != nil {
		log.Fatal().Msgf("Failed to marshal updated JSON: %v", err)
	}

	fmt.Println(string(updatedJSONData))
}

// initialize logging
func initLogging() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		PartsOrder: []string{zerolog.MessageFieldName},
	})
	log.Logger = log.Level(zerolog.ErrorLevel)
}

// setup log level
func setupLogging(level string) {
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		log.Error().Msgf("Undefined log level %s, fallback to error level", level)
		log.Logger = log.Level(zerolog.ErrorLevel)
	}

	log.Logger = log.Level(logLevel)

	if logLevel < zerolog.InfoLevel {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out: os.Stderr,
			PartsOrder: []string{
				zerolog.TimestampFieldName,
				zerolog.LevelFieldName,
				zerolog.CallerFieldName,
				zerolog.MessageFieldName,
			},
			TimeFormat: time.RFC3339,
		})
	} else if logLevel < zerolog.ErrorLevel {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out: os.Stderr,
			PartsOrder: []string{
				zerolog.LevelFieldName,
				zerolog.MessageFieldName,
			},
			TimeFormat: time.RFC3339,
		})
	}

	log.Debug().Msgf("Logger setup with level %s", level)
}
