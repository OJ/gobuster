package dir

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	internalcli "github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobusterdir"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	cmd := cli.Command{
		Name:   "dir",
		Usage:  "Uses directory/file enumeration mode",
		Action: run,
		Flags:  getFlags(),
	}
	return &cmd
}

func getFlags() []cli.Flag {
	var flags []cli.Flag
	flags = append(flags, internalcli.CommonHTTPOptions()...)
	flags = append(flags, internalcli.GlobalOptions()...)
	flags = append(flags, []cli.Flag{
		&cli.StringFlag{Name: "status-codes", Aliases: []string{"s"}, Usage: "Positive status codes (will be overwritten with status-codes-blacklist if set). Can also handle ranges like 200,300-400,404"},
		&cli.StringFlag{Name: "status-codes-blacklist", Aliases: []string{"b"}, Usage: "Negative status codes (will override status-codes if set). Can also handle ranges like 200,300-400,404.", Value: "404"},
		&cli.StringFlag{Name: "extensions", Aliases: []string{"x"}, Usage: "File extension(s) to search for"},
		&cli.StringFlag{Name: "extensions-file", Aliases: []string{"X"}, Usage: "Read file extension(s) to search from the file"},
		&cli.BoolFlag{Name: "expanded", Aliases: []string{"e"}, Value: false, Usage: "Expanded mode, print full URLs"},
		&cli.BoolFlag{Name: "no-status", Aliases: []string{"n"}, Value: false, Usage: "Don't print status codes"},
		&cli.BoolFlag{Name: "hide-length", Aliases: []string{"hl"}, Value: false, Usage: "Hide the length of the body in the output"},
		&cli.BoolFlag{Name: "add-slash", Aliases: []string{"f"}, Value: false, Usage: "Append / to each request"},
		&cli.BoolFlag{Name: "discover-backup", Aliases: []string{"db"}, Value: false, Usage: "Upon finding a file search for backup files by appending multiple backup extensions"},
		&cli.StringFlag{Name: "exclude-length", Aliases: []string{"xl"}, Usage: "exclude the following content lengths (completely ignores the status). You can separate multiple lengths by comma and it also supports ranges like 203-206"},
		&cli.BoolFlag{Name: "force", Value: false, Usage: "Continue even if the prechecks fail. Please only use this if you know what you are doing, it can lead to unexpected results."},
		&cli.StringFlag{Name: "list", Aliases: []string{"l"}, Usage: "File containing target URLs"},
		&cli.BoolFlag{Name: "autocalibrate", Aliases: []string{"ac"}, Value: false, Usage: "Automatically calibrate wildcard responses"},
	}...)
	return flags
}

