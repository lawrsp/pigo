package genrpc

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"

	"go/ast"
	"log"
	"os"
	"strings"

	"github.com/lawrsp/pigo/generator/builder"
	"github.com/lawrsp/pigo/generator/nameutil"
	"github.com/lawrsp/pigo/generator/parser"
	"golang.org/x/tools/imports"
)

type ImportLine struct {
	Name string
	Path string
}

func (l ImportLine) String() string {
	if l.Path == l.Name || strings.HasSuffix(l.Path, "/"+l.Name) {
		return fmt.Sprintf("\"%s\"", l.Path)
	}

	return fmt.Sprintf("%s \"%s\"", l.Name, l.Path)
}

type Generator struct {
	buf    bytes.Buffer
	parser *parser.Parser
	pkg    *parser.Package

	File *parser.File

	PackageName  string
	Imports      []ImportLine
	Receiver     *parser.DeclNode
	ReceiverType parser.Type

	InterfaceType parser.Type
	Functions     map[string]parser.Type

	ServiceReceiver     *parser.DeclNode
	ServiceReceiverType string
	ServicePackage      *parser.Package

	Taskes []*Task
}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

// Bytes return the generated bytes
func (g *Generator) Bytes() []byte {
	return g.buf.Bytes()
}

// format returns the gofmt-ed contents of the Generator's buffer.
func (g *Generator) Format() ([]byte, error) {
	options := &imports.Options{
		Fragment:  false,
		AllErrors: true,

		TabWidth:  2,
		TabIndent: true,
		Comments:  true,
	}
	res, err := imports.Process("", g.buf.Bytes(), options)
	// src, err := format.Source(g.buf.Bytes())
	if err != nil {
		// Should never happen, but can arise when developing this code.
		// The user can compile the output to see the error.
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return nil, err
	}
	return res, nil
}

func (g *Generator) addImportWithOutCheck(name string, path string) string {
	for _, impt := range g.Imports {
		if impt.Name == name {
			//TODO: same name different path
			return g.addImportWithOutCheck(nameutil.IncreaseName(name), path)
		}
	}

	g.Imports = append(g.Imports, ImportLine{Name: name, Path: path})
	return name
}

func (g *Generator) AddImport(name string, path string) string {
	//check already added
	for _, impt := range g.Imports {
		if impt.Path == path {
			return name
		}
	}

	return g.addImportWithOutCheck(name, path)
}

func (g *Generator) PrepareParser() {
	g.parser = parser.NewParser()
}

func (g *Generator) PrepareImports(imports map[string]string) {
	if imports != nil {
		for k, v := range imports {
			g.AddImport(k, v)
		}
	}
}

func (g *Generator) PreparePackage(dir string, files []string) {
	var p = g.parser
	var pkg *parser.Package
	if len(files) == 0 {
		pkg = p.ParsePackageDir(dir)
	} else {
		pkg = p.ParsePackageFiles(files)
	}

	// add custom imports
	if g.Imports != nil && len(g.Imports) > 0 {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "package %v\n", pkg.Name)
		fmt.Fprintf(&buf, "import (\n")
		for _, imptl := range g.Imports {
			fmt.Fprintf(&buf, "%s\n", imptl)
		}
		fmt.Fprintf(&buf, ")\n")
		// fmt.Println(string(buf.Bytes()))
		file := p.ParseFileContent("_custom", buf.Bytes())
		// printutil.PrintNodef(file, "_custom file is :")
		p.InsertFileToPackage(pkg, file, 0)
		g.File = file
	}

	g.pkg = pkg
	g.PackageName = pkg.Name
}

func (g *Generator) PrepareInterface(name string) {
	if name == "" {
		return
	}
	/*
		finder := parser.NewInterfaceTypeFinder()
		node := g.pkg.FindDecl(name, finder)
		if node == nil {
			log.Fatalf("cannot find interface %v", name)
		}
		// printutil.PrintNodef(node, "finded:")

		itft := parser.NewInterfaceType()
		itft.Fill(node)
	*/
	expr, err := parser.ParseExpr(name)
	if err != nil {
		log.Fatalf("interface %s defination error: %v", name, err)
	}

	_, typ := g.pkg.ReduceType(expr)
	if typ == nil {
		log.Fatalf("cannot reduce interface type: %s", name)
	}

	if _, ok := typ.Underlying().(*parser.InterfaceType); ok {
		g.InterfaceType = typ
	} else {
		log.Fatalf("%s is not interface type", name)
	}

	// printutil.PrintNodef(itft, "filled:")
}

