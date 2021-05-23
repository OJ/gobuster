package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/spf13/cobra"
)

// nolint:gochecknoglobals
var rootCmd = &cobra.Command{
	Use:          "gobuster",
	SilenceUsage: true,
}

// nolint:gochecknoglobals
var mainContext context.Context

// Execute is the main cobra method
func Execute() {
	var cancel context.CancelFunc
	mainContext, cancel = context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()
	go func() {
		select {
		case <-signalChan:
			// caught CTRL+C
			fmt.Println("\n[!] Keyboard interrupt detected, terminating.")
			cancel()
		case <-mainContext.Done():
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		// Leaving this in results in the same error appearing twice
		// Once before and once after the help output. Not sure if
		// this is going to be needed to output other errors that
		// aren't automatically outputted.
		// fmt.Println(err)
		os.Exit(1)
	}
}

func parseGlobalOptions() (*libgobuster.Options, error) {
	globalopts := libgobuster.NewOptions()

	threads, err := rootCmd.Flags().GetInt("threads")
	if err != nil {
		return nil, fmt.Errorf("invalid value for threads: %w", err)
	}

	if threads <= 0 {
		return nil, fmt.Errorf("threads must be bigger than 0")
	}
	globalopts.Threads = threads

	delay, err := rootCmd.Flags().GetDuration("delay")
	if err != nil {
		return nil, fmt.Errorf("invalid value for delay: %w", err)
	}

	if delay < 0 {
		return nil, fmt.Errorf("delay must be positive")
	}
	globalopts.Delay = delay

	globalopts.Wordlist, err = rootCmd.Flags().GetString("wordlist")
	if err != nil {
		return nil, fmt.Errorf("invalid value for wordlist: %w", err)
	}

	if globalopts.Wordlist == "-" {
		// STDIN
	} else if _, err2 := os.Stat(globalopts.Wordlist); os.IsNotExist(err2) {
		return nil, fmt.Errorf("wordlist file %q does not exist: %w", globalopts.Wordlist, err2)
	}

	globalopts.PatternFile, err = rootCmd.Flags().GetString("pattern")
	if err != nil {
		return nil, fmt.Errorf("invalid value for pattern: %w", err)
	}

	if globalopts.PatternFile != "" {
		if _, err = os.Stat(globalopts.PatternFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("pattern file %q does not exist: %w", globalopts.PatternFile, err)
		}
		patternFile, err := os.Open(globalopts.PatternFile)
		if err != nil {
			return nil, fmt.Errorf("could not open pattern file %q: %w", globalopts.PatternFile, err)
		}
		defer patternFile.Close()

		scanner := bufio.NewScanner(patternFile)
		for scanner.Scan() {
			globalopts.Patterns = append(globalopts.Patterns, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("could not read pattern file %q: %w", globalopts.PatternFile, err)
		}
	}

	globalopts.OutputFilename, err = rootCmd.Flags().GetString("output")
	if err != nil {
		return nil, fmt.Errorf("invalid value for output filename: %w", err)
	}

	globalopts.Verbose, err = rootCmd.Flags().GetBool("verbose")
	if err != nil {
		return nil, fmt.Errorf("invalid value for verbose: %w", err)
	}

	globalopts.Quiet, err = rootCmd.Flags().GetBool("quiet")
	if err != nil {
		return nil, fmt.Errorf("invalid value for quiet: %w", err)
	}

	globalopts.NoProgress, err = rootCmd.Flags().GetBool("no-progress")
	if err != nil {
		return nil, fmt.Errorf("invalid value for no-progress: %w", err)
	}

	globalopts.NoError, err = rootCmd.Flags().GetBool("no-error")
	if err != nil {
		return nil, fmt.Errorf("invalid value for no-error: %w", err)
	}

	return globalopts, nil
}

// This has to be called as part of the pre-run for sub commands. Including
// this in the init() function results in the built-in `help` command not
// working as intended. The required flags should only be marked as required
// on the global flags when one of the non-help commands is used.
func configureGlobalOptions() {
	if err := rootCmd.MarkPersistentFlagRequired("wordlist"); err != nil {
		log.Fatalf("error on marking flag as required: %v", err)
	}
}

// nolint:gochecknoinits
func init() {
	rootCmd.PersistentFlags().DurationP("delay", "", 0, "Time each thread waits between requests (e.g. 1500ms)")
	rootCmd.PersistentFlags().IntP("threads", "t", 10, "Number of concurrent threads")
	rootCmd.PersistentFlags().StringP("wordlist", "w", "", "Path to the wordlist")
	rootCmd.PersistentFlags().StringP("output", "o", "", "Output file to write results to (defaults to stdout)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output (errors)")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Don't print the banner and other noise")
	rootCmd.PersistentFlags().BoolP("no-progress", "z", false, "Don't display progress")
	rootCmd.PersistentFlags().Bool("no-error", false, "Don't display errors")
	rootCmd.PersistentFlags().StringP("pattern", "p", "", "File containing replacement patterns")
}
