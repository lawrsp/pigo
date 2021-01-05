package setdb

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"
)

var Usage = "generate enum valid function"
var Description = "make a isXXXXValid function"
var Flags = []cli.Flag{
	//setdb  --type target --name SetDB --output xxxx.go
	cli.StringFlag{
		Name:  "type,t",
		Usage: "the variable type",
	},
	cli.StringFlag{
		Name:  "tag,g",
		Usage: "the tag name",
		Value: "setdb",
	},
	cli.StringFlag{
		Name:  "name,n",
		Usage: "the function `NAME`",
		Value: "SetSQL",
	},
	cli.StringFlag{
		Name:  "output,o",
		Usage: "output to `FILE`",
	},
	cli.StringFlag{
		Name:  "db",
		Usage: "specify the db type",
		Value: "sql.DB",
	},
	cli.StringSliceFlag{
		Name:  "import,i",
		Usage: "the requried imports",
	},
}

func Action(c *cli.Context) error {
	t := c.String("type")
	if t == "" {
		return fmt.Errorf("type should be specified")
	}
	name := c.String("name")
	tagName := c.String("tag")

	output := c.String("output")
	input := c.Args().Get(0)

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

	config := &Config{
		Type:    t,
		Name:    name,
		TagName: tagName,
		Output:  output,
		Input:   input,
		DBType:  c.String("db"),
		Imports: configImports,
	}

	g := NewGenerator()

	return g.Generate(config)
}
