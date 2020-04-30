package parser

import (
	"go/ast"
	"go/token"
)

func checkNameInternal(name string) bool {
	switch name {
	case "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64":
	case "float32", "float64":
	case "complex64", "complex128":
	case "byte", "rune":
	case "string":
	case "bool":
	default:
		return false
	}
	return true
}

type DeclNode struct {
	Name  *ast.Ident
	Node  ast.Node
	Alias ast.Expr

	PreNode *DeclNode
	File    *File
}

func (node *DeclNode) WithPre(dn *DeclNode) *DeclNode {
	node.PreNode = dn
	return node
}
func (node *DeclNode) InFile(f *File) *DeclNode {
	node.File = f
	return node
}

type DeclFinder interface {
	Find(ast.Decl) *DeclNode
	WithName(name string) DeclFinder
}

type GenDeclFinder struct {
	Tok token.Token
}

func NewGenDeclFinder(tok token.Token) DeclFinder {
	return &GenDeclFinder{Tok: tok}
}
func (finder *GenDeclFinder) WithName(name string) DeclFinder {
	return &GenDeclFinder{}
}
func (finder *GenDeclFinder) Find(decl ast.Decl) *DeclNode {
	genDecl, ok := decl.(*ast.GenDecl)
	if !ok || genDecl.Tok != finder.Tok {
		// We only care about type spec
		return nil
	}
	return &DeclNode{Node: genDecl}
}

type ValueSpecFinder struct {
	parent DeclFinder
	Name   string
}

func NewValueSpecFinder() DeclFinder {
	return &ValueSpecFinder{parent: NewGenDeclFinder(token.VAR)}
}

func (finder *ValueSpecFinder) WithName(name string) DeclFinder {
	return &ValueSpecFinder{parent: finder.parent, Name: name}
}

func (finder *ValueSpecFinder) Find(decl ast.Decl) *DeclNode {
	node := finder.parent.Find(decl)
	if node == nil {
		return nil
	}

	genDecl := node.Node.(*ast.GenDecl)
	for _, spec := range genDecl.Specs {
		vspc := spec.(*ast.ValueSpec)
		for _, ident := range vspc.Names {
			if ident.Name == finder.Name {
				// utils.PrintNodef(vspc, "finded:")
				return &DeclNode{Name: ident, Node: vspc}
			}
		}
	}

	return nil
}

type TypeSpecFinder struct {
	parent DeclFinder
	Name   string
}

func NewTypeSpecFinder() DeclFinder {
	return &TypeSpecFinder{parent: NewGenDeclFinder(token.TYPE)}
}
func (finder *TypeSpecFinder) WithName(name string) DeclFinder {
	return &TypeSpecFinder{parent: finder.parent, Name: name}
}
func (finder *TypeSpecFinder) Find(decl ast.Decl) *DeclNode {
	node := finder.parent.Find(decl)
	if node == nil {
		return nil
	}

	genDecl := node.Node.(*ast.GenDecl)

	for _, spec := range genDecl.Specs {
		tspc := spec.(*ast.TypeSpec)
		ident := tspc.Name
		if ident.Name != finder.Name {
			continue
		}
		//type alias:
		if !tspc.Assign.IsValid() {
			return &DeclNode{Node: tspc, Name: ident}
		}
		return &DeclNode{Name: ident, Node: tspc, Alias: tspc.Type}
	}

	return nil
}

type StructTypeFinder struct {
	parent DeclFinder
}

func NewStructTypeFinder() DeclFinder {
	return &StructTypeFinder{parent: NewTypeSpecFinder()}
}

func (finder *StructTypeFinder) WithName(name string) DeclFinder {
	return &StructTypeFinder{parent: finder.parent.WithName(name)}
}

func (finder *StructTypeFinder) Find(decl ast.Decl) *DeclNode {
	node := finder.parent.Find(decl)
	if node == nil {
		return nil
	}

	if node.Alias != nil {
		return node
	}

	if t, ok := node.Node.(*ast.TypeSpec).Type.(*ast.StructType); ok {
		return &DeclNode{Name: node.Name, Node: t}
	}

	return nil
}

type InterfaceTypeFinder struct {
	parent DeclFinder
}

func NewInterfaceTypeFinder() DeclFinder {
	return &InterfaceTypeFinder{NewTypeSpecFinder()}
}

func (finder *InterfaceTypeFinder) WithName(name string) DeclFinder {
	return &InterfaceTypeFinder{parent: finder.parent.WithName(name)}
}

func (finder *InterfaceTypeFinder) Find(decl ast.Decl) *DeclNode {
	node := finder.parent.Find(decl)
	if node == nil {
		return nil
	}

	if node.Alias != nil {
		return node
	}

	if t, ok := node.Node.(*ast.TypeSpec).Type.(*ast.InterfaceType); ok {
		return &DeclNode{Name: node.Name, Node: t}
	}

	return nil
}

type FuncDeclFinder struct {
	Name     string
	Receiver string
}

func NewFuncDeclFinder(recvName string) DeclFinder {
	return &FuncDeclFinder{Receiver: recvName}
}
func (finder *FuncDeclFinder) WithName(name string) DeclFinder {
	return &FuncDeclFinder{Name: name, Receiver: finder.Receiver}
}

func getReceiverName(node ast.Node) string {
	switch t := node.(type) {
	case *ast.Field:
		return getReceiverName(t.Type)
	case *ast.StarExpr:
		return getReceiverName(t.X)
	case *ast.Ident:
		return t.Name
	}
	return ""
}
func (finder *FuncDeclFinder) Find(decl ast.Decl) *DeclNode {
	fnDecl, ok := decl.(*ast.FuncDecl)
	if !ok {
		// We only care about type spec
		return nil
	}

	if finder.Receiver == "" && fnDecl.Recv != nil {
		return nil
	}

	if finder.Receiver != "" {
		if fnDecl.Recv == nil || len(fnDecl.Recv.List) == 0 {
			return nil
		}
		name := getReceiverName(fnDecl.Recv.List[0])
		if name != finder.Receiver {
			// log.Printf("not equal %s != %s", name, finder.Receiver)
			return nil
		}
	}

	if fnDecl.Name.Name == finder.Name {
		return &DeclNode{Name: fnDecl.Name, Node: fnDecl}
	}

	return nil
}
