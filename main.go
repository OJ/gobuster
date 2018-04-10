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
	"flag"
	"fmt"

	"github.com/OJ/gobuster/libgobuster"
)

// Parse all the command line options into a settings
// instance for future use.
func ParseCmdLine() *libgobuster.State {
	var extensions string
	var codes string
	var proxy string

	s := libgobuster.InitState()

	// Set up the variables we're interested in parsing.
	flag.IntVar(&s.Threads, "t", 10, "Number of concurrent threads")
	flag.StringVar(&s.Mode, "m", "dir", "Directory/File mode (dir) or DNS mode (dns)")
	flag.StringVar(&s.Wordlist, "w", "", "Path to the wordlist")
	flag.StringVar(&codes, "s", "200,204,301,302,307", "Positive status codes (dir mode only)")
	flag.StringVar(&s.OutputFileName, "o", "", "Output file to write results to (defaults to stdout)")
	flag.StringVar(&s.Url, "u", "", "The target URL or Domain")
	flag.StringVar(&s.Cookies, "c", "", "Cookies to use for the requests (dir mode only)")
	flag.StringVar(&s.Username, "U", "", "Username for Basic Auth (dir mode only)")
	flag.StringVar(&s.Password, "P", "", "Password for Basic Auth (dir mode only)")
	flag.StringVar(&extensions, "x", "", "File extension(s) to search for (dir mode only)")
	flag.StringVar(&s.UserAgent, "a", "", "Set the User-Agent string (dir mode only)")
	flag.StringVar(&proxy, "p", "", "Proxy to use for requests [http(s)://host:port] (dir mode only)")
	flag.BoolVar(&s.Verbose, "v", false, "Verbose output (errors)")
	flag.BoolVar(&s.ShowIPs, "i", false, "Show IP addresses (dns mode only)")
	flag.BoolVar(&s.ShowCNAME, "cn", false, "Show CNAME records (dns mode only, cannot be used with '-i' option)")
	flag.BoolVar(&s.FollowRedirect, "r", false, "Follow redirects")
	flag.BoolVar(&s.Quiet, "q", false, "Don't print the banner and other noise")
	flag.BoolVar(&s.Expanded, "e", false, "Expanded mode, print full URLs")
	flag.BoolVar(&s.NoStatus, "n", false, "Don't print status codes")
	flag.BoolVar(&s.IncludeLength, "l", false, "Include the length of the body in the output (dir mode only)")
	flag.BoolVar(&s.UseSlash, "f", false, "Append a forward-slash to each directory request (dir mode only)")
	flag.BoolVar(&s.WildcardForced, "fw", false, "Force continued operation when wildcard found")
	flag.BoolVar(&s.InsecureSSL, "k", false, "Skip SSL certificate verification")
	flag.BoolVar(&s.ExpandRanges, "xr", false, "Expand words in wordlist that contain range patterns")

	flag.Parse()

	libgobuster.Banner(&s)

	if err := libgobuster.ValidateState(&s, extensions, codes, proxy); err.ErrorOrNil() != nil {
		fmt.Printf("%s\n", err.Error())
		return nil
	} else {
		libgobuster.Ruler(&s)
		return &s
	}
}

func main() {
	state := ParseCmdLine()
	if state != nil {
		libgobuster.Process(state)
	}
}
