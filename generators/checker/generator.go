package checker

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/lawrsp/pigo/pkg/builder"
	"github.com/lawrsp/pigo/pkg/generator"
	"github.com/lawrsp/pigo/pkg/parser"
	"github.com/lawrsp/pigo/pkg/printutil"
	"github.com/lawrsp/stringstyles"
)

type Config struct {
	Type    string
	Name    string
	Output  string
	TagName string
	Imports map[string]string
}

type Generator struct {
	generator.Generator
}

func NewGenerator() *Generator {
	return &Generator{}
}

type CheckerInfo struct {
	chk        string
	name       string
	expr       string
	targetType parser.Type
	file       *parser.File
}

func (c *CheckerInfo) Copy() *CheckerInfo {
	x := &CheckerInfo{}
	*x = *c
	return x
}

type Checker struct {
	c     *CheckerInfo
	p     builder.Printer
	index int
	procs []CheckerProc
}

func NewChecker(c *CheckerInfo, p builder.Printer, tagSrc string) *Checker {
	cc := &Checker{c: c, p: p, index: 0, procs: []CheckerProc{}}

	confs := strings.Split(tagSrc, ";")

	for _, cf := range confs {
		kv := strings.Split(cf, ":")
		switch kv[0] {
		case "noempty":
			proc := &NoEmptyProc{}
			if len(kv) > 1 {
				proc.emptyVal = kv[1]
			}
			cc.procs = append(cc.procs, proc)
		case "isvalid":
			proc := &IsValidProc{}
			proc.call = kv[1]
			cc.procs = append(cc.procs, proc)
		case "stars":
			proc := &StarProc{}
			proc.stars = getInt(kv[1])
			cc.procs = append(cc.procs, proc)
		case "default":
			s := strings.Replace(kv[1], "'", "\"", -1)
			proc := &DefaultValueProc{}
			proc.defaultValue = s
			cc.procs = append(cc.procs, proc)
		case "arrays":
			proc := &ArrayProc{}
			proc.arrays = getInt(kv[1])
			cc.procs = append(cc.procs, proc)
		case "call":
			proc := &CallProc{}
			proc.call = kv[1]
			proc.result = kv[2]
			cc.procs = append(cc.procs, proc)
		case "compare":
			proc := &CompareProc{}
			proc.operand = kv[1]
			proc.value = kv[2]
			cc.procs = append(cc.procs, proc)
		case "merge":
			proc := &MergeProc{}
			proc.method = kv[1]
			cc.procs = append(cc.procs, proc)
		case "convert":
			proc := &ConvertProc{}
			proc.convertTo = kv[1]
			if len(kv) >= 3 {
				proc.convertName = kv[2]
			}
			cc.procs = append(cc.procs, proc)
		}
	}

	return cc
}

func (c *Checker) Next() {
	for c.index < len(c.procs) {
		idx := c.index
		c.index += 1
		c.procs[idx].Print(c)
	}
}

type CheckerProc interface {
	Print(cm *Checker)
}

type NoEmptyProc struct {
	emptyVal string
}

func (proc *NoEmptyProc) Print(cm *Checker) {
	p := cm.p
	c := cm.c

	if _, ok := c.targetType.(*parser.ArrayType); ok {
		p.Printf("%s.Assert(len(%s) != 0, %s, \"should not be empty\")\n", c.chk, c.expr, c.name)
	} else if proc.emptyVal != "" {
		p.Printf("%s.Assert(%s != %s, %s, \"should not be empty\")\n", c.chk, c.expr, proc.emptyVal, c.name)
	} else {
		emptyValueExpr := parser.TypeZeroValue(c.targetType, c.file)
		emptyValue := getEmptyValueString(emptyValueExpr)
		p.Printf("%s.Assert(%s != %s, %s, \"should not be empty\")\n", c.chk, c.expr, emptyValue, c.name)
	}
}

type IsValidProc struct {
	call string
}

func (cp *IsValidProc) Print(cm *Checker) {
	p := cm.p
	c := cm.c

	p.Printf("%s.Assert(%s(%s), %s, \"invalid\" )\n", c.chk, cp.call, c.expr, c.name)
}

type CallProc struct {
	call   string
	result string
}

func (cp *CallProc) Print(cm *Checker) {
	p := cm.p
	c := cm.c

	expr := c.expr
	for expr[0] == '*' {
		expr = expr[1:]
	}

	if cp.result == "bool" || cp.result == "true" || cp.result == "" {
		p.Printf("%s.Assert(%s.%s(), %s, \"invalid\" )\n", c.chk, expr, cp.call, c.name)
	} else if cp.result == "false" {
		p.Printf("%s.Assert(!%s.%s(), %s, \"invalid\" )\n", c.chk, expr, cp.call, c.name)
	} else if cp.result == "error" {
		p.Printf("%s.AssertError(%s.%s(), %s, \"invalid\"  )\n", c.chk, expr, cp.call, c.name)
	} else {
		log.Fatalf("cannot generate call checker %s:", cp.call)
	}

}

type CompareProc struct {
	operand string
	value   string
}

func (cp *CompareProc) Print(cm *Checker) {
	p := cm.p
	c := cm.c

	p.Printf("%s.Assert(%s %s %s, %s, \"invalid\" )\n", c.chk, c.expr, cp.operand, cp.value, c.name)
}

type DefaultValueProc struct {
	defaultValue string
}

func (cp *DefaultValueProc) Print(cm *Checker) {
	p := cm.p
	c := cm.c

	emptyValueExpr := parser.TypeZeroValue(c.targetType, c.file)
	emptyValue := getEmptyValueString(emptyValueExpr)

	p.Printf("if %s == %s {\n  %s = %s  }\n", c.expr, emptyValue, c.expr, cp.defaultValue)
}

