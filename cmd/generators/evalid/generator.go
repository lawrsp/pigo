package evalid

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strings"

	"github.com/lawrsp/pigo/pkg/builder"
	"github.com/lawrsp/pigo/pkg/generator"
	"github.com/lawrsp/pigo/pkg/parser"
)

type Config struct {
	Const  string
	Name   string
	Output string
	Input  string
	Type   string
}

type nameWithPos struct {
	name     string
	position token.Position
}

type Generator struct {
	generator.Generator
	Name         string
	NamePrefix   string
	NamesWithPos map[string][]nameWithPos
	Type         string
}

func NewGenerator() *Generator {
	return &Generator{
		NamesWithPos: map[string][]nameWithPos{},
	}
}

func (g *Generator) collectNames(decl *ast.GenDecl) bool {

	for _, spec := range decl.Specs {
		switch x := spec.(type) {
		case *ast.ValueSpec:
			name := x.Names[0].Name
			if strings.HasPrefix(name, g.NamePrefix) {
				namePos := x.Names[0].NamePos
				position := g.Parser.FileSet.Position(namePos)
				names := g.NamesWithPos[position.Filename]
				if names == nil {
					names = []nameWithPos{}
					g.NamesWithPos[position.Filename] = names
				}

				names = append(names, nameWithPos{
					name:     name,
					position: position,
				})

				g.NamesWithPos[position.Filename] = names
			}
		}
	}

	return true
}

func (g *Generator) PrepareTask(conf *Config) {
	if conf.Const == "" {
		log.Fatalf("no const name specified")
	}
	g.NamePrefix = conf.Const
	g.Name = conf.Name
	g.Type = conf.Type

	if conf.Input != "" {
		file := g.Pkg.GetFile(conf.Input)
		parser.WalkFile(file, parser.NewGenDeclWalker(token.CONST, g.collectNames))
	} else {
		parser.WalkPackage(g.Pkg, parser.NewGenDeclWalker(token.CONST, g.collectNames))
	}

}

func groupNamesByPos(names []nameWithPos) [][]string {

	groups := [][]string{}
	index := -1
	lastLine := -1
	for _, np := range names {
		if index < 0 {
			g := []string{np.name}
			groups = append(groups, g)
			index = 0
			lastLine = np.position.Line
			continue
		}

		thisLine := np.position.Line
		if thisLine > lastLine+1 {
			g := []string{np.name}
			groups = append(groups, g)
			index += 1
		} else {
			groups[index] = append(groups[index], np.name)
		}

		lastLine = thisLine
	}

	return groups
}

func groupNames(names map[string][]nameWithPos) [][]string {
	groups := [][]string{}

	for _, cns := range names {
		groups = append(groups, groupNamesByPos(cns)...)
	}
	return groups
}

func (g *Generator) Run() {
	// argType := parser.NewType(g.Type)
	// resultType := parser.BasicType("bool")
	// fntType := parser.TypeWithName(&parser.FuncType{
	//	Params: []*parser.Field{
	//		parser.NewField(argType, "t", ""),
	//	},
	//	Results: []*parser.Field{
	//		parser.NewField(resultType, "", ""),
	//	},
	// }, name)
	// fb := builder.NewFunction(file, nil, fntType, nil)
	// v := builder.GetVariable(fb, argType, builder.READ_MODE, builder.Scope_Function)
	// if v == nil {
	//	log.Fatalf("cannot find argument variable")
	// }

	file := builder.NewFile(nil, g.File)
	name := g.Name
	if name == "" {
		name = fmt.Sprintf("Is%sValid", g.NamePrefix)
	}

	bd := builder.NewFuncBuffer(file, name)
	bd.Printf("func %s(t %s) bool {\n", name, g.Type)
	bd.Printf("  switch t {\n")

	groupNames := groupNames(g.NamesWithPos)
	for _, group := range groupNames {
		bd.Printf("    case %s:\n", strings.Join(group, ","))
	}
	bd.Printf("    default:\n")
	bd.Printf("    return false\n")
	bd.Printf("  }\n")
	bd.Printf("  return true\n")
	bd.Printf("}\n")

	file.Add(bd)
}

func (g *Generator) Generate(conf *Config) error {
	g.Prepare(".", nil, conf.Output)
	g.PrepareTask(conf)
	g.Run()

	g.Output(conf.Output)
	return nil
}
