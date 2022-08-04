package checker

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/lawrsp/pigo/generator"
	"github.com/lawrsp/pigo/generator/builder"
	"github.com/lawrsp/pigo/generator/parser"
	"github.com/lawrsp/pigo/generator/printutil"
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
	fullExpr   string
	targetType parser.Type
	file       *parser.File
}

func (c *CheckerInfo) Copy() *CheckerInfo {
	x := &CheckerInfo{}
	*x = *c
	return x
}

type Checker struct {
	c *CheckerInfo
	p builder.Printer

	index int
	procs []CheckerProc
}

func NewChecker(c *CheckerInfo, p builder.Printer, tagSrc string) *Checker {
	cc := &Checker{c: c, p: p, index: 0, procs: []CheckerProc{}}

	confs := strings.Split(tagSrc, ";")

	for _, cf := range confs {
		vs := strings.Split(cf, ":")
		switch vs[0] {
		case "noempty":
			proc := NewNoEmptyProc(vs)
			cc.procs = append(cc.procs, proc)
		case "isvalid":
			proc := NewIsValidProc(vs)
			cc.procs = append(cc.procs, proc)
		case "stars":
			proc := NewStarProc(vs)
			cc.procs = append(cc.procs, proc)
		case "default":
			proc := NewDefaultValueProc(vs)
			cc.procs = append(cc.procs, proc)
		case "arrays":
			proc := NewArrayProc(vs)
			cc.procs = append(cc.procs, proc)
		case "call":
			proc := NewCallProc(vs)
			cc.procs = append(cc.procs, proc)
		case "compare":
			proc := NewCompareProc(vs)
			cc.procs = append(cc.procs, proc)
		case "merge":
			proc := NewMergeProc(vs)
			cc.procs = append(cc.procs, proc)
		case "convert":
			proc := NewConvertProc(vs)
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

type baseCheckerProc struct {
	ErrType  string
	Messages []string
}

type NoEmptyProc struct {
	baseCheckerProc
	emptyVal string
}

// noemtpy:emptyValue:CustomType:msgA:msgB:msgC...
func NewNoEmptyProc(vs []string) CheckerProc {
	p := &NoEmptyProc{}
	p.ErrType = "\"IsEmpty\""

	if len(vs) > 1 {
		p.emptyVal = strings.Replace(vs[1], "'", "\"", -1)
	}

	if len(vs) > 2 && vs[2] != "" {
		p.ErrType = vs[2]
	}

	if len(vs) > 3 {
		p.Messages = vs[3:]
	}

	return p
}

func (proc *NoEmptyProc) Print(cm *Checker) {
	p := cm.p
	c := cm.c

	p.Printf("%s.Assert(", c.chk)

	if _, ok := c.targetType.(*parser.ArrayType); ok {
		p.Printf("len(%s) != 0", c.expr)
	} else if proc.emptyVal != "" {
		p.Printf("%s != %s", c.expr, proc.emptyVal)
	} else {
		emptyValueExpr := parser.TypeZeroValue(c.targetType, c.file)
		emptyValue := getEmptyValueString(emptyValueExpr)
		p.Printf("%s != %s", c.expr, emptyValue)
	}
	p.Printf(", %s, %s", c.name, proc.ErrType)
	for _, msg := range proc.Messages {
		if msg != "" {
			p.Printf(", \"%s\"", msg)
		}
	}
	p.Printf(")\n")
}

type IsValidProc struct {
	baseCheckerProc
	call string
}

// isvalid:valid-function:CustomType:MsgA:MsgB....
func NewIsValidProc(vs []string) CheckerProc {
	p := &IsValidProc{}
	p.ErrType = "\"Invalid\""
	p.call = vs[1]

	if len(vs) > 2 && vs[2] != "" {
		p.ErrType = vs[2]
	}

	if len(vs) > 3 {
		p.Messages = vs[3:]
	}

	return p
}

func (cp *IsValidProc) Print(cm *Checker) {
	p := cm.p
	c := cm.c

	p.Printf("%s.Assert(%s(%s), %s, %s", c.chk, cp.call, c.expr, c.name, cp.ErrType)

	for _, msg := range cp.Messages {
		if msg != "" {
			p.Printf(", \"%s\"", msg)
		}
	}
	p.Printf(")\n")

}

type CallProc struct {
	baseCheckerProc
	call   string
	result string
}

// call:some-function:error/bool/true/false:CustomType:MsgA:MsgB:....
func NewCallProc(vs []string) CheckerProc {
	proc := &CallProc{}
	proc.call = vs[1]
	proc.result = "error"
	if len(vs) > 2 && vs[2] != "" {
		proc.result = vs[2]
	}
	proc.ErrType = "\"Invalid\""
	if len(vs) > 3 && vs[3] != "" {
		proc.ErrType = vs[3]
	}

	if len(vs) > 4 {
		proc.Messages = vs[4:]
	}

	return proc
}

func (cp *CallProc) Print(cm *Checker) {
	p := cm.p
	c := cm.c

	expr := c.expr
	for expr[0] == '*' {
		expr = expr[1:]
	}

	if cp.result == "error" || cp.result == "" {
		p.Printf("%s.AssertError(%s.%s(), %s, %s", c.chk, expr, cp.call, c.name, cp.ErrType)
	} else if cp.result == "bool" || cp.result == "true" {
		p.Printf("%s.Assert(%s.%s(), %s, %s", c.chk, expr, cp.call, c.name, cp.ErrType)
	} else if cp.result == "false" {
		p.Printf("%s.Assert(!%s.%s(), %s, %s", c.chk, expr, cp.call, c.name, cp.ErrType)
	} else {
		log.Fatalf("cannot generate call checker %s:", cp.call)
	}

	for _, msg := range cp.Messages {
		if msg != "" {
			p.Printf(", \"%s\"", msg)
		}
	}

	p.Printf(")\n")
}

type CompareProc struct {
	baseCheckerProc
	operand string
	value   string
}

// compare:!=><:value:CustomType:MsgA:MsgB...
func NewCompareProc(vs []string) CheckerProc {
	proc := &CompareProc{}
	proc.operand = vs[1]
	proc.value = strings.Replace(vs[2], "'", "\"", -1)
	proc.ErrType = "\"Invalid\""

	if len(vs) > 3 && vs[3] != "" {
		proc.ErrType = vs[3]
	}

	if len(vs) > 4 {
		proc.Messages = vs[4:]
	}

	return proc
}

func (cp *CompareProc) Print(cm *Checker) {
	p := cm.p
	c := cm.c

	p.Printf("%s.Assert(%s %s %s, %s, %s", c.chk, c.expr, cp.operand, cp.value, c.name, cp.ErrType)

	for _, msg := range cp.Messages {
		if msg != "" {
			p.Printf(", \"%s\"", msg)
		}
	}

	p.Printf(")\n")
}

type DefaultValueProc struct {
	defaultValue string
}

// default:somevalue
// default:'stringvalue'
func NewDefaultValueProc(vs []string) CheckerProc {
	proc := &DefaultValueProc{}
	s := strings.Replace(vs[1], "'", "\"", -1)
	proc.defaultValue = s
	return proc
}

func (cp *DefaultValueProc) Print(cm *Checker) {
	p := cm.p
	c := cm.c

	emptyValueExpr := parser.TypeZeroValue(c.targetType, c.file)
	emptyValue := getEmptyValueString(emptyValueExpr)

	if c.fullExpr != "" {
		p.Printf("if %s == %s {\n%s = %s\n}\n", c.expr, emptyValue, c.fullExpr, cp.defaultValue)
	} else {
		p.Printf("if %s == %s {\n%s = %s\n}\n", c.expr, emptyValue, c.expr, cp.defaultValue)
	}

}

type MergeProc struct {
	method string
}

// merge:other-valid-function
func NewMergeProc(vs []string) CheckerProc {
	proc := &MergeProc{}
	proc.method = vs[1]
	return proc
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

// arrays:123
func NewArrayProc(vs []string) CheckerProc {
	proc := &ArrayProc{}
	proc.arrays = getInt(vs[1])
	return proc
}

func (cp *ArrayProc) Print(cm *Checker) {
	if cp.arrays == 0 {
		return
	}

	p := cm.p
	c := cm.c

	expr := c.expr
	fullExpr := expr
	it := "i"
	idx := ""

	for i := 0; i < cp.arrays; i++ {
		it += "t"
		idx += "i"
		fullExpr = fmt.Sprintf("%s[%s]", expr, idx)
		p.Printf("for %s, %s := range %s {\n", idx, it, expr)
		expr = it
	}

	nc := c.Copy()
	nc.expr = it
	theName := c.name
	if theName[0] == '"' {
		var err error
		theName, err = strconv.Unquote(theName)
		if err != nil {
			log.Fatalf("unquote %s failed: %v", c.name, err)
		}
	}
	nc.name = "fmt.Sprintf(\"" + theName + ".%d\", " + idx + ")"
	nc.targetType = parser.TypeSkipBracket(c.targetType, cp.arrays)
	nc.fullExpr = fullExpr
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

// convert:convert-to-type:custom-new-name
func NewConvertProc(vs []string) CheckerProc {

	proc := &ConvertProc{}
	proc.convertTo = vs[1]
	if len(vs) > 2 {
		proc.convertName = vs[2]
	}

	return proc
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

// stars:123
func NewStarProc(vs []string) CheckerProc {
	p := &StarProc{}
	p.stars = getInt(vs[1])
	return p
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

	log.Println(string(bd.Bytes()))
	file.Add(bd)
}

func (g *Generator) Generate(c *Config) error {
	g.Prepare(".", nil, c.Output)
	g.PrepareImports(c.Imports)
	g.Run(c)
	g.Output(c.Output)
	return nil
}
