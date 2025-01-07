package main

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
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
		client.SetBufferSize(uint16(bufferSize))
	}

	return client, nil
}

func printJson(data any) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}

func printJsonWithDayZ(info *a2s.Info) {
	jsonData, err := json.Marshal(info)
	if err != nil {
		log.Fatalf("Failed to marshal Info: %v", err)
	}

	// Unmarshal into a map to add custom fields
	var jsonMap map[string]any
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Add the parsed DayZ structure to the JSON map
	delete(jsonMap, "keywords")
	dayZData := keywords.ParseDayZ(info.Keywords)
	jsonMap["keywords"] = dayZData

	// Marshal back to JSON for output
	updatedJsonData, err := json.MarshalIndent(jsonMap, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal updated JSON: %v", err)
	}

	fmt.Println(string(updatedJsonData))
}

// initialize logging
func prepareLogging() *log.TextFormatter {
	formatter := log.TextFormatter{
		ForceColors:            true,
		DisableQuote:           false,
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
		PadLevelText:           true,
	}

	log.SetFormatter(&formatter)
	log.SetLevel(log.InfoLevel)
	log.SetOutput(os.Stderr)

	return &formatter
}

// setup log level
func setupLogging(level string, formatter *log.TextFormatter) {
	logLevel, err := log.ParseLevel(level)
	if err != nil {
		log.Errorf("Undefined log level %s, fallback to error level", level)
		logLevel = log.ErrorLevel
	}

	log.SetLevel(logLevel)

	if logLevel == log.DebugLevel {
		formatter.DisableTimestamp = false
		log.SetFormatter(formatter)
	}

	if logLevel == log.TraceLevel {
		formatter.DisableTimestamp = false
		formatter.FullTimestamp = true
		log.SetFormatter(formatter)
		log.SetReportCaller(true)
	}

	log.Debugf("Logger setup with level %s", level)
}
