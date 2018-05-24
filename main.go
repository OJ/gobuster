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
	"syscall"
	"time"

	"github.com/OJ/gobuster/gobusterdir"
	"github.com/OJ/gobuster/gobusterdns"
	"github.com/OJ/gobuster/libgobuster"
	"golang.org/x/crypto/ssh/terminal"
)

func ruler() {
	fmt.Println("=====================================================")
}

func banner() {
	fmt.Println("")
	fmt.Printf("Gobuster v%s              OJ Reeves (@TheColonial)\n", libgobuster.VERSION)
}

func resultWorker(g *libgobuster.Gobuster, filename string) {
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

func errorWorker(g *libgobuster.Gobuster) {
	for e := range g.Errors() {
		g.ClearProgress()
		log.Printf("[!] %v", e)
	}
}

func progressWorker(g *libgobuster.Gobuster) {
	tick := time.NewTicker(1 * time.Second)

	for range tick.C {
		g.PrintProgress()
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
	flag.BoolVar(&o.Quiet, "q", false, "Don't print the banner and other noise")
	flag.BoolVar(&o.Expanded, "e", false, "Expanded mode, print full URLs")
	flag.BoolVar(&o.NoStatus, "n", false, "Don't print status codes")
	flag.BoolVar(&o.IncludeLength, "l", false, "Include the length of the body in the output (dir mode only)")
	flag.BoolVar(&o.UseSlash, "f", false, "Append a forward-slash to each directory request (dir mode only)")
	flag.BoolVar(&o.WildcardForced, "fw", false, "Force continued operation when wildcard found")
	flag.BoolVar(&o.InsecureSSL, "k", false, "Skip SSL certificate verification")

	flag.Parse()

	// Prompt for PW if not provided
	if o.Username != "" && o.Password == "" {
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

	var funcSetup func(*libgobuster.Gobuster) error
	var funcProcessor func(*libgobuster.Gobuster, string) ([]libgobuster.Result, error)
	var funcResToString func(*libgobuster.Gobuster, *libgobuster.Result) (*string, error)

	switch o.Mode {
	case libgobuster.ModeDir:
		funcSetup = gobusterdir.SetupDir
		funcProcessor = gobusterdir.ProcessDirEntry
		funcResToString = gobusterdir.DirResultToString
	case libgobuster.ModeDNS:
		funcSetup = gobusterdns.SetupDNS
		funcProcessor = gobusterdns.ProcessDNSEntry
		funcResToString = gobusterdns.DNSResultToString
	}

	gobuster, err := libgobuster.NewGobuster(ctx, o, funcSetup, funcProcessor, funcResToString)
	if err != nil {
		log.Fatalf("[!] %v", err)
	}

	if !o.Quiet {
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

	go errorWorker(gobuster)
	go resultWorker(gobuster, outputFilename)

	if !o.Quiet {
		go progressWorker(gobuster)
	}

	if err := gobuster.Start(); err != nil {
		log.Fatalf("[!] %v", err)
	}

	if !o.Quiet {
		gobuster.ClearProgress()
		ruler()
		log.Println("Finished")
		ruler()
	}
}
