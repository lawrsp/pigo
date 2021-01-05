package convert

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"github.com/lawrsp/pigo/pkg/configutil"
)

var Usage = "convert one type to another"
var Description = "As default, it will read task defination from stdin, if --file was given, it would read from the file."
var Flags = []cli.Flag{
	cli.StringFlag{
		Name:  "file,f",
		Usage: "read task defination from `FILE`",
	},
	cli.StringFlag{
		Name:  "source,s",
		Usage: "the `TYPE` convert fromt",
	},
	cli.StringFlag{
		Name:  "target,t",
		Usage: "the `TYPE` convert to",
	},
	cli.StringFlag{
		Name:  "name",
		Usage: "the convert `FUNCTION` name ",
	},
	cli.StringSliceFlag{
		Name:  "assign,a",
		Usage: "the exists covnert functions",
	},
	cli.StringFlag{
		Name:  "error",
		Usage: "the error returns",
	},
	cli.StringFlag{
		Name:  "output,o",
		Usage: "the `FILE` to output",
	},
	cli.StringSliceFlag{
		Name:  "import,i",
		Usage: "the requried imports",
	},
	cli.StringFlag{
		Name:  "tag,g",
		Usage: "the tag name",
		Value: "pc",
	},
}

/*
	Dir       string
	Files     []string
	Imports   map[string]string
	Output    string
	Assigns   map[string]YamlCustomAssign
	Generates map[string]YamlTaskElem
*/

func Action(c *cli.Context) error {
	filePath := c.String("file")
	if filePath != "" {
		config := &YamlConfig{}
		if err := configutil.ReadConfig(filePath, config); err != nil {
			return err
		}

		if config.Version != "" && config.Version != "1" {
			return fmt.Errorf("Version not supported")
		}
		g := NewGenerator()
		return g.Generate(config)
	}
	/*
			Version   string
		Dir       string
		Files     []string
		Imports   map[string]string
		Output    string
		Assigns   map[string]YamlCustomAssign
		Generates map[string]YamlTaskElem
	*/
	config := &YamlConfig{}
	file := c.Args().Get(0)
	if file == "" {
		config.Dir = "."
	} else {
		config.Files = []string{file}
	}
	config.Output = c.String("output")

	src := c.String("source")
	if src == "" {
		return fmt.Errorf("should specify a source type")
	}
	tgt := c.String("target")
	if tgt == "" {
		return fmt.Errorf("should specify a target type")
	}
	name := c.String("name")
	if name == "" {
		name = fmt.Sprintf("convert%sTo%s", src, tgt)
	}

	genTask := YamlTaskElem{
		Name:   name,
		Source: src,
		Target: tgt,
	}
	ec := c.String("error")
	if ec == "false" {
		genTask.WithoutError = true
	} else if ec == "true" {
		genTask.WithoutError = false
	} else if ec != "" {
		genTask.SourceError = ec
	}
	config.Generates = map[string]YamlTaskElem{name: genTask}

	assigns := c.StringSlice("assign")
	configAssigns := map[string]YamlCustomAssign{}
	for _, as := range assigns {
		names := strings.Split(as, ":")
		ya := YamlCustomAssign{}
		ya.Source = names[0]
		ya.Target = names[1]
		ya.Assign = names[2]
		if len(names) > 3 {
			ya.Check = names[3]
		}
		configAssigns[ya.Assign] = ya
	}
	config.Assigns = configAssigns

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

	tagname := c.String("tag")
	config.TagName = tagname

	// jsonutil.Pretty(config, os.Stdout)
	g := NewGenerator()
	return g.Generate(config)
}
