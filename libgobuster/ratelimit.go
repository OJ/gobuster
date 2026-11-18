package libgobuster

import (
	"context"
	"sync"
	"time"
)

// rateLimitGate coordinates all workers sharing a HTTPClient when the target
// answers with 429 Too Many Requests.
//
// The first request that receives a 429 becomes the leader of a pause: the gate
// closes, so no other worker sends new requests. The leader sleeps for the
// configured duration and then retries its own request as a probe. While the
// probe keeps receiving 429 the leader keeps sleeping. Once the probe succeeds
// the gate opens again and every worker retries the request that was rate
// limited.
type rateLimitGate struct {
	sleep time.Duration

	mu sync.Mutex
	// paused is non-nil while a pause is active and is closed when it ends.
	paused chan struct{}
	// generation is incremented every time a new pause starts. It lets a request
	// that was sent before the latest pause tell its 429 apart from a fresh one.
	generation uint64
}

func newRateLimitGate(sleep time.Duration) *rateLimitGate {
	return &rateLimitGate{sleep: sleep}
}

// wait blocks while a pause is active. It returns the current generation which
// should be passed to tryPause if the subsequent request is rate limited.
func (r *rateLimitGate) wait(ctx context.Context) (uint64, error) {
	for {
		r.mu.Lock()
		paused := r.paused
		gen := r.generation
		r.mu.Unlock()

		if paused == nil {
			return gen, nil
		}

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-paused:
		}
	}
}

// tryPause is called after a request sent in generation gen received a 429.
// It returns true if the caller became the leader of a new pause. If a pause is
// active or one already happened since the request was sent, it returns false
// and the caller should simply wait and retry.
func (r *rateLimitGate) tryPause(gen uint64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.paused != nil || r.generation != gen {
		return false
	}

	r.paused = make(chan struct{})
	r.generation++
	return true
}

// resume ends the current pause and releases all waiting workers.
func (r *rateLimitGate) resume() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.paused != nil {
		close(r.paused)
		r.paused = nil
	}
}

// sleepFor sleeps for the configured duration or until the context is done.
func (r *rateLimitGate) sleepFor(ctx context.Context) error {
	t := time.NewTimer(r.sleep)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
