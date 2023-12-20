package gobustertftp

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/OJ/gobuster/v3/libgobuster"

	"github.com/pin/tftp/v3"
)

// GobusterTFTP is the main type to implement the interface
type GobusterTFTP struct {
	globalopts *libgobuster.Options
	options    *OptionsTFTP
}

// New creates a new initialized NewGobusterTFTP
func New(globalopts *libgobuster.Options, opts *OptionsTFTP) (*GobusterTFTP, error) {
	if globalopts == nil {
		return nil, fmt.Errorf("please provide valid global options")
	}

	if opts == nil {
		return nil, fmt.Errorf("please provide valid plugin options")
	}

	g := GobusterTFTP{
		options:    opts,
		globalopts: globalopts,
	}
	return &g, nil
}

// Name should return the name of the plugin
func (d *GobusterTFTP) Name() string {
	return "TFTP enumeration"
}

// PreRun is the pre run implementation of gobustertftp
func (d *GobusterTFTP) PreRun(_ context.Context, _ *libgobuster.Progress) error {
	_, err := tftp.NewClient(d.options.Server)
	if err != nil {
		return err
	}
	return nil
}

// ProcessWord is the process implementation of gobustertftp
func (d *GobusterTFTP) ProcessWord(_ context.Context, word string, progress *libgobuster.Progress) error {
	// add some debug output
	if d.globalopts.Debug {
		progress.MessageChan <- libgobuster.Message{
			Level:   libgobuster.LevelDebug,
			Message: fmt.Sprintf("trying word %s", word),
		}
	}

	c, err := tftp.NewClient(d.options.Server)
	if err != nil {
		return err
	}
	c.SetTimeout(d.options.Timeout)
	wt, err := c.Receive(word, "octet")
	if err != nil {
		// file not found
		return nil
	}
	result := Result{
		Filename: word,
	}
	if n, ok := wt.(tftp.IncomingTransfer).Size(); ok {
		result.Size = n
	}
	progress.ResultChan <- result
	return nil
}

func (d *GobusterTFTP) AdditionalWords(_ string) []string {
	return []string{}
}

// GetConfigString returns the string representation of the current config
func (d *GobusterTFTP) GetConfigString() (string, error) {
	var buffer bytes.Buffer
	bw := bufio.NewWriter(&buffer)
	tw := tabwriter.NewWriter(bw, 0, 5, 3, ' ', 0)
	o := d.options

	if _, err := fmt.Fprintf(tw, "[+] Server:\t%s\n", o.Server); err != nil {
		return "", err
	}

	if _, err := fmt.Fprintf(tw, "[+] Threads:\t%d\n", d.globalopts.Threads); err != nil {
		return "", err
	}

	if d.globalopts.Delay > 0 {
		if _, err := fmt.Fprintf(tw, "[+] Delay:\t%s\n", d.globalopts.Delay); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(tw, "[+] Timeout:\t%s\n", o.Timeout.String()); err != nil {
		return "", err
	}

	wordlist := "stdin (pipe)"
	if d.globalopts.Wordlist != "-" {
		wordlist = d.globalopts.Wordlist
	}
	if _, err := fmt.Fprintf(tw, "[+] Wordlist:\t%s\n", wordlist); err != nil {
		return "", err
	}

	if d.globalopts.PatternFile != "" {
		if _, err := fmt.Fprintf(tw, "[+] Patterns:\t%s (%d entries)\n", d.globalopts.PatternFile, len(d.globalopts.Patterns)); err != nil {
			return "", err
		}
	}

	if err := tw.Flush(); err != nil {
		return "", fmt.Errorf("error on tostring: %w", err)
	}

	if err := bw.Flush(); err != nil {
		return "", fmt.Errorf("error on tostring: %w", err)
	}

	return strings.TrimSpace(buffer.String()), nil
}
