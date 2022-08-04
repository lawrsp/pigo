package parser

import (
	"go/ast"
	"log"

	"github.com/lawrsp/pigo/generator/printutil"
)

type OldFuncType struct {
	Expr        ast.Expr
	Package     *Package
	Name        *ast.Ident
	OldFuncType *ast.FuncType
	Signature   string
}

func NewOldFuncType(name string) *OldFuncType {
	if expr, err := ParseExpr(name); err == nil {
		return &OldFuncType{Expr: expr}
	}

	log.Fatalf("func type name error:%s", name)
	return nil

}

func (ft *OldFuncType) Fill(node *DeclNode) {
	decl, ok := node.Node.(*ast.FuncDecl)
	if !ok {
		printutil.FatalNodef(node, "cannot fill OldFuncType:")
	}

	ft.Name = decl.Name
	if ft.Name == nil {
		ft.Name = node.Name
	}
	ft.Signature = ""
	ft.OldFuncType = decl.Type
}

func (ft *OldFuncType) copyFieldListToPackage(pkg *Package, src *ast.FieldList) *ast.FieldList {
	if src == nil {
		return nil
	}

	dstList := make([]*ast.Field, len(src.List))

	for i, field := range src.List {
		// Names   []*Ident  //copy
		// Type    Expr     //direct assgin
		var names []*ast.Ident
		for _, fdId := range field.Names {
			ident := &ast.Ident{}
			*ident = *fdId
			names = append(names, ident)
		}
		dstList[i] = &ast.Field{
			Names: names,
			Type:  field.Type,
		}
	}

	return &ast.FieldList{List: dstList}
}

func (ft *OldFuncType) CopyAstOldFuncType() *ast.FuncType {
	nft := &ast.FuncType{}
	nft.Params = ft.copyFieldListToPackage(nil, ft.OldFuncType.Params)
	nft.Results = ft.copyFieldListToPackage(nil, ft.OldFuncType.Results)
	return nft
}

func (ft *OldFuncType) CopyAstOldFuncTypeToPackage(pkg *Package) *ast.FuncType {
	nft := &ast.FuncType{}
	nft.Params = ft.copyFieldListToPackage(pkg, ft.OldFuncType.Params)
	nft.Results = ft.copyFieldListToPackage(pkg, ft.OldFuncType.Results)
	return nft
}

type OldInterface struct {
	OldFuncTypes []*OldFuncType
}

func NewOldInterface() *OldInterface {
	return &OldInterface{}
}

func (i *OldInterface) GetFuncByName(name string) *OldFuncType {
	for _, ft := range i.OldFuncTypes {
		if ft.Name.Name == name {
			return ft
		}
	}

	return nil
}

func (i *OldInterface) parseType(pkg *Package, name *ast.Ident, typ interface{}) []*OldFuncType {
	switch t := typ.(type) {
	case *ast.FuncType:
		return []*OldFuncType{
			&OldFuncType{
				Package:     pkg,
				Name:        name,
				OldFuncType: t,
				Signature:   "",
			},
		}
	case *ast.TypeSpec:
		return i.parseType(pkg, t.Name, t.Type)
	case *ast.InterfaceType:
		return i.parseFieldList(pkg, t.Methods)
	}

	return nil
}

func (i *OldInterface) parseField(pkg *Package, field *ast.Field) []*OldFuncType {
	switch t := field.Type.(type) {
	case *ast.Ident:
		return i.parseType(pkg, t, t.Obj.Decl)
	case *ast.FuncType:
		return i.parseType(pkg, field.Names[0], t)
	}

	return nil
}

func (i *OldInterface) parseFieldList(pkg *Package, list *ast.FieldList) []*OldFuncType {
	fnctyps := []*OldFuncType{}
	for _, field := range list.List {
		fnctyps = append(fnctyps, i.parseField(pkg, field)...)
	}
	return fnctyps
}

func (i *OldInterface) Fill(node *DeclNode) {
	itf, ok := node.Node.(*ast.InterfaceType)
	if !ok {
		printutil.FatalNodef(node.Node, "type Error:")
	}

	i.OldFuncTypes = i.parseFieldList(node.File.BelongTo, itf.Methods)
}
