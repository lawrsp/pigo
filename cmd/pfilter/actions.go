package pfilter

import (
	"fmt"
	"log"

	"github.com/urfave/cli"
)

var Usage = "generate struct paths filter function"
var Description = "make a FitlerXXX() function"
var Flags = []cli.Flag{
	//jsonfield -t UpdateParam -n Validate -g tagname -o checkers.go
	cli.StringFlag{
		Name:  "type,t",
		Usage: "the `TYPE` to filter field",
	},
	cli.StringFlag{
		Name:  "name,n",
		Usage: "the `FUNCTION` to generate, default is FilterXXX",
	},
	cli.StringFlag{
		Name:  "tag,g",
		Usage: "tag name to define the checker rule, defualt is checker",
	},
	cli.StringFlag{
		Name:  "output,o",
		Usage: "the `FILE` to output",
	},
}

func Action(c *cli.Context) error {
	t := c.String("type")
	if t == "" {
		return fmt.Errorf("type should be given")
	}

	tag := c.String("tag")
	if tag == "" {
		tag = "pfilter"
	}
	name := c.String("name")
	if name == "" {
		name = "FilterByPaths"
	}

	workfile := c.Args().Get(0)

	output := c.String("output")
	config := &Config{
		Type:     t,
		Output:   output,
		TagName:  tag,
		Name:     name,
		WorkFile: workfile,
	}

	g := NewGenerator()
	log.Printf("start generate pfilter %s:", t)
	return g.Generate(config)
}
