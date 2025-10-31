package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/OJ/gobuster/v3/cli/dir"
	"github.com/OJ/gobuster/v3/cli/dns"
	"github.com/OJ/gobuster/v3/cli/fuzz"
	"github.com/OJ/gobuster/v3/cli/gcs"
	"github.com/OJ/gobuster/v3/cli/s3"
	"github.com/OJ/gobuster/v3/cli/tftp"
	"github.com/OJ/gobuster/v3/cli/vhost"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/urfave/cli/v2"

	"go.uber.org/automaxprocs/maxprocs"
)

func main() {
	if _, err := maxprocs.Set(); err != nil {
		fmt.Printf("Error on gomaxprocs: %v\n", err) // nolint forbidigo
	}

	cli.VersionPrinter = func(_ *cli.Context) {
		fmt.Printf("gobuster version %s\n", libgobuster.VERSION) // nolint:forbidigo
		if info, ok := debug.ReadBuildInfo(); ok {
			fmt.Printf("Build info:\n") // nolint forbidigo
			fmt.Printf("%s", info)      // nolint forbidigo
		}
	}

	app := &cli.App{
		Name:      "gobuster",
		Usage:     "the tool you love",
		UsageText: "gobuster command [command options]",
		Authors: []*cli.Author{
			{
				Name: "Christian Mehlmauer (@firefart)",
			},
			{
				Name: "OJ Reeves (@TheColonial)",
			},
		},
		Version: libgobuster.GetVersion(),
		Commands: []*cli.Command{
			dir.Command(),
			vhost.Command(),
			dns.Command(),
			fuzz.Command(),
			tftp.Command(),
			s3.Command(),
			gcs.Command(),
		},
		DisableSliceFlagSeparator: true, // needed so we can specify ',' in slice flags. Otherwise urfave/cli splits on ','
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