func (g *Generator) PrepareReceiver(name string) {
	if name == "" {
		return
	}
	expr, err := parser.ParseExpr(name)
	if err != nil {
		log.Fatalf("receiver defination error: %s", name)
	}
	if _, g.ReceiverType = g.pkg.ReduceType(expr); g.ReceiverType == nil {
		log.Fatalf("cannot reduct receiver type: %s", name)
	}

	// g.pkg.Resolve(ReceiverType)

	log.Printf("finded reciver: %s", g.ReceiverType)

	/*
		finder := parser.NewStructTypeFinder()
		dn := g.pkg.FindDecl(name, finder)
		if dn == nil {
			log.Fatalf("cannot get receiver's type: %s", name)
		}

		g.Receiver = dn
	*/
}

func (g *Generator) PrepareFunctions(names map[string]string) {
	if len(names) == 0 {
		return
	}

	fns := map[string]parser.Type{}
	for name, def := range names {
		expr, err := parser.ParseExpr(def)
		if err != nil {
			log.Fatalf("function defination %s error: %v", def, err)
		}
		_, ft := g.pkg.ReduceType(expr)
		if ft == nil {
			log.Fatalf("cannot find function %v", name)
		}
		log.Printf("function %s", ft)
		fns[name] = parser.TypeWithExpr(ft, expr)

		/*
			node := g.pkg.FindFuncDecl(def)
			if node == nil {
				log.Fatalf("cannot find interface %v", name)
			}
			// printutil.PrintNodef(node, "finded:")
			fncT := parser.NewFuncType(def)
			fncT.Fill(node)
			fns[name] = fncT
		*/
	}

	g.Functions = fns
}

type ConverterProc struct {
	Source    string
	Target    string
	Converter *parser.FuncType
}

/*
func (g *Generator) PrepareConverters(names map[string]ConverterDesc) {

	results := map[string]*ConverterProc{}

	for name, desc := range names {
		assign := &ConverterProc{}
		assign.Source = desc.Source
		assign.Target = desc.Target
		convertName := desc.Converter
		if node := g.pkg.FindFuncDecl(convertName); node != nil {
			fnt := parser.NewFuncType(convertName)
			fnt.Fill(node)
			assign.Converter = fnt
		} else {
			log.Fatalf("cannot find convert function %s", convertName)
		}
		results[name] = assign
	}

	g.Converters = results
}
*/

func (g *Generator) PrepareService(name string) {

	pkg := g.pkg

	//1. find StructType
	//2. find ValueSpec
	finder := parser.NewStructTypeFinder()
	srv := pkg.FindDecl(name, finder)
	if srv != nil {
		g.ServiceReceiver = srv
		g.ServiceReceiverType = name
	} else {
		finder = parser.NewValueSpecFinder()
		srv = pkg.FindDecl(name, finder)
		if srv != nil {
			g.ServiceReceiver = srv
			typ := parser.ParseType(srv.Node)
			g.ServiceReceiverType = typ.String()
		}
	}
	if srv == nil {
		log.Fatalf("cannot find Service Recevier: %v", name)
	}

	g.ServicePackage = srv.File.BelongTo
}

type TaskProc struct {
	Checks  []string
	Call    parser.Type
	RCall   string
	Params  []ParamDesc
	Error   string
	Ignores []int
	Assigns []string
	Returns []string
}

type Task struct {
	Name         string
	RpcFunction  parser.Type
	ErrorWrapper parser.Type
	CheckParams  []CheckDesc
	Sequence     []*TaskProc
}

// func (g *Generator) getConverters(names []string) []*ConverterProc {
// 	assigns := []*ConverterProc{}
// 	if names == nil || len(names) == 0 {
// 		return assigns
// 	}

// 	for _, nm := range names {
// 		desc := g.Converters[nm]
// 		if desc == nil {
// 			log.Fatalf("assign  %s not defined", nm)
// 		}
// 		assigns = append(assigns, desc)
// 	}
// 	return assigns
// }

func (g *Generator) getFunction(name string) parser.Type {
	if name == "" {
		return nil
	}

	fn := g.Functions[name]
	if fn != nil {
		return fn
	}

	expr, err := parser.ParseExpr(name)
	if err != nil {
		log.Fatalf("function defination %s error: %v", name, err)
	}

	_, fnt := g.pkg.ReduceType(expr)
	if fnt == nil {
		log.Fatalf("function %s not found", name)
	}

	return parser.TypeWithExpr(fnt, expr)

	// node := g.pkg.FindDecl(name, parser.NewFuncDeclFinder(""))
	// if node != nil {
	// 	fnt := parser.NewFuncType(name)
	// 	fnt.Fill(node)
	// 	return fnt
	// }
	// return nil
}

