package fuzz

import (
	"fmt"
	"strings"

	internalcli "github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobusterfuzz"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	cmd := cli.Command{
		Name:   "fuzz",
		Usage:  fmt.Sprintf("Uses fuzzing mode. Replaces the keyword %s in the URL, Headers and the request body", gobusterfuzz.FuzzKeyword),
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
		&cli.StringFlag{Name: "exclude-statuscodes", Aliases: []string{"b"}, Usage: "Excluded status codes. Can also handle ranges like 200,300-400,404."},
		&cli.StringFlag{Name: "exclude-length", Aliases: []string{"xl"}, Usage: "exclude the following content lengths (completely ignores the status). You can separate multiple lengths by comma and it also supports ranges like 203-206"},
		&cli.StringFlag{Name: "body", Aliases: []string{"B"}, Usage: "Request body"},
	}...)

	return flags
}

func run(c *cli.Context) error {
	pluginOpts := gobusterfuzz.NewOptions()

	httpOptions, err := internalcli.ParseCommonHTTPOptions(c)
	if err != nil {
		return err
	}
	pluginOpts.HTTPOptions = httpOptions

	pluginOpts.ExcludedStatusCodes = c.String("exclude-statuscodes")
	ret, err := libgobuster.ParseCommaSeparatedInt(pluginOpts.ExcludedStatusCodes)
	if err != nil {
		return fmt.Errorf("invalid value for excludestatuscodes: %w", err)
	}
	pluginOpts.ExcludedStatusCodesParsed = ret

	pluginOpts.ExcludeLength = c.String("exclude-length")
	ret2, err := libgobuster.ParseCommaSeparatedInt(pluginOpts.ExcludeLength)
	if err != nil {
		return fmt.Errorf("invalid value for exclude-length: %w", err)
	}
	pluginOpts.ExcludeLengthParsed = ret2

	pluginOpts.RequestBody = c.String("body")

	globalOpts, err := internalcli.ParseGlobalOptions(c)
	if err != nil {
		return err
	}

	if !containsFuzzKeyword(*pluginOpts) {
		return fmt.Errorf("please provide the %s keyword", gobusterfuzz.FuzzKeyword)
	}

	log := libgobuster.NewLogger(globalOpts.Debug)

	plugin, err := gobusterfuzz.New(&globalOpts, pluginOpts, log)
	if err != nil {
		return fmt.Errorf("error on creating gobusterfuzz: %w", err)
	}

	if err := internalcli.Gobuster(c.Context, &globalOpts, plugin, log); err != nil {
		log.Debugf("%#v", err)
		return fmt.Errorf("error on running gobuster on %s: %w", pluginOpts.URL, err)
	}
	return nil
}

func containsFuzzKeyword(pluginopts gobusterfuzz.OptionsFuzz) bool {
	if strings.Contains(pluginopts.URL, gobusterfuzz.FuzzKeyword) {
		return true
	}

	if strings.Contains(pluginopts.RequestBody, gobusterfuzz.FuzzKeyword) {
		return true
	}

	for _, h := range pluginopts.Headers {
		if strings.Contains(h.Name, gobusterfuzz.FuzzKeyword) || strings.Contains(h.Value, gobusterfuzz.FuzzKeyword) {
			return true
		}
	}

	if strings.Contains(pluginopts.Username, gobusterfuzz.FuzzKeyword) {
		return true
	}

	if strings.Contains(pluginopts.Password, gobusterfuzz.FuzzKeyword) {
		return true
	}

	return false
}
