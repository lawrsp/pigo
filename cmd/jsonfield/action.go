package jsonfield

import (
	"log"

	"github.com/urfave/cli"
)

var Usage = "generate map to json field function"
var Description = "make a map[string]interface{} key to json field name"
var Flags = []cli.Flag{
	//jsonfield -t UpdateParam -n JsonMap -o setters.go
	cli.StringFlag{
		Name:  "type,t",
		Usage: "the `TYPE` to map field",
	},
	cli.StringFlag{
		Name:  "name,n",
		Usage: "the `FUNCTION` to generate",
	},
	cli.StringFlag{
		Name:  "output,o",
		Usage: "the `FILE` to output",
	},
	cli.StringFlag{
		Name:  "tag,g",
		Usage: "tag name to define the key name",
	},
}

func Action(c *cli.Context) error {
	t := c.String("type")
	tag := c.String("tag")
	name := c.String("name")
	output := c.String("output")
	config := &Config{
		Type:    t,
		Output:  output,
		TagName: tag,
		Name:    name,
	}
	g := NewGenerator()

	log.Printf("start jsonfield:")
	return g.Generate(config)
}
