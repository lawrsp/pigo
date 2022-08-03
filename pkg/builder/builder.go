package builder

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"log"
	"strconv"
	"strings"

	"github.com/lawrsp/pigo/pkg/parser"
	"github.com/lawrsp/pigo/pkg/printutil"
)

type ScopeLevel int

const (
	Scope_Local ScopeLevel = iota
	Scope_Function
	Scope_File
)

type builderInternal interface {
	Package() *parser.Package
	File() *parser.File
	addVariable(*Variable) bool
}

type Builder interface {
	builderInternal
	Outer() Builder
	Variables() *VariableList
	Block() *BlockBuilder
	Add(Builder)
	Bytes() []byte
	String() string
}

type baseBuilder struct {
	outer     Builder
	variables *VariableList
	pkg       *parser.Package
	file      *parser.File
}

func newBaseFromOuter(outer Builder) *baseBuilder {
	pkg := outer.Package()
	file := outer.File()
	return &baseBuilder{outer, NewVariableList(), pkg, file}
}

func newbaseBuilder(outer Builder, pkg *parser.Package, file *parser.File) *baseBuilder {
	return &baseBuilder{outer, NewVariableList(), pkg, file}
}
func (b *baseBuilder) Package() *parser.Package {
	return b.pkg
}
func (b *baseBuilder) File() *parser.File {
	return b.file
}
func (b *baseBuilder) addVariable(v *Variable) bool {
	if v.isVisible {
		log.Fatalf("add visible variable: %s(%s)", v.Name(), v.Type)
	}
	if ok := b.variables.Insert(v, 0); !ok {
		return false
	}
	v.isVisible = true
	return true
}

func (b *baseBuilder) String() string {
	if b.outer == nil {
		return "base"
	}
	return b.outer.String()
}
func (b *baseBuilder) Outer() Builder {
	return b.outer
}
func (b *baseBuilder) Variables() *VariableList {
	return b.variables
}
func (b *baseBuilder) Add(bd Builder) {
	log.Fatalf("not implemented Add")
}
func (b *baseBuilder) Block() *BlockBuilder {
	log.Fatalf("not implemented Block()")
	return nil
}
func (b *baseBuilder) Bytes() []byte {
	log.Fatalf("not implemented Bytes()")
	return nil
}

type FileBuilder struct {
	*baseBuilder
	file *ast.File
}

func NewFile(outer Builder, file *parser.File) *FileBuilder {
	var pkg *parser.Package
	if file != nil {
		pkg = file.BelongTo
	} else if outer != nil {
		pkg = outer.Package()
	}

	if file != nil && file.BelongTo != nil && file.BelongTo != pkg {
		log.Fatalf("cannot create builder with different package")
	}

	base := newbaseBuilder(outer, pkg, file)
	return &FileBuilder{baseBuilder: base, file: file.File}
}
func (b *FileBuilder) String() string {
	return fmt.Sprintf("%s.file", b.baseBuilder.String())
}
func (b *FileBuilder) Block() *BlockBuilder {
	return nil
}
func (b *FileBuilder) Bytes() []byte {
	return nil
}

func (b *FileBuilder) AddImport(name, path string) {
	for _, spec := range b.file.Imports {
		oldPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			log.Fatalf("get old path error:%v", err)
		}
		var oldName string
		if spec.Name != nil {
			oldName = spec.Name.Name
		} else {
			names := strings.Split(oldPath, "/")
			oldName = names[len(names)-1]
		}

		if oldName == name {
			if oldPath != path {
				log.Fatalf("cannot add import %s with same name and different path:\n%s\n%s ", name, oldPath, path)
			}
			return
		}
	}

	pathNames := strings.Split(path, "/")
	nameFromPath := pathNames[len(pathNames)-1]

	importSpec := &ast.ImportSpec{
		Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("\"%s\"", path)},
	}
	if nameFromPath != name {
		importSpec.Name = ast.NewIdent(name)
	}

	b.file.Imports = append(b.file.Imports, importSpec)
	for _, decl := range b.file.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok {
			if gen.Tok == token.IMPORT {
				gen.Specs = append(gen.Specs, importSpec)
				return
			}
		}
	}

	gen := &ast.GenDecl{
		Tok:   token.IMPORT,
		Specs: []ast.Spec{importSpec},
	}
	b.file.Decls = append([]ast.Decl{gen}, b.file.Decls...)
}

