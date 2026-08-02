package proxy

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
)

func TestNewPollerValidatesConfiguration(t *testing.T) {
	cache, err := NewCache([]a2s.QueryType{a2s.InfoRequest})
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	upstream := &pollerUpstream{}
	tests := []struct {
		name     string
		cache    *Cache
		upstream Upstream
		config   PollerConfig
		wantErr  bool
	}{
		{
			name:     "nil cache",
			upstream: upstream,
			config:   PollerConfig{TTL: time.Second},
			wantErr:  true,
		},
		{
			name:    "nil upstream",
			cache:   cache,
			config:  PollerConfig{TTL: time.Second},
			wantErr: true,
		},
		{
			name:     "zero TTL",
			cache:    cache,
			upstream: upstream,
			config:   PollerConfig{},
			wantErr:  true,
		},
		{
			name:     "negative inactive TTL",
			cache:    cache,
			upstream: upstream,
			config:   PollerConfig{TTL: time.Second, InactiveTTL: -time.Second},
			wantErr:  true,
		},
		{
			name:     "negative jitter",
			cache:    cache,
			upstream: upstream,
			config:   PollerConfig{TTL: time.Second, Jitter: -time.Second},
			wantErr:  true},
		{
			name:     "negative retries",
			cache:    cache,
			upstream: upstream,
			config:   PollerConfig{TTL: time.Second, Retries: -1},
			wantErr:  true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			poller, err := NewPoller(test.cache, test.upstream, test.config)
			if test.wantErr {
				if !errors.Is(err, ErrPoller) {
					t.Fatalf("NewPoller() error = %v, want ErrPoller", err)
				}
				return
			}
			if err != nil || poller == nil {
				t.Fatalf("NewPoller() = %v, %v; want poller", poller, err)
			}
		})
	}
}

func TestNewPollerInheritsInactiveTTL(t *testing.T) {
	cache, err := NewCache([]a2s.QueryType{a2s.InfoRequest})
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	poller, err := NewPoller(cache, &pollerUpstream{}, PollerConfig{TTL: 3 * time.Second})
	if err != nil {
		t.Fatalf("NewPoller() error = %v", err)
	}
	if poller.config.InactiveTTL != poller.config.TTL {
		t.Fatalf("inactive TTL = %s, want %s", poller.config.InactiveTTL, poller.config.TTL)
	}
}

func TestPollerRefreshesColdCacheImmediately(t *testing.T) {
	cache := mustCache(t, a2s.InfoRequest)
	upstream := &pollerUpstream{fn: func(a2s.QueryType, int) (a2s.Packet, error) {
		return infoPacket(), nil
	}}
	poller := mustPoller(t, cache, upstream, PollerConfig{TTL: time.Minute})

	var waits []time.Duration
	ctx, cancel := context.WithCancel(context.Background())
	poller.wait = func(_ context.Context, delay time.Duration) bool {
		waits = append(waits, delay)
		cancel()
		return false
	}

	if err := poller.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := upstream.Count(a2s.InfoRequest); got != 1 {
		t.Fatalf("upstream calls = %d, want 1", got)
	}
	if _, ok := cache.Load(a2s.InfoRequest); !ok {
		t.Fatal("cold cache was not populated")
	}
	if len(waits) != 1 || waits[0] != time.Minute {
		t.Fatalf("waits = %v, want [%s]", waits, time.Minute)
	}
}

func TestPollerRefreshesActiveCacheAfterTTL(t *testing.T) {
	cache := mustCache(t, a2s.InfoRequest)
	if err := cache.Store(a2s.InfoRequest, infoPacket(), time.Time{}); err != nil {
		t.Fatalf("Store() error = %v", err)
	}

	upstream := &pollerUpstream{fn: func(a2s.QueryType, int) (a2s.Packet, error) {
		return a2s.Packet{Type: a2s.ResponseInfo, Payload: []byte("new")}, nil
	}}
	poller := mustPoller(t, cache, upstream, PollerConfig{TTL: 15 * time.Second})

	var waits []time.Duration
	ctx, cancel := context.WithCancel(context.Background())
	poller.wait = func(_ context.Context, delay time.Duration) bool {
		waits = append(waits, delay)
		if len(waits) == 2 {
			cancel()
			return false
		}
		return true
	}

	if err := poller.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := upstream.Count(a2s.InfoRequest); got != 1 {
		t.Fatalf("upstream calls = %d, want 1", got)
	}
	packet, ok := cache.Load(a2s.InfoRequest)
	if !ok || string(packet.Payload) != "new" {
		t.Fatalf("cached packet = %#v, %v; want new packet", packet, ok)
	}
	if len(waits) != 2 || waits[0] != 15*time.Second {
		t.Fatalf("waits = %v, want first wait %s", waits, 15*time.Second)
	}
}

