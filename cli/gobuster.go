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

// resultWorker outputs the results as they come in. This needs to be a range and should not handle
// the context so the channel always has a receiver and libgobuster will not block.
func resultWorker(g *libgobuster.Gobuster, filename string, wg *sync.WaitGroup) {
	defer wg.Done()

	var f *os.File
	var err error
	if filename != "" {
		f, err = os.Create(filename)
		if err != nil {
			g.Logger.Fatalf("error on creating output file: %v", err)
		}
		defer f.Close()
	}

	for r := range g.Progress.ResultChan {
		s, err := r.ResultToString()
		if err != nil {
			g.Logger.Fatal(err)
		}
		if s != "" {
			s = strings.TrimSpace(s)
			if g.Opts.NoProgress || g.Opts.Quiet {
				_, _ = fmt.Printf("%s\n", s)
			} else {
				// only print the clear line when progress output is enabled
				_, _ = fmt.Printf("%s%s\n", TERMINAL_CLEAR_LINE, s)
			}
			if f != nil {
				err = writeToFile(f, s)
				if err != nil {
					g.Logger.Fatalf("error on writing output file: %v", err)
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
			g.Logger.Error(e.Error())
			g.Logger.Debugf("%#v", e)
		}
	}
}

// messageWorker outputs messages as they come in. This needs to be a range and should not handle
// the context so the channel always has a receiver and libgobuster will not block.
func messageWorker(g *libgobuster.Gobuster, wg *sync.WaitGroup) {
	defer wg.Done()

	for msg := range g.Progress.MessageChan {
		if !g.Opts.Quiet {
			switch msg.Level {
			case libgobuster.LevelDebug:
				g.Logger.Debug(msg.Message)
			case libgobuster.LevelError:
				g.Logger.Error(msg.Message)
			case libgobuster.LevelWarn:
				g.Logger.Warn(msg.Message)
			case libgobuster.LevelInfo:
				g.Logger.Info(msg.Message)
			default:
				panic(fmt.Sprintf("invalid level %d", msg.Level))
			}
		}
	}
}

func printProgress(g *libgobuster.Gobuster) {
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

// progressWorker outputs the progress every tick. It will stop once cancel() is called
// on the context
func progressWorker(ctx context.Context, g *libgobuster.Gobuster, wg *sync.WaitGroup) {
	defer wg.Done()

	tick := time.NewTicker(cliProgressUpdate)

	for {
		select {
		case <-tick.C:
			printProgress(g)
		case <-ctx.Done():
			// print the final progress so we end at 100%
			printProgress(g)
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
func Gobuster(ctx context.Context, opts *libgobuster.Options, plugin libgobuster.GobusterPlugin, log *libgobuster.Logger) error {
	// Sanity checks
	if opts == nil {
		return fmt.Errorf("please provide valid options")
	}

	if plugin == nil {
		return fmt.Errorf("please provide a valid plugin")
	}

	ctxCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	gobuster, err := libgobuster.NewGobuster(opts, plugin, log)
	if err != nil {
		return err
	}

	if !opts.Quiet {
		log.Println(ruler)
		log.Printf("Gobuster v%s\n", libgobuster.VERSION)
		log.Println("by OJ Reeves (@TheColonial) & Christian Mehlmauer (@firefart)")
		log.Println(ruler)
		c, err := gobuster.GetConfigString()
		if err != nil {
			return fmt.Errorf("error on creating config string: %w", err)
		}
		log.Println(c)
		log.Println(ruler)
		gobuster.Logger.Printf("Starting gobuster in %s mode", plugin.Name())
		if opts.WordlistOffset > 0 {
			gobuster.Logger.Printf("Skipping the first %d elements...", opts.WordlistOffset)
		}
		log.Println(ruler)
	}

	fi, err := os.Stdout.Stat()
	if err != nil {
		return err
	}
	// check if we are not in a terminal. If so, disable output
	if (fi.Mode() & os.ModeCharDevice) != os.ModeCharDevice {
		opts.NoProgress = true
	}

	// our waitgroup for all goroutines
	// this ensures all goroutines are finished
	// when we call wg.Wait()
	var wg sync.WaitGroup

	wg.Add(1)
	go resultWorker(gobuster, opts.OutputFilename, &wg)

	wg.Add(1)
	go errorWorker(gobuster, &wg)

	wg.Add(1)
	go messageWorker(gobuster, &wg)

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
		log.Println(ruler)
		gobuster.Logger.Println("Finished")
		log.Println(ruler)
	}
	return nil
}
