package pfilter

import (
	"go/ast"
	"log"
	"reflect"
	"strings"

	"github.com/lawrsp/pigo/generator/builder"
	"github.com/lawrsp/pigo/generator"
	"github.com/lawrsp/pigo/generator/parser"
)

type Config struct {
	Type     string
	Name     string
	Output   string
	TagName  string
	WorkFile string
}

type Generator struct {
	generator.Generator
	CheckerType parser.Type
}

func NewGenerator() *Generator {
	return &Generator{}
}

func getTagName(tag reflect.StructTag, tagName string) string {
	jtag := tag.Get(tagName)
	if jtag == "" {
		return ""
	}

	name := strings.Split(jtag, ",")[0]
	return name
}

func (g *Generator) Run(c *Config) {

	file := builder.NewFile(nil, g.File)
	receiverType := g.ReduceTypeSrc(c.Type)

	rName := "p"

	bd := builder.NewFuncBuffer(file, c.Name)
	bd.Printf("func (%s *%s)%s(paths []string)  {\n", rName, c.Type, c.Name)

	bd.Printf("if paths == nil || len(paths) == 0 {\n")
	bd.Printf("return")
	bd.Printf("}\n")

	srcSt := builder.NewFieldList(c.TagName)
	_ = parser.InspectUnderlyingStruct(receiverType, srcSt.SpreadInspector)
	for _, fd := range srcSt.Fields {
		name := fd.Name
		if name == "-" {
			continue
		}
		fieldName := fd.Field.Name()

		bd.Printf("if ok := fieldmaskutil.IsValid(paths, \"%s\"); ok {\n", name)

		stars := parser.GetTypeStars(fd.Field.Type)
		if stars > 0 {
			t := parser.TypeSkipPointer(fd.Field.Type, 1)
			var zeroVal ast.Expr
			if _, ok := t.(*parser.ArrayType); ok {
				zeroVal = parser.TypeInitValue(t, g.WorkFile())
			} else {
				zeroVal = parser.TypeZeroValue(t, g.WorkFile())
			}
			zeroValStr := g.GetExprValueString(zeroVal)
			bd.Printf("if %s.%s == nil {\n", rName, fieldName)
			if bt, ok := t.(*parser.BasicType); ok {
				bd.Printf("var mid %s = %s\n", bt.String(), zeroValStr)
			} else {
				bd.Printf(" mid := %s\n", zeroValStr)
			}

			bd.Printf("%s.%s = &mid\n", rName, fieldName)
			bd.Printf("}\n")
		} else if _, ok := fd.Field.Type.(*parser.ArrayType); ok {
			zeroVal := parser.TypeInitValue(fd.Field.Type, g.WorkFile())
			zeroValStr := g.GetExprValueString(zeroVal)
			bd.Printf("if %s.%s == nil {\n", rName, fieldName)
			bd.Printf("%s.%s = %s\n", rName, fieldName, zeroValStr)
			bd.Printf("}\n")
		} else {
			zeroVal := parser.TypeInitValue(fd.Field.Type, g.WorkFile())
			zeroValStr := g.GetExprValueString(zeroVal)
			bd.Printf("%s.%s = %s\n", rName, fieldName, zeroValStr)
		}

		bd.Printf("} else {")
		if stars > 0 {
			bd.Printf("%s.%s = nil", rName, fieldName)
		} else {
			nilValue := parser.TypeZeroValue(fd.Field.Type, g.WorkFile())
			nilValString := g.GetExprValueString(nilValue)
			bd.Printf("%s.%s = %s", rName, fieldName, nilValString)
		}
		bd.Printf("}\n")
	}

	// bd.Printf("  return %s\n", rName)
	bd.Printf("}\n")

	log.Printf(string(bd.Bytes()))
	file.Add(bd)
}

func (g *Generator) Generate(c *Config) error {
	if c.WorkFile == "" {
		g.Prepare(".", nil, c.Output)
	} else {
		g.PrepareWithFile(c.WorkFile, c.Output)
	}
	g.Run(c)
	g.Output(c.Output)
	return nil
}
