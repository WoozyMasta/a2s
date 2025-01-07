package main

import (
	"time"
)

const pingBuffSize = 65535

type PingStats struct {
	Min time.Duration
	Max time.Duration
	Avg time.Duration
}

type PingRingBuff struct {
	data  []time.Duration
	head  int
	tail  int
	count int
}

func newPingRingBuff() *PingRingBuff {
	return &PingRingBuff{
		data:  make([]time.Duration, pingBuffSize),
		head:  0,
		tail:  0,
		count: 0,
	}
}

func (r *PingRingBuff) Add(value time.Duration) {
	r.data[r.tail] = value
	r.tail = (r.tail + 1) % len(r.data)

	if r.count < len(r.data) {
		r.count++
	} else {
		r.head = (r.head + 1) % len(r.data)
	}
}

func (r *PingRingBuff) GetAll() []time.Duration {
	var result []time.Duration

	if r.count == len(r.data) {
		result = append(result, r.data[r.head:]...)
	}
	result = append(result, r.data[:r.tail]...)

	return result
}

// Ping statistics calculation function
func calculateStats(buffer *PingRingBuff) PingStats {
	pings := buffer.GetAll()

	if len(pings) == 0 {
		return PingStats{Min: 0, Max: 0, Avg: 0}
	}

	minPing, maxPing := pings[0], pings[0]
	totalPing := time.Duration(0)

	for _, ping := range pings {
		if ping < minPing {
			minPing = ping
		}
		if ping > maxPing {
			maxPing = ping
		}
		totalPing += ping
	}

	avgPing := totalPing / time.Duration(len(pings))

	return PingStats{
		Min: minPing,
		Max: maxPing,
		Avg: avgPing,
	}
}
