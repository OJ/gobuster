package s3

import (
	"fmt"

	internalcli "github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobusters3"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	cmd := cli.Command{
		Name:   "s3",
		Usage:  "Uses aws bucket enumeration mode",
		Action: run,
		Flags:  getFlags(),
	}
	return &cmd
}

func getFlags() []cli.Flag {
	var flags []cli.Flag
	flags = append(flags, []cli.Flag{
		&cli.IntFlag{Name: "maxfiles", Aliases: []string{"m"}, Value: 5, Usage: "max files to list when listing buckets"},
		&cli.BoolFlag{Name: "show-files", Aliases: []string{"s"}, Value: true, Usage: "show files from found buckets"},
	}...)
	flags = append(flags, internalcli.GlobalOptions()...)
	flags = append(flags, internalcli.BasicHTTPOptions()...)
	return flags
}

func run(c *cli.Context) error {
	pluginOpts := gobusters3.NewOptionsS3()

	httpOptions, err := internalcli.ParseBasicHTTPOptions(c)
	if err != nil {
		return err
	}
	pluginOpts.BasicHTTPOptions = httpOptions

	pluginOpts.MaxFilesToList = c.Int("maxfiles")
	pluginOpts.ShowFiles = c.Bool("show-files")

	globalOpts, err := internalcli.ParseGlobalOptions(c)
	if err != nil {
		return err
	}

	plugin, err := gobusters3.NewGobusterS3(&globalOpts, pluginOpts)
	if err != nil {
		return fmt.Errorf("error on creating gobusters3: %w", err)
	}

	log := libgobuster.NewLogger(globalOpts.Debug)
	if err := internalcli.Gobuster(c.Context, &globalOpts, plugin, log); err != nil {
		log.Debugf("%#v", err)
		return fmt.Errorf("error on running gobuster: %w", err)
	}
	return nil
}
