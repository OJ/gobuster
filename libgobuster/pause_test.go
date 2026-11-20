package libgobuster

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewPauseController(t *testing.T) {
	t.Parallel()
	pc := NewPauseController()
	if pc == nil {
		t.Fatal("NewPauseController returned nil")
	}
	if pc.IsPaused() {
		t.Fatal("NewPauseController should start in unpaused state")
	}
}

func TestPauseControllerPause(t *testing.T) {
	t.Parallel()
	pc := NewPauseController()

	pc.Pause()
	if !pc.IsPaused() {
		t.Fatal("IsPaused should return true after Pause()")
	}

	pc.Pause()
	if !pc.IsPaused() {
		t.Fatal("IsPaused should still return true after double Pause()")
	}
}

func TestPauseControllerResume(t *testing.T) {
	t.Parallel()
	pc := NewPauseController()

	pc.Pause()
	pc.Resume()
	if pc.IsPaused() {
		t.Fatal("IsPaused should return false after Resume()")
	}

	pc.Resume()
	if pc.IsPaused() {
		t.Fatal("IsPaused should still return false after double Resume()")
	}
}

func TestPauseControllerToggle(t *testing.T) {
	t.Parallel()
	pc := NewPauseController()

	paused := pc.Toggle()
	if !paused {
		t.Fatal("Toggle should return true when transitioning to paused")
	}
	if !pc.IsPaused() {
		t.Fatal("IsPaused should return true after Toggle()")
	}

	paused = pc.Toggle()
	if paused {
		t.Fatal("Toggle should return false when transitioning to unpaused")
	}
	if pc.IsPaused() {
		t.Fatal("IsPaused should return false after second Toggle()")
	}
}

func TestPauseControllerWaitNotPaused(t *testing.T) {
	t.Parallel()
	pc := NewPauseController()
	ctx := context.Background()

	start := time.Now()
	err := pc.Wait(ctx)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Wait should return nil when not paused, got: %v", err)
	}
	if elapsed > 10*time.Millisecond {
		t.Fatalf("Wait should return immediately when not paused, took: %v", elapsed)
	}
}

func TestPauseControllerWaitPausedThenResume(t *testing.T) {
	t.Parallel()
	pc := NewPauseController()
	ctx := context.Background()

	pc.Pause()

	var wg sync.WaitGroup
	var waitErr error
	var waitDuration time.Duration

	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()
		waitErr = pc.Wait(ctx)
		waitDuration = time.Since(start)
	}()

	time.Sleep(50 * time.Millisecond)
	pc.Resume()
	wg.Wait()

	if waitErr != nil {
		t.Fatalf("Wait should return nil after resume, got: %v", waitErr)
	}
	if waitDuration < 40*time.Millisecond {
		t.Fatalf("Wait should have blocked for at least 40ms, took: %v", waitDuration)
	}
}

func TestPauseControllerWaitContextCancel(t *testing.T) {
	t.Parallel()
	pc := NewPauseController()
	ctx, cancel := context.WithCancel(context.Background())

	pc.Pause()

	var wg sync.WaitGroup
	var waitErr error
	var waitDuration time.Duration

	wg.Add(1)
	go func() {
		defer wg.Done()
		start := time.Now()
		waitErr = pc.Wait(ctx)
		waitDuration = time.Since(start)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()
	wg.Wait()

	if waitErr != context.Canceled {
		t.Fatalf("Wait should return context.Canceled, got: %v", waitErr)
	}
	if waitDuration < 40*time.Millisecond {
		t.Fatalf("Wait should have blocked for at least 40ms, took: %v", waitDuration)
	}
}

func TestPauseControllerConcurrency(t *testing.T) {
	t.Parallel()
	pc := NewPauseController()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				pc.Pause()
				_ = pc.IsPaused()
				pc.Resume()
			}
		}()
	}

	wg.Wait()
}
