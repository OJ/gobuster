package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/OJ/gobuster/v3/libgobuster"
)

const ruler = "==============================================================="
const cliProgressUpdate = 500 * time.Millisecond

func banner() {
	fmt.Printf("Gobuster v%s\n", libgobuster.VERSION)
	fmt.Println("by OJ Reeves (@TheColonial) & Christian Mehlmauer (@firefart)")
}

// resultWorker outputs the results as they come in. This needs to be a range and should not handle
// the context so the channel always has a receiver and libgobuster will not block.
func resultWorker(g *libgobuster.Gobuster, filename string, wg *sync.WaitGroup) {
	defer wg.Done()

	var f *os.File
	var err error
	if filename != "" {
		f, err = os.Create(filename)
		if err != nil {
			g.LogError.Fatalf("error on creating output file: %v", err)
		}
		defer f.Close()
	}

	for r := range g.Progress.ResultChan {
		s, err := r.ResultToString()
		if err != nil {
			g.LogError.Fatal(err)
		}
		if s != "" {
			s = strings.TrimSpace(s)
			_, _ = fmt.Printf("%s%s\n", TERMINAL_CLEAR_LINE, s)
			if f != nil {
				err = writeToFile(f, s)
				if err != nil {
					g.LogError.Fatalf("error on writing output file: %v", err)
				}
			}
		}
	}
}

// errorWorker outputs the errors as they come in. This needs to be a range and should not handle
// the context so the channel always has a receiver and libgobuster will not block.
func errorWorker(g *libgobuster.Gobuster, wg *sync.WaitGroup) {
	defer wg.Done()

	for e := range g.Progress.ErrorChan {
		if !g.Opts.Quiet && !g.Opts.NoError {
			g.LogError.Printf("[!] %s\n", e.Error())
		}
	}
}

// progressWorker outputs the progress every tick. It will stop once cancel() is called
// on the context
func progressWorker(ctx context.Context, g *libgobuster.Gobuster, wg *sync.WaitGroup) {
	defer wg.Done()

	tick := time.NewTicker(cliProgressUpdate)

	for {
		select {
		case <-tick.C:
			if !g.Opts.Quiet && !g.Opts.NoProgress {
				requestsIssued := g.Progress.RequestsIssued()
				requestsExpected := g.Progress.RequestsExpected()
				if g.Opts.Wordlist == "-" {
					s := fmt.Sprintf("%sProgress: %d", TERMINAL_CLEAR_LINE, requestsIssued)
					_, _ = fmt.Fprint(os.Stderr, s)
					// only print status if we already read in the wordlist
				} else if requestsExpected > 0 {
					s := fmt.Sprintf("%sProgress: %d / %d (%3.2f%%)", TERMINAL_CLEAR_LINE, requestsIssued, requestsExpected, float32(requestsIssued)*100.0/float32(requestsExpected))
					_, _ = fmt.Fprint(os.Stderr, s)
				}
			}
		case <-ctx.Done():
			fmt.Println()
			return
		}
	}
}

func writeToFile(f *os.File, output string) error {
	_, err := f.WriteString(fmt.Sprintf("%s\n", output))
	if err != nil {
		return fmt.Errorf("[!] Unable to write to file %w", err)
	}
	return nil
}

// Gobuster is the main entry point for the CLI
func Gobuster(ctx context.Context, opts *libgobuster.Options, plugin libgobuster.GobusterPlugin) error {
	// Sanity checks
	if opts == nil {
		return fmt.Errorf("please provide valid options")
	}

	if plugin == nil {
		return fmt.Errorf("please provide a valid plugin")
	}

	ctxCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	gobuster, err := libgobuster.NewGobuster(opts, plugin)
	if err != nil {
		return err
	}

	if !opts.Quiet {
		fmt.Println(ruler)
		banner()
		fmt.Println(ruler)
		c, err := gobuster.GetConfigString()
		if err != nil {
			return fmt.Errorf("error on creating config string: %w", err)
		}
		fmt.Println(c)
		fmt.Println(ruler)
		gobuster.LogInfo.Printf("Starting gobuster in %s mode", plugin.Name())
		if opts.WordlistOffset > 0 {
			gobuster.LogInfo.Printf("Skipping the first %d elements...", opts.WordlistOffset)
		}
		fmt.Println(ruler)
	}

	// our waitgroup for all goroutines
	// this ensures all goroutines are finished
	// when we call wg.Wait()
	var wg sync.WaitGroup

	wg.Add(1)
	go resultWorker(gobuster, opts.OutputFilename, &wg)

	wg.Add(1)
	go errorWorker(gobuster, &wg)

	if !opts.Quiet && !opts.NoProgress {
		// if not quiet add a new workgroup entry and start the goroutine
		wg.Add(1)
		go progressWorker(ctxCancel, gobuster, &wg)
	}

	err = gobuster.Run(ctxCancel)

	// call cancel func so progressWorker will exit (the only goroutine in this
	// file using the context) and to free resources
	cancel()
	// wait for all spun up goroutines to finish (all have to call wg.Done())
	wg.Wait()

	// Late error checking to finish all threads
	if err != nil {
		return err
	}

	if !opts.Quiet {
		fmt.Println(ruler)
		gobuster.LogInfo.Println("Finished")
		fmt.Println(ruler)
	}
	return nil
}
