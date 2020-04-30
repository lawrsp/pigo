package setter

import (
	"fmt"
	"go/ast"
	"log"
	"reflect"
	"strings"

	"github.com/lawrsp/pigo/pkg/builder"
	"github.com/lawrsp/pigo/pkg/generator"
	"github.com/lawrsp/pigo/pkg/parser"
	"github.com/lawrsp/pigo/pkg/tagutil"
)

type CustomAssign struct {
	Source parser.Type
	Target parser.Type
	Assign parser.Type
	Check  string
}

type AssignConfig struct {
	Source string
	Target string
	Assign string
	Check  string
}

type Config struct {
	Type       string
	Receiver   string
	Target     string
	Name       string
	Withmap    bool
	WithOldMap bool
	CheckDiff  bool
	Output     string
	MapTag     string
	Imports    map[string]string
	Assigns    []*AssignConfig
}

type Generator struct {
	generator.Generator

	TagName    string
	Type       parser.Type
	Receiver   parser.Type
	FuncType   parser.Type
	Withmap    bool
	WithOldMap bool
	CheckDiff  bool
	Reverse    bool
	Assigns    []*CustomAssign
	MapTag     string

	updateV *builder.Variable
	oldV    *builder.Variable
}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) PrepareAssigns(configAssigns []*AssignConfig) {
	assigns := []*CustomAssign{}

	for _, v := range configAssigns {
		cp := &CustomAssign{}
		cp.Source = g.ReduceTypeSrc(v.Source)
		if cp.Source == nil {
			log.Fatalf("type %s not reduced", v.Source)
		}
		cp.Target = g.ReduceTypeSrc(v.Target)
		if cp.Target == nil {
			log.Fatalf("type %s not reduced", v.Target)
		}
		cp.Assign = g.ReduceTypeSrc(v.Assign)
		if cp.Assign == nil {
			log.Fatalf("type %s not reduced", v.Assign)
		}
		cp.Check = v.Check
		assigns = append(assigns, cp)
	}

	g.Assigns = assigns
}

func (g *Generator) PrepareTask(config *Config) {
	g.Type = parser.TypeWithPointer(g.File.ReduceTypeSrc(config.Type))
	funcName := config.Name
	var fnt *parser.FuncType

	if config.Receiver != "" {
		g.Receiver = parser.TypeWithPointer(g.File.ReduceTypeSrc(config.Receiver))

		if funcName == "" {
			funcName = fmt.Sprintf("Set%s", config.Type)
		}

		fnt = &parser.FuncType{
			Receiver: parser.NewField(g.Receiver, "r", ""),
			Params: []*parser.Field{
				parser.NewField(g.Type, "info", ""),
			},
			Results: []*parser.Field{},
		}
		g.Reverse = false
	} else if config.Target != "" {
		log.Printf("target: %s", config.Target)
		tRecv := g.File.ReduceTypeSrc(config.Target)
		if tRecv == nil {
			log.Fatalf("cannot reduce type %s", config.Target)
		}
		g.Receiver = parser.TypeWithPointer(tRecv)
		if funcName == "" {
			funcName = fmt.Sprintf("Set%s", config.Target)
		}

		fnt = &parser.FuncType{
			Receiver: parser.NewField(g.Type, "t", ""),
			Params: []*parser.Field{
				parser.NewField(g.Receiver, "target", ""),
			},
			Results: []*parser.Field{},
		}
		g.Reverse = true
	}

	if config.Withmap {
		mapType := parser.MapType(parser.NewBasicType("string"), parser.NewBasicType("interface"))
		fnt.Results = append(fnt.Results, parser.NewField(mapType, "updated", ""))
	}

	if config.WithOldMap {
		mapType := parser.MapType(parser.NewBasicType("string"), parser.NewBasicType("interface"))
		fnt.Results = append(fnt.Results, parser.NewField(mapType, "old", ""))
	}
	g.MapTag = config.MapTag

	g.FuncType = parser.TypeWithFile(parser.TypeWithName(fnt, funcName), g.File)

	g.Withmap = config.Withmap
	g.WithOldMap = config.WithOldMap
	g.CheckDiff = config.CheckDiff
}

