package main

//----------------------------------------------------
// Gobuster -- by OJ Reeves
//
// A crap attempt at building something that resembles
// dirbuster or dirb using Go. The goal was to build
// a tool that would help learn Go and to actually do
// something useful. The idea of having this compile
// to native code is also appealing.
//
// Run: gobuster -h
//
// Please see THANKS file for contributors.
// Please see LICENSE file for license details.
//
//----------------------------------------------------

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Margular/gobuster/gobusterdir"
	"github.com/Margular/gobuster/gobusterdns"
	"github.com/Margular/gobuster/libgobuster"
	"golang.org/x/crypto/ssh/terminal"
)

func ruler() {
	fmt.Println("=====================================================")
}

func banner() {
	fmt.Printf("Gobuster v%s              OJ Reeves (@TheColonial)\n", libgobuster.VERSION)
}

func resultWorker(g *libgobuster.Gobuster, f *os.File, wg *sync.WaitGroup) {
	defer wg.Done()

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

func main() {
	var outputFilename string
	o := libgobuster.NewOptions()
	flag.IntVar(&o.Threads, "t", 10, "Number of concurrent threads")
	flag.StringVar(&o.Mode, "m", "dir", "Directory/File mode (dir) or DNS mode (dns)")
	flag.StringVar(&o.Wordlist, "w", "", "Path to the wordlist")
	flag.StringVar(&o.StatusCodes, "s", "200,204,301,302,307,403", "Positive status codes (dir mode only)")
	flag.StringVar(&outputFilename, "o", "", "Output file to write results to (defaults to stdout)")
	flag.StringVar(&o.URL, "u", "", "The target URL or Domain")
	flag.StringVar(&o.URLFile, "file", "", "The path to store urls")
	flag.StringVar(&o.Cookies, "c", "", "Cookies to use for the requests (dir mode only)")
	flag.StringVar(&o.Username, "U", "", "Username for Basic Auth (dir mode only)")
	flag.StringVar(&o.Password, "P", "", "Password for Basic Auth (dir mode only)")
	flag.StringVar(&o.Extensions, "x", "", "File extension(s) to search for (dir mode only)")
	flag.StringVar(&o.UserAgent, "a", "", "Set the User-Agent string (dir mode only)")
	flag.StringVar(&o.Proxy, "p", "", "Proxy to use for requests [http(s)://host:port] (dir mode only)")
	flag.DurationVar(&o.Timeout, "to", 10*time.Second, "HTTP Timeout in seconds (dir mode only)")
	flag.BoolVar(&o.Verbose, "v", false, "Verbose output (errors)")
	flag.BoolVar(&o.ShowIPs, "i", false, "Show IP addresses (dns mode only)")
	flag.BoolVar(&o.ShowCNAME, "cn", false, "Show CNAME records (dns mode only, cannot be used with '-i' option)")
	flag.BoolVar(&o.FollowRedirect, "r", false, "Follow redirects")
	flag.BoolVar(&o.Recursive, "R", false, "Recursive scan")
	flag.BoolVar(&o.Quiet, "q", false, "Don't print the banner and other noise")
	flag.BoolVar(&o.Expanded, "e", false, "Expanded mode, print full URLs")
	flag.BoolVar(&o.NoStatus, "n", false, "Don't print status codes")
	flag.BoolVar(&o.IncludeLength, "l", false, "Include the length of the body in the output (dir mode only)")
	flag.BoolVar(&o.UseSlash, "f", false, "Append a forward-slash to each directory request (dir mode only)")
	flag.BoolVar(&o.WildcardForced, "fw", false, "Force continued operation when wildcard found")
	flag.BoolVar(&o.InsecureSSL, "k", false, "Skip SSL certificate verification")
	flag.BoolVar(&o.NoProgress, "np", false, "Don't display progress")

	flag.Parse()

	var urlPool []string

	if o.URL == "" {
		urlPool = []string{}
	} else {
		urlPool = []string{o.URL}
	}

	urlPool = append(urlPool, o.ReadUrls()...)

	times := 0

	var f *os.File
	var err error

	if outputFilename != "" {
		f, err = os.OpenFile(outputFilename, os.O_APPEND, 0666)
		if err != nil {
			f, err = os.Create(outputFilename)
			if err != nil {
				log.Fatalf("error on creating output file: %v", err)
			}
		}
	}

	for _, url := range urlPool {
		times += 1

		o.URL = url

		// Prompt for PW if not provided
		if o.Username != "" && o.Password == "" && times <= 1 {
			fmt.Printf("[?] Auth Password: ")
			passBytes, err := terminal.ReadPassword(int(syscall.Stdin))
			// print a newline to simulate the newline that was entered
			// this means that formatting/printing after doesn't look bad.
			fmt.Println("")
			if err != nil {
				log.Fatal("[!] Auth username given but reading of password failed")
			}
			o.Password = string(passBytes)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var plugin libgobuster.GobusterPlugin
		switch o.Mode {
		case libgobuster.ModeDir:
			plugin = gobusterdir.GobusterDir{}
		case libgobuster.ModeDNS:
			plugin = gobusterdns.GobusterDNS{}
		}

		gobuster, err := libgobuster.NewGobuster(ctx, o, plugin)
		if err != nil {
			log.Fatalf("[!] %v", err)
		}

		if !o.Quiet && times <= 1 {
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

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)
		go func() {
			for range signalChan {
				// caught CTRL+C
				if !gobuster.Opts.Quiet {
					fmt.Println("\n[!] Keyboard interrupt detected, terminating.")
				}
				cancel()
			}
		}()

		var wg sync.WaitGroup
		wg.Add(2)
		go errorWorker(gobuster, &wg)
		go resultWorker(gobuster, f, &wg)

		if !o.Quiet && !o.NoProgress {
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

		if !o.Quiet {
			gobuster.ClearProgress()
		}
	}

	ruler()
	log.Println("Finished")
	ruler()
}