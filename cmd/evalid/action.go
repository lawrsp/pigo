package evalid

import (
	"github.com/urfave/cli"
)

var Usage = "generate enum valid function"
var Description = "make a isXXXXValid function"
var Flags = []cli.Flag{
	//valid --const target --type string --name isTargetValid --output xxxx.go
	cli.StringFlag{
		Name:  "const,c",
		Usage: "the const to validate",
	},
	cli.StringFlag{
		Name:  "type,t",
		Usage: "the variable type",
	},
	cli.StringFlag{
		Name:  "name,n",
		Usage: "the function `NAME`",
	},
	cli.StringFlag{
		Name:  "output,o",
		Usage: "output to `FILE`",
	},
}

func Action(c *cli.Context) error {
	ct := c.String("const")
	t := c.String("type")
	name := c.String("name")
	output := c.String("output")

	input := c.Args().Get(0)

	config := &Config{
		Const:  ct,
		Type:   t,
		Name:   name,
		Output: output,
		Input:  input,
	}

	g := NewGenerator()

	return g.Generate(config)
}
