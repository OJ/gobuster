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

type outputType struct {
	Mu              *sync.RWMutex
	MaxCharsWritten int
}

// right pad a string
// nolint:unparam
func rightPad(s string, padStr string, overallLen int) string {
	strLen := len(s)
	if overallLen <= strLen {
		return s
	}

	toPad := overallLen - strLen - 1
	pad := strings.Repeat(padStr, toPad)
	return fmt.Sprintf("%s%s", s, pad)
}

// resultWorker outputs the results as they come in. This needs to be a range and should not handle
// the context so the channel always has a receiver and libgobuster will not block.
func resultWorker(g *libgobuster.Gobuster, filename string, wg *sync.WaitGroup, output *outputType) {
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

	for r := range g.Results() {
		s, err := r.ResultToString()
		if err != nil {
			g.LogError.Fatal(err)
		}
		if s != "" {
			s = strings.TrimSpace(s)
			output.Mu.Lock()
			w, _ := fmt.Printf("\r%s\n", rightPad(s, " ", output.MaxCharsWritten))
			// -1 to remove the newline, otherwise it's always bigger
			if (w - 1) > output.MaxCharsWritten {
				output.MaxCharsWritten = w - 1
			}
			output.Mu.Unlock()
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
func errorWorker(g *libgobuster.Gobuster, wg *sync.WaitGroup, output *outputType) {
	defer wg.Done()

	for e := range g.Errors() {
		if !g.Opts.Quiet && !g.Opts.NoError {
			output.Mu.Lock()
			g.LogError.Printf("[!] %v", e)
			output.Mu.Unlock()
		}
	}
}

// progressWorker outputs the progress every tick. It will stop once cancel() is called
// on the context
func progressWorker(ctx context.Context, g *libgobuster.Gobuster, wg *sync.WaitGroup, output *outputType) {
	defer wg.Done()

	tick := time.NewTicker(cliProgressUpdate)

	for {
		select {
		case <-tick.C:
			if !g.Opts.Quiet && !g.Opts.NoProgress {
				g.RequestsCountMutex.RLock()
				output.Mu.Lock()
				var charsWritten int
				if g.Opts.Wordlist == "-" {
					s := fmt.Sprintf("\rProgress: %d", g.RequestsIssued)
					s = rightPad(s, " ", output.MaxCharsWritten)
					charsWritten, _ = fmt.Fprint(os.Stderr, s)
					// only print status if we already read in the wordlist
				} else if g.RequestsExpected > 0 {
					s := fmt.Sprintf("\rProgress: %d / %d (%3.2f%%)", g.RequestsIssued, g.RequestsExpected, float32(g.RequestsIssued)*100.0/float32(g.RequestsExpected))
					s = rightPad(s, " ", output.MaxCharsWritten)
					charsWritten, _ = fmt.Fprint(os.Stderr, s)
				}
				if charsWritten > output.MaxCharsWritten {
					output.MaxCharsWritten = charsWritten
				}

				output.Mu.Unlock()
				g.RequestsCountMutex.RUnlock()
			}
		case <-ctx.Done():
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
		fmt.Println(ruler)
	}

	// our waitgroup for all goroutines
	// this ensures all goroutines are finished
	// when we call wg.Wait()
	var wg sync.WaitGroup

	outputMutex := new(sync.RWMutex)
	o := &outputType{
		Mu:              outputMutex,
		MaxCharsWritten: 0,
	}

	wg.Add(1)
	go resultWorker(gobuster, opts.OutputFilename, &wg, o)

	wg.Add(1)
	go errorWorker(gobuster, &wg, o)

	if !opts.Quiet && !opts.NoProgress {
		// if not quiet add a new workgroup entry and start the goroutine
		wg.Add(1)
		go progressWorker(ctxCancel, gobuster, &wg, o)
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
		// clear stderr progress
		fmt.Fprintf(os.Stderr, "\r%s\n", rightPad("", " ", o.MaxCharsWritten))
		fmt.Println(ruler)
		gobuster.LogInfo.Println("Finished")
		fmt.Println(ruler)
	}
	return nil
}
