package main

import (
	"log"
	"os"

	"github.com/lawrsp/pigo/generators/checker"
	"github.com/lawrsp/pigo/generators/convert"
	"github.com/lawrsp/pigo/generators/evalid"
	"github.com/lawrsp/pigo/generators/genrpc"
	"github.com/lawrsp/pigo/generators/jsonfield"
	"github.com/lawrsp/pigo/generators/pfilter"
	"github.com/lawrsp/pigo/generators/setdb"
	"github.com/lawrsp/pigo/generators/setter"
	"github.com/urfave/cli"
)

var version = "0.1.0"

func main() {

	log.SetFlags(log.Lshortfile)
	log.SetPrefix("pigo: ")

	app := cli.NewApp()
	app.Name = "pigo"
	app.Usage = "go auto generate framework"

	app.Version = version
	app.UsageText = "pigo command [command options] [arguments...]"

	commands := []cli.Command{
		{
			Name:        "convert",
			Aliases:     []string{"t"},
			UsageText:   "pigo convert [command options]",
			Usage:       convert.Usage,
			Description: convert.Description,
			Flags:       convert.Flags,
			Action:      convert.Action,
		},
		{
			Name:        "genrpc",
			Aliases:     []string{"g"},
			UsageText:   "pigo genrpc [command options]",
			Usage:       genrpc.Usage,
			Description: genrpc.Description,
			Flags:       genrpc.Flags,
			Action:      genrpc.Action,
		},
		{
			Name:        "setter",
			Aliases:     []string{"s"},
			UsageText:   "pigo convert [command options]",
			Usage:       setter.Usage,
			Description: setter.Description,
			Flags:       setter.Flags,
			Action:      setter.Action,
		},
		{
			Name:        "evalid",
			Aliases:     []string{"e"},
			UsageText:   "pigo evalid [command options]",
			Usage:       evalid.Usage,
			Description: evalid.Description,
			Flags:       evalid.Flags,
			Action:      evalid.Action,
		},
		{
			Name:        "jsonfield",
			Aliases:     []string{"j"},
			UsageText:   "pigo jsonfield [command options]",
			Usage:       jsonfield.Usage,
			Description: jsonfield.Description,
			Flags:       jsonfield.Flags,
			Action:      jsonfield.Action,
		},
		{
			Name:        "checker",
			Aliases:     []string{"c"},
			UsageText:   "pigo checker [command options]",
			Usage:       checker.Usage,
			Description: checker.Description,
			Flags:       checker.Flags,
			Action:      checker.Action,
		},
		{
			Name:        "setdb",
			Aliases:     []string{"d"},
			UsageText:   "pigo setdb [command options]",
			Usage:       setdb.Usage,
			Description: setdb.Description,
			Flags:       setdb.Flags,
			Action:      setdb.Action,
		},
		{
			Name:        "pfilter",
			Aliases:     []string{"p"},
			UsageText:   "pigo pfilter [command options]",
			Usage:       pfilter.Usage,
			Description: pfilter.Description,
			Flags:       pfilter.Flags,
			Action:      pfilter.Action,
		},
	}

	app.Commands = commands

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
