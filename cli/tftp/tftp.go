package tftp

import (
	"fmt"
	"strings"
	"time"

	internalcli "github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobustertftp"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	cmd := cli.Command{
		Name:   "tftp",
		Usage:  "Uses TFTP enumeration mode",
		Action: run,
		Flags:  getFlags(),
	}
	return &cmd
}

func getFlags() []cli.Flag {
	var flags []cli.Flag
	flags = append(flags, []cli.Flag{
		&cli.StringFlag{Name: "server", Aliases: []string{"s"}, Usage: "The target TFTP server", Required: true},
		&cli.DurationFlag{Name: "timeout", Aliases: []string{"to"}, Value: 1 * time.Second, Usage: "TFTP timeout"},
	}...)
	flags = append(flags, internalcli.GlobalOptions()...)
	return flags
}

func run(c *cli.Context) error {
	pluginOpts := gobustertftp.NewOptions()

	pluginOpts.Server = c.String("server")
	if !strings.Contains(pluginOpts.Server, ":") {
		pluginOpts.Server = fmt.Sprintf("%s:69", pluginOpts.Server)
	}

	pluginOpts.Timeout = c.Duration("timeout")

	globalOpts, err := internalcli.ParseGlobalOptions(c)
	if err != nil {
		return err
	}

	plugin, err := gobustertftp.New(&globalOpts, pluginOpts)
	if err != nil {
		return fmt.Errorf("error on creating gobustertftp: %w", err)
	}

	log := libgobuster.NewLogger(globalOpts.Debug)
	if err := internalcli.Gobuster(c.Context, &globalOpts, plugin, log); err != nil {
		log.Debugf("%#v", err)
		return fmt.Errorf("error on running gobuster on %s: %w", pluginOpts.Server, err)
	}
	return nil
}
