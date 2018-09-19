package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/OJ/gobuster/libgobuster"
)

func ruler() {
	fmt.Println("=====================================================")
}

func banner() {
	fmt.Printf("Gobuster v%s              OJ Reeves (@TheColonial)\n", libgobuster.VERSION)
}

func resultWorker(g *libgobuster.Gobuster, filename string, wg *sync.WaitGroup) {
	defer wg.Done()
	var f *os.File
	var err error
	if filename != "" {
		f, err = os.Create(filename)
		if err != nil {
			log.Fatalf("error on creating output file: %v", err)
		}
	}
	for r := range g.Results() {
		s, err := r.ToString(g)
		if err != nil {
			log.Fatal(err)
		}
		if s != "" {
			g.ClearProgress()
			s = strings.TrimSpace(s)
			fmt.Println(s)
			if f != nil {
				err = writeToFile(f, s)
				if err != nil {
					log.Fatalf("error on writing output file: %v", err)
				}
			}
		}
	}
}

func errorWorker(g *libgobuster.Gobuster, wg *sync.WaitGroup) {
	defer wg.Done()
	for e := range g.Errors() {
		if !g.Opts.Quiet {
			g.ClearProgress()
			log.Printf("[!] %v", e)
		}
	}
}

func progressWorker(c context.Context, g *libgobuster.Gobuster) {
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
		log.Fatalf("[!] %v", err)
	}

	if !opts.Quiet {
		fmt.Println("")
		ruler()
		banner()
		ruler()
		c, err := gobuster.GetConfigString()
		if err != nil {
			log.Fatalf("error on creating config string: %v", err)
		}
		fmt.Println(c)
		ruler()
		log.Println("Starting gobuster")
		ruler()
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go errorWorker(gobuster, &wg)
	go resultWorker(gobuster, opts.OutputFilename, &wg)

	if !opts.Quiet && !opts.NoProgress {
		go progressWorker(ctx, gobuster)
	}

	if err := gobuster.Start(); err != nil {
		log.Printf("[!] %v", err)
	} else {
		// call cancel func to free ressources and stop progressFunc
		cancel()
		// wait for all output funcs to finish
		wg.Wait()
	}

	if !opts.Quiet {
		gobuster.ClearProgress()
		ruler()
		log.Println("Finished")
		ruler()
	}
	return nil
}