func (g *Generator) addTask(name string, desc *TaskDesc) {

	for _, t := range g.Taskes {
		if t.Name == name {
			log.Fatalf("repeated function name")
		}
	}

	task := &Task{}
	task.Name = name

	//rpc function
	var rpcFuncName string
	if desc.Name == "" {
		rpcFuncName = name
	} else {
		rpcFuncName = desc.Name
	}
	log.Printf("interface: %s", g.InterfaceType)

	fnt := parser.GetInterfaceFuncByName(g.InterfaceType, rpcFuncName)
	if fnt == nil {
		log.Fatalf("cannot found rpc function definiation: %s", rpcFuncName)
	}
	task.RpcFunction = fnt

	//error-function
	if len(desc.ErrorWrapper) > 0 {
		fn := g.Functions[desc.ErrorWrapper]
		if fn == nil {
			log.Fatalf("error function not found %s", desc.ErrorWrapper)
		}
		task.ErrorWrapper = fn
	}

	//sequence calls:
	task.Sequence = []*TaskProc{}
	for _, seq := range desc.Sequence {
		proc := &TaskProc{}
		proc.Checks = seq.Checks
		proc.Error = seq.Error
		proc.Params = seq.Params
		proc.Call = g.getFunction(seq.Call)
		proc.RCall = seq.RCall
		proc.Returns = seq.Returns
		proc.Assigns = seq.Assigns
		proc.Ignores = seq.Ignores
		task.Sequence = append(task.Sequence, proc)
	}

	g.Taskes = append(g.Taskes, task)

}

func (g *Generator) PrepareTaskes(taskes map[string]TaskDesc) {
	for k, v := range taskes {
		name := k
		if len(v.Name) > 0 {
			name = v.Name
		}

		g.addTask(name, &v)
	}
}

type RCallDesc struct {
	Receiver string
	Func     string
	Params   []string
}

func NewRCallDsec(src string) *RCallDesc {

	var rcallRegexp = regexp.MustCompile("^(\\(([a-zA-Z0-9\\.\\*\\[\\]]*)\\))?([a-zA-Z0-9\\.]*)(\\(([a-zA-Z0-9,\\[\\]\\.\\*]*)\\))*$")

	result := rcallRegexp.ReplaceAllString(src, "$2,$3,$5")
	if result == src {
		return &RCallDesc{Func: src}
	}

	results := strings.Split(result, ",")
	desc := &RCallDesc{}
	desc.Receiver = results[0]
	desc.Func = results[1]
	desc.Params = results[2:]
	return desc
}

func (g *Generator) generateTask(outer builder.Builder, task *Task) {

	// bd.SetErrorWrapper(task.ErrorWrapper)
	log.Printf("do task %s: %s :%s", task.Name, g.ReceiverType, task.RpcFunction)
	bd := builder.NewFunction(outer, g.ReceiverType, task.RpcFunction, builder.NewErrorWrapper(task.ErrorWrapper))

	topNames := []interface{}{}
	for _, name := range bd.Variables().Names() {
		topNames = append(topNames, name)
	}

	for _, seq := range task.Sequence {
		var errExpr ast.Expr
		var err error
		if len(seq.Error) > 0 {
			errExpr, err = parser.ParseExpr(seq.Error)
			if err != nil {
				log.Fatalf("check param error defination %s error: \n %v", seq.Error, err)
			}
		}

		vl := builder.NewVariableList()
		namedParams := map[string]*builder.Variable{}
		if len(seq.Params) > 0 {
			for _, param := range seq.Params {
				exprSrc := param.Expr
				if exprSrc != "" && strings.Contains(exprSrc, "%[") {
					exprSrc = fmt.Sprintf(exprSrc, topNames...)
				}
				var t parser.Type
				if param.Type != "" {
					t = parser.ParseTypeString(param.Type)
				} else {
					t = parser.ParseTypeString(exprSrc)
				}
				if t == nil {
					log.Fatalf("cannot decied param type: %v", param)
				}

				v := builder.NewVariable(t).WithName(param.Name).WithMode(builder.READ_MODE)
				var valueExpr ast.Expr
				if valueExpr, err = parser.ParseExpr(exprSrc); err != nil {
					log.Fatalf("cannot parser value expr")
				}
				if v.Name() != "" {
					namedParams[param.Name] = builder.AddVariableAssign(bd, v, valueExpr)
				} else {
					v.SetExpr(valueExpr)
				}

				log.Printf("add variable %s", v.Type)
				vl.Add(v)
			}
		}

		if len(seq.Checks) > 0 {
			for _, chk := range seq.Checks {
				exprStr := fmt.Sprintf(chk, topNames...)
				expr, err := parser.ParseExpr(exprStr)
				if err != nil {
					log.Fatalf("check param defination %s error:\n %v", chk, err)
				}

				ev := builder.NewVariable(parser.ErrorType()).WithExpr(errExpr).ReadOnly()
				builder.AddCheckReturn(bd, expr, ev)
			}

		}

		//assign
		if len(seq.Assigns) > 0 {
			for _, asn := range seq.Assigns {
				var lhs, rhs string
				st := strings.Split(asn, "=")
				lhs = st[0]
				rhs = st[1]
				rhsType := parser.ParseTypeString(rhs)
				lhsVar := builder.NewVariable(rhsType)

				if strings.Contains(lhs, "%[") {
					lhs = fmt.Sprintf(lhs, topNames...)
				}

				for name, v := range namedParams {
					if strings.Contains(lhs, name) {
						lhs = strings.Replace(lhs, name, v.Name(), -1)
					}
				}

				lhsVar.SetExprSrc(lhs)
				builder.AddAutoAssignExists(bd, lhsVar)
				// builder.AddAssignStmt(bd, lhsVar)
			}
		}

		//call / rcall
		var callFunc parser.Type = nil
		if seq.Call != nil {
			callFunc = seq.Call
		} else if seq.RCall != "" {
			desc := NewRCallDsec(seq.RCall)
			var receiver *builder.Variable
			if desc.Receiver == "" {
				receiver = bd.Receiver()
			} else {
				t := g.File.ReduceTypeSrc(desc.Receiver)
				receiver = builder.GetVariable(bd.Block(), t, builder.READ_MODE, builder.Scope_Function)
			}

			rcallExpr := receiver.DotExpr(desc.Func)
			rcallType := receiver.DotTypeInFile(desc.Func, g.File)
			if rcallType == nil {
				log.Fatalf("cannot reduce type %s", desc.Func)
			}
			callFunc = parser.TypeWithExpr(rcallType, rcallExpr)
			log.Printf("call func: %s", rcallType)
		}

		if callFunc != nil {
			// log.Printf("!======ignores: %v", seq.Ignores)
			results := builder.AddCallStmtWithIgnores(bd, vl, callFunc, seq.Ignores)
			v := results.GetByType(parser.ErrorType(), builder.READ_MODE)
			if v != nil {
				if errExpr == nil {
					builder.AddCheckReturn(bd, v.CheckNilExpr(false), v)
				} else {
					nv := v.Copy().WithExpr(errExpr)
					builder.AddCheckReturn(bd, v.CheckNilExpr(false), nv)
				}
			}
		}
	}

	builder.AddSuccessReturn(bd)

	fmt.Println(string(bd.Bytes()))

	g.Printf("%s\n", string(bd.Bytes()))

	return
}

