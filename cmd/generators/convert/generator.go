package convert

import (
	"fmt"

	"log"
	"strings"

	"go/ast"

	"github.com/lawrsp/pigo/pkg/builder"
	"github.com/lawrsp/pigo/pkg/generator"
	"github.com/lawrsp/pigo/pkg/parser"
)

type CustomAssign struct {
	Source parser.Type
	Target parser.Type
	Assign parser.Type
	Check  string
}

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

type Type struct {
	parser.Type
	PackageName string
	Name        string
	Stars       int
}

func NewType(t parser.Type) *Type {
	return &Type{
		Type:  t,
		Name:  t.Name(),
		Stars: parser.GetTypeStars(t),
	}
}

func (t *Type) FullName() string {
	if t.Type != nil {
		return t.Type.String()
	}
	name := ""
	if t.PackageName == "" {
		name = t.Name
	} else {
		name = fmt.Sprintf("%s.%s", t.PackageName, t.Name)
	}

	if t.Stars > 0 {
		name = strings.Repeat("*", t.Stars) + name
	}
	return name
}

func (t *Type) ElemName() string {
	name := ""
	if t.PackageName == "" {
		name = t.Name
	} else {
		name = fmt.Sprintf("%s.%s", t.PackageName, t.Name)
	}
	return name
}

type StructInfo struct {
	*Type
	Origin     string
	expr       ast.Expr
	importPath string
	st         *ast.StructType
	file       *parser.File
	pkg        *parser.Package
	fields     []*Field
}

func (si *StructInfo) Print() {
	for _, f := range si.fields {
		fmt.Println(f)
	}
}

func NewStructInfo(origin string) *StructInfo {
	st := &StructInfo{Origin: origin}
	//fs := token.NewFileSet()
	expr, err := parser.ParseExpr(origin)
	if err != nil {
		log.Fatalf("parsing type: %s: %s", origin, err)
		return nil
	}

	//_ = ast.Print(fs, expr)
	//@CHECK:
	st.Type = NewType(parser.ParseType(expr))
	st.expr = expr

	return st
}

type Field struct {
	Name string
	//@CHECK:
	Type *Type

	file  *parser.File
	field *ast.Field
}

type genTask struct {
	Source   parser.Type
	Target   parser.Type
	FuncType parser.Type

	Depend      string
	SourceError ast.Expr
}

type Generator struct {
	generator.Generator

	TagName           string
	Imports           []ImportLine
	CustomAssigns     []*CustomAssign
	Resolves          []string
	IgnoreImportPaths []string

	taskes map[string]*genTask //naem : task

	packageName string
}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) PrepareImports(imports map[string]string) {
	var result []ImportLine
	if imports != nil {
		for k, v := range imports {
			result = append(result, ImportLine{k, v})
		}
	}
	g.Imports = result

	bd := builder.NewFile(nil, g.File)

	if g.Imports != nil && len(g.Imports) > 0 {
		for _, imptl := range g.Imports {
			bd.AddImport(imptl.Name, imptl.Path)
		}
	}
}

