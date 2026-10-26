//go:build !windows

package cli

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/OJ/gobuster/v3/libgobuster"
	"golang.org/x/term"
)

func StartKeyboardListener(ctx context.Context, g *libgobuster.Gobuster, cancel context.CancelFunc) func() {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return func() {}
	}

	var restoreOnce sync.Once
	restoreTerminal := func() {
		restoreOnce.Do(func() {
			_ = term.Restore(fd, oldState)
		})
	}

	// restore terminal on signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case sig := <-sigCh:
			restoreTerminal()
			signal.Stop(sigCh)
			// re-raise signal
			p, _ := os.FindProcess(os.Getpid())
			_ = p.Signal(sig.(syscall.Signal))
		case <-ctx.Done():
			signal.Stop(sigCh)
		}
	}()

	// read keys and send to channel
	keyCh := make(chan byte, 1)
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				return
			}
			select {
			case keyCh <- buf[0]:
			case <-ctx.Done():
				return
			}
		}
	}()

	// handle key events
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case key := <-keyCh:
				if key == ' ' {
					g.Pause.Toggle()
				}
				// Ctrl+C
				if key == 3 {
					restoreTerminal()
					cancel()
					return
				}
			}
		}
	}()

	return restoreTerminal
}