func (g *Generator) GetCallFunc(field *builder.Field) parser.Type {
	tags := field.Field.Tag
	if len(tags) == 0 {
		return nil
	}
	stag := reflect.StructTag(tags).Get(g.TagName)
	if len(stag) == 0 {
		return nil
	}

	stags := strings.Split(stag, ",")
	if len(stags) < 1 {
		return nil
	}

	callFunc := stags[1]

	expr := parser.TypeExprInFile(g.Receiver, g.File)

	t := g.File.ReduceType(builder.DotExpr(expr, ast.NewIdent(callFunc)))
	// log.Printf("finded call func: %s", t)
	return t
}

func (g *Generator) GetFieldMapName(field *builder.Field) string {
	// fmt.Println("=== g.MapTag:", g.MapTag, field.Field.Tag)
	if g.MapTag != "" {
		tags := field.Field.Tag
		return tagutil.GetTagValue(reflect.StructTag(tags), g.MapTag)
	}

	return field.Field.Name()
}

func (g *Generator) AssignerFunc(field *builder.Field) func(builder.Builder, *builder.Variable, ast.Expr) {
	fieldName := g.GetFieldMapName(field)
	return func(inBuilder builder.Builder, v *builder.Variable, value ast.Expr) {
		if g.oldV != nil {
			collector := builder.NewVariable(v.Type).WriteOnly().WithExpr(g.oldV.StringKeyItemExpr(fieldName))
			builder.AddVariableAssign(inBuilder, collector, v.Ident())
		}

		builder.AddVariableAssign(inBuilder, v, value)

		if g.updateV != nil {
			collector := builder.NewVariable(v.Type).WriteOnly().WithExpr(g.updateV.StringKeyItemExpr(fieldName))
			builder.AddVariableAssign(inBuilder, collector, v.Ident())
		}
	}
}

func (g *Generator) Run() {

	outer := builder.NewFile(nil, g.File)
	var fb *builder.FuncBuilder

	if g.Reverse {
		fb = builder.NewFunction(outer, nil, g.FuncType, nil)
	} else {
		fb = builder.NewFunction(outer, nil, g.FuncType, nil)
	}

	tpaths := []*parser.TPath{}
	for _, assign := range g.Assigns {
		if _, ok := assign.Assign.Underlying().(*parser.FuncType); ok {
			tpaths = append(tpaths, parser.NewTPath(assign.Source, assign.Target).WithFunction(assign.Assign))
		} else {
			tpaths = append(tpaths, parser.NewTPath(assign.Source, assign.Target).WithTypeConversion(assign.Assign))
		}
	}

	structSetter := builder.NewStructSetter(fb, g.TagName, nil, tpaths)
	if g.CheckDiff {
		structSetter = structSetter.WithCheckDiff()
	}
	srcV := builder.GetVariable(fb, g.Type, builder.READ_MODE, builder.Scope_Function)
	var checkV *builder.Variable
	if g.Reverse {
		checkV = builder.GetVariable(fb, g.Receiver, builder.READ_MODE, builder.Scope_Function)
	} else {
		checkV = srcV
	}
	builder.AddCheckReturn(fb, checkV.CheckNilExpr(true), nil)

	if g.WithOldMap {
		mapType := parser.MapType(parser.NewBasicType("string"), parser.NewBasicType("interface"))
		oldV := builder.NewVariable(mapType).WithName("old").ReadOnly()
		g.oldV = builder.AddVariableAssign(fb, oldV, parser.TypeInitValue(mapType, g.File))
	}
	if g.Withmap {
		mapType := parser.MapType(parser.NewBasicType("string"), parser.NewBasicType("interface"))
		updateV := builder.NewVariable(mapType).WithName("updated").ReadOnly()
		g.updateV = builder.AddVariableAssign(fb, updateV, parser.TypeInitValue(mapType, g.File))
	}

	if g.Withmap || g.WithOldMap {
		structSetter = structSetter.WithAssigner(g.AssignerFunc)
	}

	skip := 0
	if g.Receiver.EqualTo(g.Type) {
		skip = 1
	}
	dstV := builder.GetVariableWithSkip(fb, g.Receiver, builder.READ_MODE, builder.Scope_Function, skip)

	if ok := structSetter.TryAssign(srcV, dstV); ok {
		fb.Block().Add(structSetter)
	}

	if fb.HasResults() {
		builder.AddSuccessReturn(fb)
	}

	outer.Add(fb)
}

func (g *Generator) Generate(config *Config) error {

	g.Prepare(".", nil, config.Output)
	g.PrepareImports(config.Imports)
	g.PrepareAssigns(config.Assigns)
	g.PrepareTask(config)
	g.Run()
	g.Output(config.Output)

	return nil
}