func getType(e ast.Expr) (int, string) {
	switch me := e.(type) {
	case *ast.Ident:
		return 0, me.Name
	case *ast.SelectorExpr:
		_, name := getType(me.X)
		return 1, fmt.Sprintf("%s.%s", name, me.Sel.Name)
	case *ast.StarExpr:
		_, name := getType(me.X)
		return 2, name
	default:
		printutil.PrintNodef(me, "")
		log.Fatalf("not supported")

	}
	return -1, ""
}

func exprEqual(x ast.Expr, y ast.Expr) bool {
	xt, xn := getType(x)
	yt, yn := getType(y)
	return xt == yt && xn == yn
}

func funcDeclEqual(x *ast.FuncDecl, y *ast.FuncDecl) bool {
	if x.Name == nil || y.Name == nil {
		return false
	}

	if x.Name.Name != y.Name.Name {
		return false
	}
	if x.Recv == nil && y.Recv == nil {
		return true
	}
	if x.Recv == nil || y.Recv == nil {
		return false
	}
	if len(x.Recv.List) == 0 && len(y.Recv.List) == 0 {
		return true
	}
	if len(x.Recv.List) == 0 || len(y.Recv.List) == 0 {
		return false
	}

	xR := x.Recv.List[0]
	yR := y.Recv.List[0]

	// log.Printf("%s == %s ? %v==============", xR.Names[0].Name, yR.Names[0].Name, xR.Names[0].Name == yR.Names[0].Name)

	return xR.Names[0].Name == yR.Names[0].Name && exprEqual(xR.Type, yR.Type)

}

func (b *FileBuilder) AddFuncDecl(fd *ast.FuncDecl) {
	for i, decl := range b.file.Decls {
		//replace old
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if funcDeclEqual(funcDecl, fd) {
				b.file.Decls[i] = fd
				return
			}
		}
	}

	//add new
	b.file.Decls = append(b.file.Decls, fd)
}

func (b *FileBuilder) Add(ab Builder) {
	switch x := ab.(type) {
	case *FuncBuilder:
		b.AddFuncDecl(x.Decl)
	case *FuncBufferBuilder:
		fd := x.Decl()
		b.AddFuncDecl(fd)
	default:
		log.Fatalf("not supported %s", ab.String())
	}
}

type BlockBuilder struct {
	*baseBuilder
	Body *ast.BlockStmt
}

func (b *BlockBuilder) String() string {
	return fmt.Sprintf("%s.block", b.baseBuilder.String())
}
func (b *BlockBuilder) Block() *BlockBuilder {
	return b
}
func (b *BlockBuilder) Bytes() []byte {
	var buf bytes.Buffer
	if err := format.Node(&buf, token.NewFileSet(), b.Body); err != nil {
		log.Fatalf("generate code error: %v", err)
	}

	return buf.Bytes()
}
func (b *BlockBuilder) Add(ab Builder) {
	switch x := ab.(type) {
	case StmtBuilder:
		b.Body.List = append(b.Body.List, x.Stmt())
		b.variables = x.Variables().Concat(b.variables)
	case StmtListBuilder:
		b.Body.List = append(b.Body.List, x.StmtList()...)
		b.variables = x.Variables().Concat(b.variables)
	}
}
func (b *BlockBuilder) StmtList() []ast.Stmt {
	return b.Body.List
}

func newBlockBuilder(outer Builder, block *ast.BlockStmt) *BlockBuilder {
	pkg := outer.Package()
	file := outer.File()
	return &BlockBuilder{newbaseBuilder(outer, pkg, file), block}
}

type FuncBuilder struct {
	*baseBuilder
	Decl        *ast.FuncDecl
	receiver    *Variable
	funcType    *parser.FuncType
	block       *BlockBuilder
	returnError ErrorWrapper
}

