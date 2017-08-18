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
        "fmt"
        "flag"
        "os"
        "strconv"
        "strings"
        "regexp"
        "net/http"
        "net/url"
        "crypto/tls"
        "syscall"
        "golang.org/x/crypto/ssh/terminal"
        "github.com/OJ/gobuster/libgobuster"
)

// Parse all the command line options into a settings
// instance for future use.
func ParseCmdLine() *libgobuster.State {
  var extensions string
  var codes string
  var proxy string
  valid := true

  s := libgobuster.State{
    StatusCodes: libgobuster.IntSet{Set: map[int]bool{}},
    WildcardIps: libgobuster.StringSet{Set: map[string]bool{}},
    IsWildcard:  false,
    StdIn:       false,
  }

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

  flag.Parse()

  libgobuster.Banner(&s)

  switch strings.ToLower(s.Mode) {
  case "dir":
    s.Printer = libgobuster.PrintDirResult
    s.Processor = libgobuster.ProcessDirEntry
    s.Setup = libgobuster.SetupDir
  case "dns":
    s.Printer = libgobuster.PrintDnsResult
    s.Processor = libgobuster.ProcessDnsEntry
    s.Setup = libgobuster.SetupDns
  default:
    fmt.Println("[!] Mode (-m): Invalid value:", s.Mode)
    valid = false
  }

  if s.Threads < 0 {
    fmt.Println("[!] Threads (-t): Invalid value:", s.Threads)
    valid = false
  }

  stdin, err := os.Stdin.Stat()
  if err != nil {
    fmt.Println("[!] Unable to stat stdin, falling back to wordlist file.")
  } else if (stdin.Mode()&os.ModeCharDevice) == 0 && stdin.Size() > 0 {
    s.StdIn = true
  }

  if !s.StdIn {
    if s.Wordlist == "" {
      fmt.Println("[!] WordList (-w): Must be specified")
      valid = false
    } else if _, err := os.Stat(s.Wordlist); os.IsNotExist(err) {
      fmt.Println("[!] Wordlist (-w): File does not exist:", s.Wordlist)
      valid = false
    }
  } else if s.Wordlist != "" {
    fmt.Println("[!] Wordlist (-w) specified with pipe from stdin. Can't have both!")
    valid = false
  }

  if s.Url == "" {
    fmt.Println("[!] Url/Domain (-u): Must be specified")
    valid = false
  }

  if s.Mode == "dir" {
    if strings.HasSuffix(s.Url, "/") == false {
      s.Url = s.Url + "/"
    }

    if strings.HasPrefix(s.Url, "http") == false {
      // check to see if a port was specified
      re := regexp.MustCompile(`^[^/]+:(\d+)`)
      match := re.FindStringSubmatch(s.Url)

      if len(match) < 2 {
        // no port, default to http on 80
        s.Url = "http://" + s.Url
      } else {
        port, err := strconv.Atoi(match[1])
        if err != nil || (port != 80 && port != 443) {
          fmt.Println("[!] Url/Domain (-u): Scheme not specified.")
          valid = false
        } else if port == 80 {
          s.Url = "http://" + s.Url
        } else {
          s.Url = "https://" + s.Url
        }
      }
    }

    // extensions are comma separated
    if extensions != "" {
      s.Extensions = strings.Split(extensions, ",")
      for i := range s.Extensions {
        if s.Extensions[i][0] != '.' {
          s.Extensions[i] = "." + s.Extensions[i]
        }
      }
    }

    // status codes are comma separated
    if codes != "" {
      for _, c := range strings.Split(codes, ",") {
        i, err := strconv.Atoi(c)
        if err != nil {
          fmt.Println("[!] Invalid status code given: ", c)
          valid = false
        } else {
          s.StatusCodes.Add(i)
        }
      }
    }

    // prompt for password if needed
    if valid && s.Username != "" && s.Password == "" {
      fmt.Printf("[?] Auth Password: ")
      passBytes, err := terminal.ReadPassword(int(syscall.Stdin))

      // print a newline to simulate the newline that was entered
      // this means that formatting/printing after doesn't look bad.
      fmt.Println("")

      if err == nil {
        s.Password = string(passBytes)
      } else {
        fmt.Println("[!] Auth username given but reading of password failed")
        valid = false
      }
    }

    if valid {
      var proxyUrlFunc func(*http.Request) (*url.URL, error)
      proxyUrlFunc = http.ProxyFromEnvironment

      if proxy != "" {
        proxyUrl, err := url.Parse(proxy)
        if err != nil {
          panic("[!] Proxy URL is invalid")
        }
        s.ProxyUrl = proxyUrl
        proxyUrlFunc = http.ProxyURL(s.ProxyUrl)
      }

      s.Client = &http.Client{
        Transport: &libgobuster.RedirectHandler{
          State: &s,
          Transport: &http.Transport{
            Proxy: proxyUrlFunc,
            TLSClientConfig: &tls.Config{
              InsecureSkipVerify: s.InsecureSSL,
            },
          },
        }}

      code, _ := libgobuster.GoGet(&s, s.Url, "", s.Cookies)
      if code == nil {
        fmt.Println("[-] Unable to connect:", s.Url)
        valid = false
      }
    } else {
      libgobuster.Ruler(&s)
    }
  }

  if valid {
    return &s
  }

  return nil
}

func main() {
 state := ParseCmdLine()
 if state != nil {
   libgobuster.Process(state)
 }
}
