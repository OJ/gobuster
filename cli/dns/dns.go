package dns

import (
	"errors"
	"fmt"
	"runtime"
	"time"

	internalcli "github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobusterdns"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	cmd := cli.Command{
		Name:   "dns",
		Usage:  "Uses DNS subdomain enumeration mode",
		Action: run,
		Flags:  getFlags(),
	}
	return &cmd
}

func getFlags() []cli.Flag {
	var flags []cli.Flag
	flags = append(flags, []cli.Flag{
		&cli.StringFlag{Name: "domain", Aliases: []string{"do"}, Usage: "The target domain", Required: true},
		&cli.BoolFlag{Name: "show-ips", Aliases: []string{"i"}, Value: false, Usage: "Show IP addresses of found domains"},
		&cli.BoolFlag{Name: "check-cname", Aliases: []string{"c"}, Value: false, Usage: "Also check CNAME records"},
		&cli.DurationFlag{Name: "timeout", Aliases: []string{"to"}, Value: 1 * time.Second, Usage: "DNS resolver timeout"},
		&cli.BoolFlag{Name: "wildcard", Aliases: []string{"wc"}, Value: false, Usage: "Force continued operation when wildcard found"},
		&cli.BoolFlag{Name: "no-fqdn", Aliases: []string{"nf"}, Value: false, Usage: "Do not automatically add a trailing dot to the domain, so the resolver uses the DNS search domain"},
		&cli.StringFlag{Name: "resolver", Usage: "Use custom DNS server (format server.com or server.com:port)"},
	}...)
	flags = append(flags, internalcli.GlobalOptions()...)
	return flags
}

func run(c *cli.Context) error {
	pluginOpts := gobusterdns.NewOptions()

	pluginOpts.Domain = c.String("domain")
	pluginOpts.ShowIPs = c.Bool("show-ips")
	pluginOpts.CheckCNAME = c.Bool("check-cname")
	pluginOpts.Timeout = c.Duration("timeout")
	pluginOpts.WildcardForced = c.Bool("wildcard")
	pluginOpts.NoFQDN = c.Bool("no-fqdn")
	pluginOpts.Resolver = c.String("resolver")

	if pluginOpts.Resolver != "" && runtime.GOOS == "windows" {
		return fmt.Errorf("currently can not set custom dns resolver on windows. See https://golang.org/pkg/net/#hdr-Name_Resolution")
	}

	globalOpts, err := internalcli.ParseGlobalOptions(c)
	if err != nil {
		return err
	}

	plugin, err := gobusterdns.New(&globalOpts, pluginOpts)
	if err != nil {
		return fmt.Errorf("error on creating gobusterdns: %w", err)
	}

	log := libgobuster.NewLogger(globalOpts.Debug)
	if err := internalcli.Gobuster(c.Context, &globalOpts, plugin, log); err != nil {
		var wErr *gobusterdns.ErrWildcard
		if errors.As(err, &wErr) {
			return fmt.Errorf("%w. To force processing of Wildcard DNS, specify the '--wildcard' switch", wErr)
		}
		log.Debugf("%#v", err)
		return fmt.Errorf("error on running gobuster on %s: %w", pluginOpts.Domain, err)
	}
	return nil
}