func TestPollerRetryKeepsPreviousEntryUntilSuccess(t *testing.T) {
	cache := mustCache(t, a2s.InfoRequest)
	old := infoPacket()
	if err := cache.Store(a2s.InfoRequest, old, time.Time{}); err != nil {
		t.Fatalf("Store() error = %v", err)
	}

	upstream := &pollerUpstream{fn: func(_ a2s.QueryType, call int) (a2s.Packet, error) {
		if call == 1 {
			return a2s.Packet{}, errors.New("temporary failure")
		}
		return a2s.Packet{Type: a2s.ResponseInfo, Payload: []byte("new")}, nil
	}}
	poller := mustPoller(t, cache, upstream, PollerConfig{
		TTL:     time.Minute,
		Jitter:  time.Nanosecond,
		Retries: 1,
	})

	var waits []time.Duration
	ctx, cancel := context.WithCancel(context.Background())
	poller.wait = func(_ context.Context, delay time.Duration) bool {
		waits = append(waits, delay)
		if len(waits) == 1 {
			packet, ok := cache.Load(a2s.InfoRequest)
			if !ok || string(packet.Payload) != "info" {
				t.Errorf("previous cache entry = %#v, %v during retry", packet, ok)
			}
			return true
		}
		if len(waits) == 3 {
			cancel()
			return false
		}
		return true
	}

	if err := poller.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := upstream.Count(a2s.InfoRequest); got != 2 {
		t.Fatalf("upstream calls = %d, want 2", got)
	}
	if len(waits) != 3 || waits[1] < pollerRetryDelay || waits[1] > pollerRetryDelay+time.Nanosecond {
		t.Fatalf("retry waits = %v, want %s plus jitter", waits, pollerRetryDelay)
	}
}

func TestPollerExhaustedRetriesInvalidateEntry(t *testing.T) {
	cache := mustCache(t, a2s.InfoRequest)
	if err := cache.Store(a2s.InfoRequest, infoPacket(), time.Time{}); err != nil {
		t.Fatalf("Store() error = %v", err)
	}

	upstream := &pollerUpstream{fn: func(a2s.QueryType, int) (a2s.Packet, error) {
		return a2s.Packet{}, errors.New("offline")
	}}
	changes := make(chan StateChange, 1)
	poller := mustPoller(t, cache, upstream, PollerConfig{
		TTL:           time.Minute,
		InactiveTTL:   time.Hour,
		Retries:       2,
		OnStateChange: func(change StateChange) { changes <- change },
	})

	var waits int
	ctx, cancel := context.WithCancel(context.Background())
	poller.wait = func(_ context.Context, _ time.Duration) bool {
		waits++
		if waits == 4 {
			cancel()
			return false
		}
		return true
	}

	if err := poller.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := upstream.Count(a2s.InfoRequest); got != 3 {
		t.Fatalf("upstream calls = %d, want initial plus 2 retries", got)
	}
	if _, ok := cache.Load(a2s.InfoRequest); ok {
		t.Fatal("exhausted query remained available")
	}
	select {
	case change := <-changes:
		if change.Query != a2s.InfoRequest || change.Active || change.Err == nil {
			t.Fatalf("state change = %#v, want inactive with error", change)
		}
	default:
		t.Fatal("inactive transition was not reported")
	}
}

