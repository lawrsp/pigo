package checker

import (
	"fmt"
	"log"
	"strings"

	"github.com/urfave/cli"
)

var Usage = "generate struct checker function"
var Description = "make a Struct.Validate() function"
var Flags = []cli.Flag{
	//jsonfield -t UpdateParam -n Validate -g tagname -o checkers.go
	cli.StringFlag{
		Name:  "type,t",
		Usage: "the `TYPE` to map field",
	},
	cli.StringFlag{
		Name:  "name,n",
		Usage: "the `FUNCTION` to generate, default is Validate",
	},
	cli.StringFlag{
		Name:  "tag,g",
		Usage: "tag name to define the checker rule, defualt is checker",
	},
	cli.StringSliceFlag{
		Name:  "import,i",
		Usage: "the requried imports",
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
		tag = "checker"
	}
	name := c.String("name")
	if name == "" {
		name = "Validate"
	}

	output := c.String("output")
	config := &Config{
		Type:    t,
		Output:  output,
		TagName: tag,
		Name:    name,
	}

	imports := c.StringSlice("import")
	configImports := map[string]string{}
	for _, impt := range imports {
		kv := strings.Split(impt, ":")
		if len(kv) == 1 {
			path := kv[0]
			names := strings.Split(path, "/")
			name := names[len(names)-1]
			configImports[name] = path
		} else {
			configImports[kv[0]] = kv[1]
		}
	}
	config.Imports = configImports

	g := NewGenerator()
	log.Printf("start generate chekcer:")
	return g.Generate(config)
}