func (g *Generator) PrepareAssigns(assignConf map[string]YamlCustomAssign) {
	assigns := []*CustomAssign{}
	if assignConf != nil {
		for _, v := range assignConf {
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
	}

	g.CustomAssigns = assigns
}

func (g *Generator) PrepareTaskes(taskConf map[string]YamlTaskElem) {
	taskes := map[string]*genTask{}
	sorted := []string{}
	needSort := [][2]string{}
	for k, v := range taskConf {
		if v.Depend == "" {
			sorted = append(sorted, k)
		} else {
			a := [2]string{k, v.Depend}
			needSort = append(needSort, a)
		}
	}

	for len(needSort) > 0 {
		left := [][2]string{}
		for i := 0; i < len(needSort); i++ {
			for j := 0; j < len(needSort); j++ {
				if needSort[i][1] == needSort[j][0] {
					left = append(left, needSort[i])
				} else {
					sorted = append(sorted, needSort[i][0])
				}
			}
		}
		needSort = left
	}

	for _, k := range sorted {
		name := k
		v := taskConf[k]
		if len(v.Name) > 0 {
			name = v.Name
		}

		task := g.newTask(name, v.Source, v.Target, v.Depend, v.SourceError, v.WithoutError)
		taskes[k] = task
	}
	g.taskes = taskes
}

/*	g.PrepareInterface(yamlConf.Interface)
	g.PrepareReceiver(yamlConf.Receiver)
	g.PrepareFunctions(yamlConf.Functions)
	g.PrepareTaskes(yamlConf.Generates)
	g.Run()
*/
func increaseName(name string) string {
	len := len(name)
	suffix := make([]byte, len)
	nameBytes := []byte(name)

	end := len - 1
	for ; end > 0; end -= 1 {
		if nameBytes[end] >= '0' && nameBytes[end] <= '9' {
			suffix[end] = nameBytes[end]
		} else {
			break
		}
	}

	num := 0
	for idx := end + 1; idx < len; idx += 1 {
		num = num*10 + int((suffix[idx] - '0'))
	}
	num += 1

	return fmt.Sprintf("%s%d", nameBytes[:end+1], num)
}

func (g *Generator) addImportWithOutCheck(name string, path string) string {
	for _, impt := range g.Imports {
		if impt.Name == name {
			//TODO: same name different path
			return g.addImportWithOutCheck(increaseName(name), path)
		}
	}

	g.Imports = append(g.Imports, ImportLine{Name: name, Path: path})
	return name
}

func (g *Generator) AddImport(name string, path string) string {
	//check ignore
	if g.IgnoreImportPaths != nil {
		for _, ignore := range g.IgnoreImportPaths {
			if ignore == path {
				return name
			}
		}
	}

	//check already added
	for _, impt := range g.Imports {
		if impt.Path == path {
			return name
		}
	}

	return g.addImportWithOutCheck(name, path)
}

func (g *Generator) newTask(funcName string, src string, dst string, depend string, sourceError string, withoutError bool) *genTask {
	task := genTask{}

	task.Source = g.ReduceTypeSrc(src)
	task.Target = g.ReduceTypeSrc(dst)
	fnt := &parser.FuncType{
		Params: []*parser.Field{
			parser.NewField(task.Source, "src", ""),
		},
		Results: []*parser.Field{
			parser.NewField(task.Target, "dst", ""),
		},
	}
	if !withoutError {
		fnt.Results = append(fnt.Results, parser.NewField(parser.ErrorType(), "err", ""))
	}
	task.FuncType = parser.TypeWithFile(parser.TypeWithName(fnt, funcName), g.File)
	task.Depend = depend
	if sourceError != "" {
		expr, err := parser.ParseExpr(sourceError)
		if err != nil {
			log.Fatalf("source_error(%s) defined error: %v", sourceError, err)
		}
		task.SourceError = expr
	}

	return &task
}

type TaskGenerator struct {
	Parent      *Generator
	Tpaths      []*parser.TPath
	FuncBuilder builder.Builder
	Task        *genTask
	Error       ast.Expr
}

func (tg *TaskGenerator) Run() bool {
	task := tg.Task

	g := tg.Parent
	fb := tg.FuncBuilder

	srcV := builder.GetVariable(fb, task.Source, builder.READ_MODE, builder.Scope_Function)
	dstV := builder.NewVariable(task.Target).WithName("dst").WriteOnly()

	if _, ok := dstV.Type.Underlying().(*parser.ArrayType); ok {
		valueExpr := parser.TypeInitValue(dstV.Type, fb.File())
		dstV = builder.AddVariableAssign(fb, dstV, valueExpr)
	}

	var ev *builder.Variable
	if tg.Error != nil {
		ev = builder.NewVariable(parser.ErrorType()).WithExpr(tg.Error).ReadOnly()
	}
	structAssign := builder.NewStructAssign(fb, g.TagName, ev, tg.Tpaths)
	if ok := structAssign.TryAssign(srcV, dstV); ok {
		fb.Block().Add(structAssign)
		return true
	}

	if ok := builder.TryDirectAssign(fb, srcV, dstV, ev, tg.Tpaths); ok {
		return true
	}

	return false
}

func (g *Generator) generateTask(outer builder.Builder, task *genTask) builder.Builder {
	// log.Printf("%s", task.src.Type)
	fb := builder.NewFunction(outer, nil, task.FuncType, nil)

	tpaths := []*parser.TPath{}
	for _, assign := range g.CustomAssigns {
		if _, ok := assign.Assign.Underlying().(*parser.FuncType); ok {
			tpaths = append(tpaths, parser.NewTPath(assign.Source, assign.Target).WithFunction(assign.Assign))
		} else {
			tpaths = append(tpaths, parser.NewTPath(assign.Source, assign.Target).WithTypeConversion(assign.Assign))
		}
	}

	if task.Depend != "" {
		if depended := g.taskes[task.Depend]; depended != nil {
			tp := parser.NewTPath(depended.Source, depended.Target).WithFunction(depended.FuncType)
			tpaths = append(tpaths, tp)
		} else {
			log.Fatalf("cannot find depended %s", task.Depend)
		}
	}

	taskGen := &TaskGenerator{
		Parent:      g,
		Tpaths:      tpaths,
		FuncBuilder: fb,
		Task:        task,
		Error:       task.SourceError,
	}

	if taskGen.Run() {
		builder.AddSuccessReturn(fb)
		return fb
	}

	log.Fatalf("cannot support %s to %s assign", task.Source, task.Target)
	return nil
}

// generate produces the function
func (g *Generator) Run() {
	// Print the header and package clause.
	/*
		g.Printf("// Code generated by \"%s\"; DO NOT EDIT.\n", strings.Join(os.Args[0:], " "))
		g.Printf("\n")
		g.Printf("package %s \n", g.packageName)
		g.Printf("\n")

		//import
		g.Printf("import (\n")
		for _, impt := range g.Imports {
			g.Printf("%s\n", impt)
		}

		g.Printf(")\n")

		g.Printf("\n")
	*/

	/*
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "package %v\n", pkg.Name)
		// add custom imports
		if g.Imports != nil && len(g.Imports) > 0 {
			fmt.Fprintf(&buf, "import (\n")
			for _, imptl := range g.Imports {
				fmt.Fprintf(&buf, "%s\n", imptl)
			}
			fmt.Fprintf(&buf, ")\n")
		}*/
	bd := builder.NewFile(nil, g.File)

	if g.Imports != nil && len(g.Imports) > 0 {
		for _, imptl := range g.Imports {
			bd.AddImport(imptl.Name, imptl.Path)
		}
	}

	for _, t := range g.taskes {
		bd.Add(g.generateTask(bd, t))
	}
}

type YamlCustomAssign struct {
	Source string
	Target string
	Assign string
	Check  string
}

type YamlTaskElem struct {
	Name         string
	Source       string
	Target       string
	Depend       string
	SourceError  string `yaml:"source_error"`
	WithoutError bool   `yaml:"without_error"`
}

type YamlConfig struct {
	Version   string
	TagName   string
	Dir       string
	Files     []string
	Imports   map[string]string
	Output    string
	Assigns   map[string]YamlCustomAssign
	Generates map[string]YamlTaskElem
}

func (g *Generator) Generate(yamlConf *YamlConfig) error {
	if yamlConf.TagName != "" {
		g.TagName = yamlConf.TagName
	} else {
		g.TagName = "pc"
	}

	g.Prepare(yamlConf.Dir, yamlConf.Files, yamlConf.Output)
	g.PrepareImports(yamlConf.Imports)
	g.PrepareAssigns(yamlConf.Assigns)
	g.PrepareTaskes(yamlConf.Generates)
	g.Run()

	g.Output(yamlConf.Output)

	/*
		fmt.Printf(string(g.Bytes()))
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
	*/

	return nil
}