func (b *FuncBuilder) String() string {
	return fmt.Sprintf("%s.func", b.baseBuilder.String())
}
func (b *FuncBuilder) Block() *BlockBuilder {
	return b.block
}
func (b *FuncBuilder) Bytes() []byte {
	var buf bytes.Buffer
	if err := format.Node(&buf, token.NewFileSet(), b.Decl); err != nil {
		log.Fatalf("generate code error: %v", err)
	}
	return buf.Bytes()
}

func (b *FuncBuilder) Receiver() *Variable {
	return b.receiver
}

func (b *FuncBuilder) HasResults() bool {
	return len(b.funcType.Results) > 0
}

func NewFunction(outer Builder, receiver parser.Type, fnt parser.Type, ew ErrorWrapper) *FuncBuilder {
	pkg := outer.Package()
	file := outer.File()
	allVars := getAllVariables(outer)

	if fnt.Name() == "" {
		log.Fatalf("func type without any name: %s", fnt)
	}

	decl := &ast.FuncDecl{}
	decl.Name = ast.NewIdent(fnt.Name())

	//receiver, params should as it order
	//so donnot use base.addVariable as it use reverse order
	base := newbaseBuilder(outer, pkg, file)
	builder := &FuncBuilder{baseBuilder: base}

	underFnt, ok := fnt.Underlying().(*parser.FuncType)
	if !ok {
		log.Fatalf("the type is not a funcType: %s", fnt)
	}

	if receiver == nil && underFnt.Receiver != nil {
		receiver = underFnt.Receiver
	}

	if receiver != nil {
		var v *Variable
		if fd, ok := receiver.(*parser.Field); ok {
			log.Printf("fieldName: %s", fd.Name())
			receiver = fd.Type
			v = NewVariable(receiver).WithName(fd.Name()).ReadOnly()
		} else {
			v = NewVariable(receiver).AutoName().ReadOnly()
		}

		for allVars.Check(v.Name()) || !base.variables.Add(v) {
			v.IncreaseName()
		}
		v.isVisible = true
		builder.receiver = v

		field := &ast.Field{}
		ident := &ast.Ident{Name: v.Name()}
		field.Names = []*ast.Ident{ident}
		field.Type = parser.TypeExprInFile(receiver, file)
		fieldList := &ast.FieldList{List: []*ast.Field{field}}
		decl.Recv = fieldList

	}

	funcType := parser.TypeExprInFile(underFnt, file).(*ast.FuncType)

	if funcType.Params != nil && funcType.Params.List != nil && len(funcType.Params.List) > 0 {
		for i, param := range funcType.Params.List {
			field := underFnt.Params[i]
			// log.Printf("params %d type: %s", i, t)
			// _, t := pkg.ReduceType(underFnt.Params[i].Type.Expr())
			v := NewVariable(field.Type).WithName(field.FieldName()).AutoName().ReadOnly()
			for allVars.Check(v.Name()) || !base.variables.Add(v) {
				v.IncreaseName()
			}
			v.isVisible = true

			// log.Printf("add variable: %s : Type(%s)mode(%d)(field: %s)", v.Name(), v.Type, v.mode, field.Name())
			ident := ast.NewIdent(v.Name())
			param.Names = []*ast.Ident{ident}
		}
	}

	if funcType.Results != nil && funcType.Results.List != nil && len(funcType.Results.List) > 0 {
		for _, res := range funcType.Results.List {
			res.Names = nil
		}
	}

	decl.Type = funcType
	decl.Body = &ast.BlockStmt{}

	builder.Decl = decl
	builder.funcType = underFnt
	builder.returnError = ew

	block := newBlockBuilder(builder, decl.Body)
	builder.block = block

	return builder
}

func (b *FuncBuilder) Add(ab Builder) {
	b.block.Add(ab)
}

func (b *FuncBuilder) Variables() *VariableList {
	if b.block == nil {
		return b.baseBuilder.Variables()
	}
	return b.block.Variables().Concat(b.baseBuilder.Variables())
}
