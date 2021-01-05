package setdb

import (
	"log"
	"reflect"
	"strings"

	"github.com/lawrsp/pigo/pkg/builder"
	"github.com/lawrsp/pigo/pkg/generator"
	"github.com/lawrsp/pigo/pkg/parser"
	"github.com/lawrsp/stringstyles"
)

type Config struct {
	Name    string
	TagName string
	Output  string
	Input   string
	Type    string
	DBType  string
	Imports map[string]string
}

// Generator holds the state of the analysis. Primarily used to buffer
// the output for format.Source.
type Generator struct {
	generator.Generator
	Conds []*Condition
}

func NewGenerator() *Generator {
	return &Generator{}
}

// // generate produces the SetDB method for the named type.
// func (g *Generator) generate(typeName string) {
// 	conds := []Condition{}
// 	for _, file := range g.pkg.files {
// 		// Set the state for this run of the walker.
// 		file.typeName = typeName
// 		file.conds = nil
// 		if file.file != nil {
// 			ast.Inspect(file.file, file.genDeclType)
// 		}

// 		if len(file.conds) > 0 {
// 			conds = file.conds
// 			break
// 		}
// 	}

// 	g.buildCondition(p, conds, typeName)
// }

//condition
type Condition struct {
	name    string
	typ     parser.Type
	rowname string
	operand string

	// kind reflect.Kind
}

func newCondition(name string, t parser.Type) *Condition {
	return &Condition{name: name, typ: t}
}

func (c *Condition) InitFromTag(tag string) bool {
	rowname := ""
	operand := ""

	if len(tag) > 0 {
		x := strings.Split(tag, ",")
		rowname = x[0]
		if len(x) > 1 {
			operand = x[1]
		}
	}

	if rowname == "-" {
		return false
	}

	if rowname == "" {
		rowname = stringstyles.SnakeCase(c.name)
	}

	switch {
	case strings.HasPrefix(rowname, "call:"):
		operand = "call"
		rowname = rowname[5:]

	case strings.HasPrefix(rowname, "min_"):
		rowname = rowname[4:]
		if operand == "" {
			operand = ">="
		}

	case strings.HasPrefix(rowname, "max_"):
		rowname = rowname[4:]
		if operand == "" {
			operand = "<="
		}
	case strings.HasSuffix(rowname, "_in"):
		rowname = rowname[0 : len(rowname)-3]
		if operand == "" {
			operand = "IN"
		}
	case strings.HasSuffix(rowname, "_notin"):
		rowname = rowname[0 : len(rowname)-6]
		if operand == "" {
			operand = "NOT IN"
		}
	case strings.HasSuffix(rowname, "_like"):
		rowname = rowname[0 : len(rowname)-5]
		if operand == "" {
			operand = "LIKE"
		}
	case strings.HasSuffix(rowname, "_not"):
		rowname = rowname[0 : len(rowname)-4]
		if operand == "" {
			operand = "IS NOT"
		}
	default:
		if operand == "" {
			operand = "="
		}
	}

	c.rowname = rowname
	if operand == "isnull" {
		c.operand = "IS"
		c.typ = nil
	} else if operand == "notnull" {
		c.operand = "IS NOT"
		c.typ = nil
	} else {
		c.operand = operand
	}

	return true

}

func (g *Generator) buildOnePtrClause(p builder.Printer, cnd *Condition) {
	p.Printf("if p.%s != nil {\n", cnd.name)
	if cnd.operand == "call" {
		p.Printf("  odb = p.%s.%s(odb) \n", cnd.name, cnd.rowname)
	} else if cnd.operand == "IS NOT" {
		p.Printf("  odb = odb.Where(\"(%s = ?) %s TRUE\", *p.%s)\n", cnd.rowname, cnd.operand, cnd.name)
	} else if cnd.operand == "LIKE" {
		p.Printf("  odb = odb.Where(\"%s %s ?\", fmt.Sprintf(\"%%%%%%s%%%%\", *p.%s))\n", cnd.rowname, cnd.operand, cnd.name)
	} else {
		p.Printf("  odb = odb.Where(\"%s %s ?\", *p.%s)\n", cnd.rowname, cnd.operand, cnd.name)
	}
	p.Printf("}\n")
}