func TestPollerInactiveProbeDoesNotRetryAndRecovers(t *testing.T) {
	cache := mustCache(t, a2s.InfoRequest)
	upstream := &pollerUpstream{fn: func(_ a2s.QueryType, call int) (a2s.Packet, error) {
		if call == 1 {
			return a2s.Packet{}, errors.New("offline")
		}
		return infoPacket(), nil
	}}
	var changes []StateChange
	var changeMu sync.Mutex
	poller := mustPoller(t, cache, upstream, PollerConfig{
		TTL:         time.Minute,
		InactiveTTL: time.Hour,
		OnStateChange: func(change StateChange) {
			changeMu.Lock()
			changes = append(changes, change)
			changeMu.Unlock()
		},
	})

	var waits int
	ctx, cancel := context.WithCancel(context.Background())
	poller.wait = func(_ context.Context, _ time.Duration) bool {
		waits++
		if waits == 2 {
			cancel()
			return false
		}
		return true
	}

	if err := poller.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := upstream.Count(a2s.InfoRequest); got != 2 {
		t.Fatalf("upstream calls = %d, want initial failure plus one probe", got)
	}
	if _, ok := cache.Load(a2s.InfoRequest); !ok {
		t.Fatal("recovered query was not cached")
	}
	changeMu.Lock()
	defer changeMu.Unlock()
	if len(changes) != 2 || changes[0].Active || !changes[1].Active || changes[1].Err != nil {
		t.Fatalf("state changes = %#v, want inactive then recovered", changes)
	}
}

func TestPollerTreatsDeadlineAsRuntimeState(t *testing.T) {
	cache := mustCache(t, a2s.InfoRequest)
	upstream := &pollerUpstream{fn: func(_ a2s.QueryType, call int) (a2s.Packet, error) {
		if call == 1 {
			return a2s.Packet{}, context.DeadlineExceeded
		}
		return infoPacket(), nil
	}}
	poller := mustPoller(t, cache, upstream, PollerConfig{TTL: time.Minute})

	var waits int
	ctx, cancel := context.WithCancel(context.Background())
	poller.wait = func(_ context.Context, _ time.Duration) bool {
		waits++
		if waits == 2 {
			cancel()
			return false
		}
		return true
	}

	if err := poller.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got := upstream.Count(a2s.InfoRequest); got != 2 {
		t.Fatalf("upstream calls = %d, want 2", got)
	}
}

func TestPollerCancellationInterruptsUpstreamQuery(t *testing.T) {
	cache := mustCache(t, a2s.InfoRequest)
	upstream := &blockingPollerUpstream{started: make(chan struct{})}
	poller := mustPoller(t, cache, upstream, PollerConfig{TTL: time.Minute})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- poller.Run(ctx)
	}()

	select {
	case <-upstream.started:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("poller did not start upstream query")
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("poller did not stop after context cancellation")
	}
}

func TestPollerJitterIsBounded(t *testing.T) {
	const max = time.Second
	for i := 0; i < 100; i++ {
		got := randomJitter(max)
		if got < 0 || got > max {
			t.Fatalf("randomJitter() = %s, want [0, %s]", got, max)
		}
	}
	if randomJitter(0) != 0 {
		t.Fatal("randomJitter(0) is not zero")
	}
}

func mustCache(t *testing.T, queries ...a2s.QueryType) *Cache {
	t.Helper()
	cache, err := NewCache(queries)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}
	return cache
}

func mustPoller(t *testing.T, cache *Cache, upstream Upstream, config PollerConfig) *Poller {
	t.Helper()
	poller, err := NewPoller(cache, upstream, config)
	if err != nil {
		t.Fatalf("NewPoller() error = %v", err)
	}
	return poller
}

type pollerUpstream struct {
	mu    sync.Mutex
	calls map[a2s.QueryType]int
	fn    func(a2s.QueryType, int) (a2s.Packet, error)
}

func (u *pollerUpstream) Query(_ context.Context, query a2s.QueryType) (a2s.Packet, a2s.QueryMeta, error) {
	u.mu.Lock()
	if u.calls == nil {
		u.calls = make(map[a2s.QueryType]int)
	}
	u.calls[query]++
	call := u.calls[query]
	u.mu.Unlock()

	if u.fn == nil {
		return infoPacket(), a2s.QueryMeta{}, nil
	}
	packet, err := u.fn(query, call)
	return packet, a2s.QueryMeta{}, err
}

func (u *pollerUpstream) Count(query a2s.QueryType) int {
	u.mu.Lock()
	defer u.mu.Unlock()

	return u.calls[query]
}

type blockingPollerUpstream struct {
	started chan struct{}
	once    sync.Once
}

func (u *blockingPollerUpstream) Query(ctx context.Context, _ a2s.QueryType) (a2s.Packet, a2s.QueryMeta, error) {
	u.once.Do(func() { close(u.started) })
	<-ctx.Done()
	return a2s.Packet{}, a2s.QueryMeta{}, ctx.Err()
}
