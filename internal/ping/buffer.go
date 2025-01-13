package ping

import (
	"time"
)

const pingBuffSize = 65535

// Stats of ping
type Stats struct {
	Min time.Duration
	Max time.Duration
	Avg time.Duration
}

// Ring buffer for ping stats
type Buffer struct {
	data  []time.Duration
	head  int
	tail  int
	count int
}

// Create new ping ring buffer
func NewBuffer() *Buffer {
	return &Buffer{
		data:  make([]time.Duration, pingBuffSize),
		head:  0,
		tail:  0,
		count: 0,
	}
}

// Add record to ping buff
func (p *Buffer) Add(value time.Duration) {
	p.data[p.tail] = value
	p.tail = (p.tail + 1) % len(p.data)

	if p.count < len(p.data) {
		p.count++
	} else {
		p.head = (p.head + 1) % len(p.data)
	}
}

// Get all record from ping buff
func (p *Buffer) Get() []time.Duration {
	var result []time.Duration

	if p.count == len(p.data) {
		result = append(result, p.data[p.head:]...)
	}
	result = append(result, p.data[:p.tail]...)

	return result
}

// Ping statistics calculation function
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