func run(c *cli.Context) error {
	urlInput := c.String("url")
	listInput := c.String("list")

	if urlInput == "" && listInput == "" {
		return errors.New("either the url flag or the list flag is required")
	}

	if urlInput != "" && listInput != "" {
		return errors.New("cannot use both the url and list flags")
	}

	var urls []string
	if urlInput != "" {
		urls = append(urls, urlInput)
	} else {
		file, err := os.Open(listInput)
		if err != nil {
			return fmt.Errorf("failed to open list file: %w", err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				urls = append(urls, line)
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed to read list file: %w", err)
		}
	}

	outputFilename := c.String("output")
	if outputFilename != "" && !c.Bool("append") {
		f, err := os.Create(outputFilename)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		f.Close()
	}

	for _, u := range urls {
		err := runForTarget(c, u, len(urls) > 1)
		if err != nil {
			// for multiple targets, we might want to continue or stop.
			// for now let's stop on the first error to be safe, or just log it?
			// if it's a connection error it might be worth continuing.
			if len(urls) > 1 {
				fmt.Printf("[-] Error on %s: %v\n", u, err)
				continue
			}
			return err
		}
	}
	return nil
}

func runForTarget(c *cli.Context, targetURL string, isMultiTarget bool) error {
	pluginOpts := gobusterdir.NewOptions()

	// Parse common options but we need to handle the URL separately
	// because ParseCommonHTTPOptions expects it in the context.
	// We'll set a dummy URL in the context if it's empty to pass validation,
	// then overwrite it.
	httpOptions, err := internalcli.ParseCommonHTTPOptions(c)
	if err != nil && targetURL == "" {
		return err
	}
	pluginOpts.HTTPOptions = httpOptions

	// Custom URL parsing (re-implementing the logic from ParseCommonHTTPOptions)
	if !strings.HasPrefix(targetURL, "http") {
		targetURL = fmt.Sprintf("http://%s", targetURL)
	}
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid url %q: %w", targetURL, err)
	}
	pluginOpts.URL = parsedURL

	pluginOpts.Extensions = c.String("extensions")
	ret, err := libgobuster.ParseExtensions(pluginOpts.Extensions)
	if err != nil {
		return fmt.Errorf("invalid value for extensions: %w", err)
	}
	pluginOpts.ExtensionsParsed = ret

	pluginOpts.ExtensionsFile = c.String("extensions-file")
	if pluginOpts.ExtensionsFile != "" {
		extensions, err := libgobuster.ParseExtensionsFile(pluginOpts.ExtensionsFile)
		if err != nil {
			return fmt.Errorf("invalid value for extensions file: %w", err)
		}
		pluginOpts.ExtensionsParsed.AddRange(extensions)
	}

	pluginOpts.StatusCodes = c.String("status-codes")
	ret2, err := libgobuster.ParseCommaSeparatedInt(pluginOpts.StatusCodes)
	if err != nil {
		return fmt.Errorf("invalid value for status-codes: %w", err)
	}
	pluginOpts.StatusCodesParsed = ret2

	pluginOpts.StatusCodesBlacklist = c.String("status-codes-blacklist")
	ret3, err := libgobuster.ParseCommaSeparatedInt(pluginOpts.StatusCodesBlacklist)
	if err != nil {
		return fmt.Errorf("invalid value for status-codes-blacklist: %w", err)
	}
	pluginOpts.StatusCodesBlacklistParsed = ret3

	if pluginOpts.StatusCodes != "" && pluginOpts.StatusCodesBlacklist != "" {
		return fmt.Errorf("status-codes (%q) and status-codes-blacklist (%q) are both set - please set only one. status-codes-blacklist is set by default so you might want to disable it by supplying an empty string",
			pluginOpts.StatusCodes, pluginOpts.StatusCodesBlacklist)
	}

	if pluginOpts.StatusCodes == "" && pluginOpts.StatusCodesBlacklist == "" {
		return errors.New("status-codes and status-codes-blacklist are both not set, please set one")
	}

	pluginOpts.UseSlash = c.Bool("add-slash")
	pluginOpts.Expanded = c.Bool("expanded")
	pluginOpts.NoStatus = c.Bool("no-status")
	pluginOpts.HideLength = c.Bool("hide-length")
	pluginOpts.DiscoverBackup = c.Bool("discover-backup")
	pluginOpts.Force = c.Bool("force")
	pluginOpts.ExcludeLength = c.String("exclude-length")
	ret4, err := libgobuster.ParseCommaSeparatedInt(pluginOpts.ExcludeLength)
	if err != nil {
		return fmt.Errorf("invalid value for exclude-length: %w", err)
	}
	pluginOpts.ExcludeLengthParsed = ret4
	pluginOpts.AutoCalibrate = c.Bool("autocalibrate")

	globalOpts, err := internalcli.ParseGlobalOptions(c)
	if err != nil {
		return err
	}

	// if an output file is specified, we always want to append
	// as we handle the initial truncation once at the start of the run function
	if globalOpts.OutputFilename != "" {
		globalOpts.Append = true
	}

	// if we have multiple targets, we want to show the full URL
	// so it's easy to identify which target a result belongs to.
	if isMultiTarget {
		pluginOpts.Expanded = true
	} else {
		pluginOpts.Expanded = c.Bool("expanded")
	}

	log := libgobuster.NewLogger(globalOpts.Debug)

	plugin, err := gobusterdir.New(&globalOpts, pluginOpts, log)
	if err != nil {
		return fmt.Errorf("error on creating gobusterdir: %w", err)
	}

	if err := internalcli.Gobuster(c.Context, &globalOpts, plugin, log); err != nil {
		var wErr *gobusterdir.WildcardError
		if errors.As(err, &wErr) {
			return fmt.Errorf("%w. To continue please exclude the status code or the length", wErr)
		}
		log.Debugf("%#v", err)
		return fmt.Errorf("error on running gobuster on %s: %w", pluginOpts.URL, err)
	}
	return nil
}