func (g *Generator) buildOneSliceClause(p builder.Printer, cnd *Condition) {
	p.Printf("if len(p.%[1]s) > 0 {\n", cnd.name)
	p.Printf("  odb = odb.Where(\"%s %s (?)\", p.%s)\n", cnd.rowname, cnd.operand, cnd.name)
	p.Printf("}\n")
}

func (g *Generator) buildOneDefaultClause(p builder.Printer, cnd *Condition) {
	if cnd.operand == "call" {
		p.Printf("  odb = p.%s.%s(odb) \n", cnd.name, cnd.rowname)
	} else if cnd.operand == "IS NOT" {
		p.Printf("  odb = odb.Where(\"(%s = ?) %s TRUE\", p.%s)\n", cnd.rowname, cnd.operand, cnd.name)
	} else if cnd.operand == "LIKE" {
		p.Printf("odb = odb.Where(\"%s %s ?\", fmt.Sprintf(\"%%%%%%s%%%%\", p.%s))\n", cnd.rowname, cnd.operand, cnd.name)
	} else {
		p.Printf("odb = odb.Where(\"%s %s ?\", p.%s)\n", cnd.rowname, cnd.operand, cnd.name)
	}
}

func (g *Generator) buildOneNilClause(p builder.Printer, cnd *Condition) {
	p.Printf("odb = odb.Where(\"%s %s NULL\")\n", cnd.rowname, cnd.operand)
}

func (g *Generator) PrepareTask(conf *Config) {

	cnds := []*Condition{}

	t := g.File.ReduceTypeSrc(conf.Type)
	fieldList := builder.NewFieldList(conf.TagName)
	_ = parser.InspectUnderlyingStruct(t, fieldList.SpreadInspector)

	for _, fd := range fieldList.Fields {
		name := fd.Field.Name()
		vtag := ""
		if tag := fd.Field.Tag; tag != "" {
			vtag = reflect.StructTag(tag).Get(conf.TagName)
		}

		cnd := newCondition(name, fd.Field.Type)
		if ok := cnd.InitFromTag(vtag); !ok {
			continue
		}

		if cnd.typ != nil {
			if _, ok := cnd.typ.Underlying().(*parser.ArrayType); ok {
				cnd.typ = cnd.typ.Underlying()
			}
		}

		cnds = append(cnds, cnd)
	}

	g.Conds = cnds
}

func (g *Generator) Run(conf *Config) {

	file := builder.NewFile(nil, g.File)

	if conf.Imports != nil && len(conf.Imports) > 0 {
		for name, path := range conf.Imports {
			file.AddImport(name, path)
		}
	}

	bd := builder.NewFuncBuffer(file, conf.Name)

	bd.Printf("\n")
	bd.Printf("func (p *%s) %s(db *%[3]s) *%[3]s {\n", conf.Type, conf.Name, conf.DBType)
	bd.Printf("if p == nil {\n")
	bd.Printf("  return db\n")
	bd.Printf("}\n")
	bd.Printf("odb := db \n")

	for _, cnd := range g.Conds {
		switch cnd.typ.(type) {
		case nil:
			g.buildOneNilClause(bd, cnd)
		case *parser.PointerType:
			g.buildOnePtrClause(bd, cnd)
		case *parser.ArrayType:
			g.buildOneSliceClause(bd, cnd)
		default:
			g.buildOneDefaultClause(bd, cnd)
		}
	}

	bd.Printf("  return odb\n")
	bd.Printf("}\n")

	log.Print(string(bd.Bytes()))
	file.Add(bd)
}

func (g *Generator) Generate(conf *Config) error {

	g.Prepare(".", nil, conf.Output)
	g.PrepareTask(conf)
	g.Run(conf)

	g.Output(conf.Output)
	return nil
}
