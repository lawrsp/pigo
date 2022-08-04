package jsonfield

import (
	"reflect"
	"strings"

	"github.com/lawrsp/pigo/generator/builder"
	"github.com/lawrsp/pigo/generator"
	"github.com/lawrsp/pigo/generator/parser"
	"github.com/lawrsp/stringstyles"
)

type Config struct {
	Type    string
	Name    string
	Output  string
	TagName string
}

type Generator struct {
	generator.Generator
	Receiver   parser.Type
	ResultType parser.Type
	FuncType   parser.Type
	TagName    string
}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) PrepareTask(c *Config) {
	g.TagName = c.TagName

	g.Receiver = parser.TypeWithPointer(g.File.ReduceTypeSrc(c.Type))
	g.ResultType = parser.MapType(parser.NewBasicType("string"), parser.NewBasicType("interface"))

	funcName := c.Name
	if funcName == "" {
		funcName = "JsonMap"
	}

	fnt := &parser.FuncType{
		Receiver: parser.NewField(g.Receiver, "r", ""),
		Params: []*parser.Field{
			parser.NewField(g.ResultType, "info", ""),
		},
		Results: []*parser.Field{
			parser.NewField(g.ResultType, "", ""),
		},
	}

	g.FuncType = parser.TypeWithFile(parser.TypeWithName(fnt, funcName), g.File)
}

func (g *Generator) Run() {

	valType := parser.NewBasicType("interface")

	fileBuilder := builder.NewFile(nil, g.File)
	fb := builder.NewFunction(fileBuilder, g.Receiver, g.FuncType, nil)

	input := builder.GetVariable(fb, g.ResultType, builder.READ_MODE, builder.Scope_Function)
	result := builder.NewVariable(g.ResultType).WithName("result").WriteOnly()
	builder.AddVariableAssign(fb, result, parser.TypeInitValue(result.Type, g.File))

	srcSt := builder.NewFieldList("")
	_ = parser.InspectUnderlyingStruct(g.Receiver, srcSt.SpreadInspector)

	for _, srcFd := range srcSt.Fields {
		fieldName := srcFd.Field.Name()
		stags := reflect.StructTag(srcFd.Field.Tag)

		var name string

		if g.TagName != "" {
			jsonmapTag := stags.Get(g.TagName)
			name = strings.Split(jsonmapTag, ",")[0]
		}

		if name == "-" {
			continue
		}

		if g.TagName != "json" {
			jtag := stags.Get("json")
			name = strings.Split(jtag, ",")[0]
		}

		if name == "-" {
			continue
		}

		if name == "" {
			name = stringstyles.SnakeCase(fieldName)
		}

		v := builder.NewVariable(parser.NewBasicType("bool")).WithName("ok").ReadWrite()
		ignore := builder.NewVariable(valType).WithName("_").ReadWrite()
		init := builder.NewDefine(fb, []*builder.Variable{ignore, v}, input.StringKeyItemExpr(fieldName))
		ifb := builder.NewIfStmt(fb)
		ifb.SetInitCond(init, v.Ident())

		target := builder.NewVariable(valType).WithExpr(result.StringKeyItemExpr(name))
		builder.AddVariableAssign(ifb, target, input.StringKeyItemExpr(fieldName))

		fb.Block().Add(ifb)
	}

	builder.AddSuccessReturn(fb)

	fileBuilder.Add(fb)
}

func (g *Generator) Generate(c *Config) error {
	g.Prepare(".", nil, c.Output)
	g.PrepareTask(c)
	g.Run()
	g.Output(c.Output)
	return nil
}
