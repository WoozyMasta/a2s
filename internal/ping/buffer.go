package ping

import (
	"time"
)

const pingBuffSize = 65535

// Stats holds aggregated ping statistics such as minimum,
// maximum, and average round-trip times.
type Stats struct {
	// Min is the smallest round-trip time observed.
	Min time.Duration

	// Max is the largest round-trip time observed.
	Max time.Duration

	// Avg is the average round-trip time.
	Avg time.Duration
}

// Buffer implements a fixed-size ring buffer for storing ping results.
type Buffer struct {
	data  []time.Duration
	head  int
	tail  int
	count int
}

// NewBuffer creates and initializes a new ring buffer for ping results.
func NewBuffer() *Buffer {
	return &Buffer{
		data:  make([]time.Duration, pingBuffSize),
		head:  0,
		tail:  0,
		count: 0,
	}
}

// Add inserts a new ping result into the buffer, overwriting the oldest entry if full.
func (p *Buffer) Add(value time.Duration) {
	p.data[p.tail] = value
	p.tail = (p.tail + 1) % len(p.data)

	if p.count < len(p.data) {
		p.count++
	} else {
		p.head = (p.head + 1) % len(p.data)
	}
}

// Get returns a slice containing all ping results currently stored in the buffer in insertion order.
func (p *Buffer) Get() []time.Duration {
	var result []time.Duration

	if p.count == len(p.data) {
		result = append(result, p.data[p.head:]...)
	}
	result = append(result, p.data[:p.tail]...)

	return result
}

// CalculateStats computes minimum, maximum, and average statistics for the values stored in the given buffer.
func CalculateStats(buffer *Buffer) Stats {
	pings := buffer.Get()

	if len(pings) == 0 {
		return Stats{Min: 0, Max: 0, Avg: 0}
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

	return Stats{
		Min: minPing,
		Max: maxPing,
		Avg: avgPing,
	}
}
