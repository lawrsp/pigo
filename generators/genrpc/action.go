package genrpc

import (
	"fmt"

	"github.com/lawrsp/pigo/pkg/configutil"
	"github.com/urfave/cli"
)

var Usage = "auto generate rpc methods"
var Description = "As default, it will read task defination from stdin, if --file was given, it would read from the file."

var Flags = []cli.Flag{
	cli.StringFlag{
		Name:  "file,f",
		Usage: "read task defination from `FILE`",
	},
}

func Action(c *cli.Context) error {
	filePath := c.String("file")
	config := &YamlConfig{}

	if err := configutil.ReadConfig(filePath, config); err != nil {
		return err
	}

	if config.Version != "" && config.Version != "1" {
		return fmt.Errorf("Version not supported")
	}

	g := NewGenerator()

	if err := g.Generate(config); err != nil {
		return err
	}
	return nil
}