// generate produces the function
func (g *Generator) Run() {
	// Print the header and package clause.
	g.Printf("// Code generated by \"%s\"; DO NOT EDIT.\n", strings.Join(os.Args[0:], " "))
	g.Printf("\n")
	g.Printf("package %s \n", g.PackageName)
	g.Printf("\n")

	//import
	g.Printf("import (\n")

	for _, impt := range g.Imports {
		g.Printf("%s\n", impt)
	}

	g.Printf(")\n")

	g.Printf("\n")

	file := g.pkg.Files[0]
	fileBuilder := builder.NewFile(nil, file)
	for _, t := range g.Taskes {
		g.generateTask(fileBuilder, t)
	}
}

type ServiceDesc struct {
	Function string
}

type ConverterDesc struct {
	Source    string
	Target    string
	Converter string
}

type CheckDesc struct {
	Expr  string
	Error string
}

//[]CheckDesc
type ParamDesc struct {
	Name string
	Expr string
	Type string
}

type ProcDesc struct {
	Checks   []string
	Call     string
	RCall    string `yaml:"rcall"`
	Converts []string
	Error    string
	Params   []ParamDesc
	Assigns  []string
	Ignores  []int
	Returns  []string
}

type TaskDesc struct {
	Name         string
	Sequence     []ProcDesc
	ErrorWrapper string `yaml:"error_wrapper"`
}

type YamlConfig struct {
	Version   string
	Dir       string
	Files     []string
	Imports   map[string]string
	Output    string
	Receiver  string
	Interface string
	Generates map[string]TaskDesc
	Functions map[string]string
}

func (g *Generator) Generate(yamlConf *YamlConfig) error {

	g.PrepareParser()
	g.PrepareImports(yamlConf.Imports)
	g.PreparePackage(yamlConf.Dir, yamlConf.Files)
	g.PrepareInterface(yamlConf.Interface)
	g.PrepareReceiver(yamlConf.Receiver)
	g.PrepareFunctions(yamlConf.Functions)
	g.PrepareTaskes(yamlConf.Generates)
	g.Run()

	// fmt.Printf(string(g.Bytes()))

	// Format the output.
	result, err := g.Format()
	if err != nil {
		log.Fatalf("Foramt Failed: %v", err)
	}

	// Write to stdout / file
	if len(yamlConf.Output) == 0 {
		fmt.Printf(string(result))
	} else {
		err = ioutil.WriteFile(yamlConf.Output, result, 0644)
		if err != nil {
			log.Fatalf("writing output: %s", err)
		}
	}

	return nil
}
