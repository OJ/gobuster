package vhost

import (
	"fmt"

	internalcli "github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobustervhost"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	cmd := cli.Command{
		Name:   "vhost",
		Usage:  "Uses VHOST enumeration mode (you most probably want to use the IP address as the URL parameter)",
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
		&cli.BoolFlag{Name: "append-domain", Aliases: []string{"ad"}, Value: false, Usage: "Append main domain from URL to words from wordlist. Otherwise the fully qualified domains need to be specified in the wordlist."},
		&cli.StringFlag{Name: "exclude-length", Aliases: []string{"xl"}, Usage: "exclude the following content lengths. You can separate multiple lengths by comma and it also supports ranges like 203-206"},
		&cli.StringFlag{Name: "exclude-status", Aliases: []string{"xs"}, Usage: "exclude the following status codes. Can also handle ranges like 200,300-400,404.", Value: ""},
		&cli.StringFlag{Name: "domain", Aliases: []string{"do"}, Usage: "the domain to append when using an IP address as URL. If left empty and you specify a domain based URL the hostname from the URL is extracted"},
	}...)

	return flags
}

func run(c *cli.Context) error {
	pluginOpts := gobustervhost.NewOptions()

	httpOptions, err := internalcli.ParseCommonHTTPOptions(c)
	if err != nil {
		return err
	}
	pluginOpts.HTTPOptions = httpOptions

	pluginOpts.AppendDomain = c.Bool("append-domain")
	pluginOpts.ExcludeLength = c.String("exclude-length")
	ret, err := libgobuster.ParseCommaSeparatedInt(pluginOpts.ExcludeLength)
	if err != nil {
		return fmt.Errorf("invalid value for exclude-length: %w", err)
	}
	pluginOpts.ExcludeLengthParsed = ret

	pluginOpts.ExcludeStatus = c.String("exclude-status")
	ret2, err := libgobuster.ParseCommaSeparatedInt(pluginOpts.ExcludeStatus)
	if err != nil {
		return fmt.Errorf("invalid value for exclude-status: %w", err)
	}
	pluginOpts.ExcludeStatusParsed = ret2

	pluginOpts.Domain = c.String("domain")

	globalOpts, err := internalcli.ParseGlobalOptions(c)
	if err != nil {
		return err
	}

	log := libgobuster.NewLogger(globalOpts.Debug)

	plugin, err := gobustervhost.New(&globalOpts, pluginOpts, log)
	if err != nil {
		return fmt.Errorf("error on creating gobustervhost: %w", err)
	}

	if err := internalcli.Gobuster(c.Context, &globalOpts, plugin, log); err != nil {
		log.Debugf("%#v", err)
		return fmt.Errorf("error on running gobuster on %s: %w", pluginOpts.URL, err)
	}
	return nil
}
