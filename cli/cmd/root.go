package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/OJ/gobuster/v3/libgobuster"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "gobuster",
	SilenceUsage: true,
}

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
		//fmt.Println(err)
		os.Exit(1)
	}
}

func parseGlobalOptions() (*libgobuster.Options, error) {
	globalopts := libgobuster.NewOptions()

	threads, err := rootCmd.Flags().GetInt("threads")
	if err != nil {
		return nil, fmt.Errorf("invalid value for threads: %v", err)
	}

	if threads <= 0 {
		return nil, fmt.Errorf("threads must be bigger than 0")
	}
	globalopts.Threads = threads

	delay, err := rootCmd.Flags().GetDuration("delay")
	if err != nil {
		return nil, fmt.Errorf("invalid value for delay: %v", err)
	}

	if delay < 0 {
		return nil, fmt.Errorf("delay must be positive")
	}
	globalopts.Delay = delay

	globalopts.Wordlist, err = rootCmd.Flags().GetString("wordlist")
	if err != nil {
		return nil, fmt.Errorf("invalid value for wordlist: %v", err)
	}

	if globalopts.Wordlist == "-" {
		// STDIN
	} else if _, err2 := os.Stat(globalopts.Wordlist); os.IsNotExist(err2) {
		return nil, fmt.Errorf("wordlist file %q does not exist: %v", globalopts.Wordlist, err2)
	}

	globalopts.OutputFilename, err = rootCmd.Flags().GetString("output")
	if err != nil {
		return nil, fmt.Errorf("invalid value for output filename: %v", err)
	}

	globalopts.Verbose, err = rootCmd.Flags().GetBool("verbose")
	if err != nil {
		return nil, fmt.Errorf("invalid value for verbose: %v", err)
	}

	globalopts.Quiet, err = rootCmd.Flags().GetBool("quiet")
	if err != nil {
		return nil, fmt.Errorf("invalid value for quiet: %v", err)
	}

	globalopts.NoProgress, err = rootCmd.Flags().GetBool("noprogress")
	if err != nil {
		return nil, fmt.Errorf("invalid value for noprogress: %v", err)
	}

	return globalopts, nil
}

// This has to be called as part of the pre-run for sub commands. Including
// this in the init() function results in the built-in `help` command not
// working as intended. The required flags should only be marked as required
// on the global flags when one of the non-help commands is utilised.
func configureGlobalOptions() {
	if err := rootCmd.MarkPersistentFlagRequired("wordlist"); err != nil {
		log.Fatalf("error on marking flag as required: %v", err)
	}
}

func init() {
	rootCmd.PersistentFlags().DurationP("delay", "", 0, "Time each thread waits between requests (e.g. 1500ms)")
	rootCmd.PersistentFlags().IntP("threads", "t", 10, "Number of concurrent threads")
	rootCmd.PersistentFlags().StringP("wordlist", "w", "", "Path to the wordlist")
	rootCmd.PersistentFlags().StringP("output", "o", "", "Output file to write results to (defaults to stdout)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output (errors)")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Don't print the banner and other noise")
	rootCmd.PersistentFlags().BoolP("noprogress", "z", false, "Don't display progress")
}
