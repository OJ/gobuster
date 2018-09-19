package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/OJ/gobuster/libgobuster"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "gobuster",
}

// Execute is the main cobra method
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
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

	wordlist, err := rootCmd.Flags().GetString("wordlist")
	if err != nil {
		return nil, fmt.Errorf("invalid value for wordlist: %v", err)
	}
	if wordlist == "-" {
		// STDIN
	} else if _, err2 := os.Stat(wordlist); os.IsNotExist(err2) {
		return nil, fmt.Errorf("wordlist file %q does not exist: %v", wordlist, err2)
	}
	globalopts.Wordlist = wordlist

	outputfilename, err := rootCmd.Flags().GetString("output")
	if err != nil {
		return nil, fmt.Errorf("invalid value for output filename: %v", err)
	}
	globalopts.OutputFilename = outputfilename

	verbose, err := rootCmd.Flags().GetBool("verbose")
	if err != nil {
		return nil, fmt.Errorf("invalid value for verbose: %v", err)
	}
	globalopts.Verbose = verbose

	quiet, err := rootCmd.Flags().GetBool("quiet")
	if err != nil {
		return nil, fmt.Errorf("invalid value for quiet: %v", err)
	}
	globalopts.Quiet = quiet

	noprogress, err := rootCmd.Flags().GetBool("noprogress")
	if err != nil {
		return nil, fmt.Errorf("invalid value for noprogress: %v", err)
	}
	globalopts.NoProgress = noprogress

	return globalopts, nil
}

func init() {
	rootCmd.PersistentFlags().IntP("threads", "t", 10, "Number of concurrent threads")
	rootCmd.PersistentFlags().StringP("wordlist", "w", "", "Path to the wordlist")
	rootCmd.PersistentFlags().StringP("output", "o", "", "Output file to write results to (defaults to stdout)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output (errors)")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Don't print the banner and other noise")
	rootCmd.PersistentFlags().BoolP("noprogress", "", false, "Don't display progress")
	if err := rootCmd.MarkPersistentFlagRequired("wordlist"); err != nil {
		log.Fatalf("error on marking flag as required: %v", err)
	}
}
