package setter

import (
	"errors"
	"log"
	"strings"

	"github.com/urfave/cli"
)

var Usage = "generate  setter function"
var Description = "make a Receiver.SetXXX function"
var Flags = []cli.Flag{
	//setter -t SomeType --target Target -name SetUpdate --withmap --checkdiff --withold
	//   => func(p *SomeType) SetUpdate(dst *Target) (updates map[string]interface{}, old map[string]interface{})
	//setter -t SomeType --receiver Receiver -name SetUpdate --withmap --checkdiff --withold
	//   => func(p *Receiver) SetUpdate(dst *SomeType) (updates map[string]interface{}, old map[string]interface{})
	cli.StringFlag{
		Name:  "type,t",
		Usage: "the `TYPE` to setter input",
	},
	cli.StringFlag{
		Name:  "receiver,r",
		Usage: "the `RECEIVER` of the setter",
	},
	cli.StringFlag{
		Name:  "target",
		Usage: "the `TARGET` of the setter",
	},
	cli.StringFlag{
		Name:  "name,n",
		Usage: "the function `NAME`",
		Value: "Set",
	},
	cli.BoolFlag{
		Name:  "withmap,m",
		Usage: "with a map[string]interface{} return",
	},
	cli.BoolFlag{
		Name:  "withold,j",
		Usage: "a map[string]interface{} with old values will return",
	},
	cli.BoolFlag{
		Name:  "checkdiff,d",
		Usage: "check if different with origin value",
	},
	cli.StringFlag{
		Name:  "maptag",
		Usage: "determin the update/old map's key by tag",
	},
	cli.StringFlag{
		Name:  "output,o",
		Usage: "output to `FILE`",
	},
	cli.StringSliceFlag{
		Name:  "import,i",
		Usage: "the requried imports",
	},
	cli.StringSliceFlag{
		Name:  "assign,a",
		Usage: "the exists covnert functions",
	},
	cli.StringFlag{
		Name:  "tag,g",
		Usage: "the tag name",
		Value: "setter",
	},
}

func Action(c *cli.Context) error {
	t := c.String("type")
	receiver := c.String("receiver")
	name := c.String("name")
	withmap := c.Bool("withmap")
	withold := c.Bool("withold")
	output := c.String("output")
	target := c.String("target")
	checkDiff := c.Bool("checkdiff")
	mapTag := c.String("maptag")

	if target != "" && receiver != "" {
		return errors.New("receiver and target cannot be used together")
	}
	if target == "" && receiver == "" {
		return errors.New("please specify recevier or target")
	}

	config := &Config{
		Type:       t,
		Receiver:   receiver,
		Name:       name,
		Withmap:    withmap,
		WithOldMap: withold,
		MapTag:     mapTag,
		Output:     output,
		Target:     target,
		CheckDiff:  checkDiff,
	}

	// assigns:
	assigns := c.StringSlice("assign")
	configAssigns := []*AssignConfig{}
	for _, as := range assigns {
		names := strings.Split(as, ":")
		ya := &AssignConfig{}
		ya.Source = names[0]
		ya.Target = names[1]
		ya.Assign = names[2]
		if len(names) > 3 {
			ya.Check = names[3]
		}
		configAssigns = append(configAssigns, ya)
	}
	config.Assigns = configAssigns

	// imports:
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
	g.TagName = c.String("tag")

	log.Printf("start setter %s", g.TagName)
	return g.Generate(config)
}
