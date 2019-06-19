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

func ruler() {
	fmt.Println("===============================================================")
}

func banner() {
	fmt.Printf("Gobuster v%s\n", libgobuster.VERSION)
	fmt.Println("by OJ Reeves (@TheColonial) & Christian Mehlmauer (@_FireFart_)")
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

	for r := range g.Results() {
		s, err := r.ToString(g)
		if err != nil {
			g.LogError.Fatal(err)
		}
		if s != "" {
			g.ClearProgress()
			s = strings.TrimSpace(s)
			fmt.Println(s)
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

	for e := range g.Errors() {
		if !g.Opts.Quiet {
			g.ClearProgress()
			g.LogError.Printf("[!] %v", e)
		}
	}
}

// progressWorker outputs the progress every tick. It will stop once cancel() is called
// on the context
func progressWorker(c context.Context, g *libgobuster.Gobuster, wg *sync.WaitGroup) {
	defer wg.Done()

	tick := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-tick.C:
			g.PrintProgress()
		case <-c.Done():
			return
		}
	}
}

func writeToFile(f *os.File, output string) error {
	_, err := f.WriteString(fmt.Sprintf("%s\n", output))
	if err != nil {
		return fmt.Errorf("[!] Unable to write to file %v", err)
	}
	return nil
}

// Gobuster is the main entry point for the CLI
func Gobuster(prevCtx context.Context, opts *libgobuster.Options, plugin libgobuster.GobusterPlugin) error {
	// Sanity checks
	if opts == nil {
		return fmt.Errorf("please provide valid options")
	}

	if plugin == nil {
		return fmt.Errorf("please provide a valid plugin")
	}

	ctx, cancel := context.WithCancel(prevCtx)
	defer cancel()

	gobuster, err := libgobuster.NewGobuster(ctx, opts, plugin)
	if err != nil {
		return err
	}

	if !opts.Quiet {
		ruler()
		banner()
		ruler()
		c, err := gobuster.GetConfigString()
		if err != nil {
			return fmt.Errorf("error on creating config string: %v", err)
		}
		fmt.Println(c)
		ruler()
		gobuster.LogInfo.Println("Starting gobuster")
		ruler()
	}

	// our waitgroup for all goroutines
	// this ensures all goroutines are finished
	// when we call wg.Wait()
	var wg sync.WaitGroup
	// 2 is the number of goroutines we spin up
	wg.Add(2)
	go errorWorker(gobuster, &wg)
	go resultWorker(gobuster, opts.OutputFilename, &wg)

	if !opts.Quiet && !opts.NoProgress {
		// if not quiet add a new workgroup entry and start the goroutine
		wg.Add(1)
		go progressWorker(ctx, gobuster, &wg)
	}

	err = gobuster.Start()

	// call cancel func so progressWorker will exit (the only goroutine in this
	// file using the context) and to free ressources
	cancel()
	// wait for all spun up goroutines to finish (all have to call wg.Done())
	wg.Wait()

	// Late error checking to finish all threads
	if err != nil {
		return err
	}

	if !opts.Quiet {
		gobuster.ClearProgress()
		ruler()
		gobuster.LogInfo.Println("Finished")
		ruler()
	}
	return nil
}