type MergeProc struct {
	method string
}

func (cp *MergeProc) Print(cm *Checker) {
	p := cm.p
	c := cm.c

	p.Printf("if err := %s.%s(); err != nil {\n", c.expr, cp.method)
	p.Printf("%s.Merge(err)\n", c.chk)
	p.Printf("}\n")

}

type ArrayProc struct {
	arrays int
}

func (cp *ArrayProc) Print(cm *Checker) {
	if cp.arrays == 0 {
		return
	}

	p := cm.p
	c := cm.c

	expr := c.expr
	it := "i"

	for i := 0; i < cp.arrays; i++ {
		it += "t"
		p.Printf("for i, %s := range %s {\n", it, expr)
	}

	nc := c.Copy()
	nc.expr = it
	theName, err := strconv.Unquote(c.name)
	if err != nil {
		log.Fatalf("unquote %s failed: %v", c.name, err)
	}
	nc.name = "fmt.Sprintf(\"" + theName + ".%d\", i)"
	nc.targetType = parser.TypeSkipBracket(c.targetType, cp.arrays)
	cm.c = nc
	cm.Next()

	for i := 0; i < cp.arrays; i++ {
		p.Printf("}\n")
	}
}

type ConvertProc struct {
	convertTo   string
	convertName string
}

func (cp *ConvertProc) Print(cm *Checker) {

	p := cm.p
	c := cm.c

	expr := c.expr
	convertedExpr := cp.convertName
	if convertedExpr == "" {
		exprDotSlice := strings.Split(expr, ".")
		convertedExpr = "cnvt" + exprDotSlice[len(exprDotSlice)-1]
	}
	convertType := cm.c.file.ReduceTypeSrc(cp.convertTo)

	p.Printf("%s := %s(%s)\n", convertedExpr, cp.convertTo, expr)

	nc := c.Copy()
	nc.expr = convertedExpr
	nc.targetType = convertType

	cm.c = nc
	cm.Next()
}

type StarProc struct {
	stars int
}

func (cp *StarProc) Print(cm *Checker) {
	if cp.stars == 0 {
		return
	}

	p := cm.p
	c := cm.c

	conditions := []string{}
	for i := 0; i < cp.stars; i++ {
		x := fmt.Sprintf("%s != nil", strings.Repeat("*", i)+c.expr)
		conditions = append(conditions, x)
	}
	cond := strings.Join(conditions, " && ")
	p.Printf("if %s {\n", cond)

	nc := c.Copy()
	nc.expr = strings.Repeat("*", cp.stars) + c.expr
	nc.targetType = parser.TypeSkipPointer(c.targetType, cp.stars)
	cm.c = nc
	cm.Next()

	p.Printf("}\n")

}

func getInt(s string) int {
	if i, err := strconv.ParseInt(s, 10, 0); err == nil {
		return int(i)
	}

	return 0
}

func getJsonTagName(tag reflect.StructTag) string {
	jtag := tag.Get("json")
	if jtag == "" {
		return ""
	}

	name := strings.Split(jtag, ",")[0]
	return name
}

func getEmptyValueString(expr ast.Expr) string {

	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.BasicLit:
		if x.Kind == token.STRING {
			return fmt.Sprintf("\"%s\"", x.Value)
		} else {
			return x.Value
		}
	case *ast.CompositeLit:
		if _, ok := x.Type.(*ast.ArrayType); ok {
			return "nil"
		}
	}

	printutil.PrintNodef(expr, "cannot support expr:")
	log.Fatalf("cannot get value string")
	return ""
}

func (g *Generator) Run(c *Config) {

	receiver := g.File.ReduceTypeSrc(c.Type)
	if receiver == nil {
		log.Fatalf("cannot reduce type %s", c.Type)
	}
	srcSt := builder.NewFieldList(c.TagName)
	_ = parser.InspectUnderlyingStruct(receiver, srcSt.SpreadInspector)

	file := builder.NewFile(nil, g.File)

	rName := "p"

	bd := builder.NewFuncBuffer(file, c.Name)
	bd.Printf("func (%s *%s)%s() error {\n", rName, c.Type, c.Name)
	bd.Printf("  chk := checker.NewParamChecker()\n")

	for _, srcFd := range srcSt.Fields {
		tags := reflect.StructTag(srcFd.Field.Tag)
		fieldName := srcFd.Field.Name()

		ctags := tags.Get(c.TagName)
		if ctags == "" {
			continue
		}

		var name string
		names := strings.Split(ctags, ",")
		if len(names) > 1 {
			name = names[0]
			ctags = names[1]
		} else {
			name = getJsonTagName(tags)
			if name == "" {
				name = stringstyles.SnakeCase(srcFd.Field.Name())
			}
		}

		info := &CheckerInfo{}
		info.expr = fmt.Sprintf("%s.%s", rName, fieldName)
		info.targetType = srcFd.Field.Type
		info.chk = "chk"
		info.name = fmt.Sprintf("\"%s\"", name)
		info.file = g.File

		cc := NewChecker(info, bd, ctags)
		cc.Next()
		bd.Printf("\n")
	}

	bd.Printf("  return chk.GetError()\n")
	bd.Printf("}\n")

	log.Printf(string(bd.Bytes()))
	file.Add(bd)
}

func (g *Generator) Generate(c *Config) error {
	g.Prepare(".", nil, c.Output)
	g.PrepareImports(c.Imports)
	g.Run(c)
	g.Output(c.Output)
	return nil
}
