package parser

import (
	"go/ast"
	"go/token"
)

type DeclWalker interface {
	Walk(ast.Decl) bool
}

type GenDeclWalker struct {
	Tok  token.Token
	walk func(*ast.GenDecl) bool
}

func NewGenDeclWalker(tok token.Token, walk func(*ast.GenDecl) bool) DeclWalker {
	return &GenDeclWalker{Tok: tok, walk: walk}
}
func (walker *GenDeclWalker) Walk(decl ast.Decl) bool {
	genDecl, ok := decl.(*ast.GenDecl)
	if ok && genDecl.Tok == walker.Tok {
		return walker.walk(genDecl)
	}
	return true
}

type FuncDeclWalker struct {
	walk func(*ast.FuncDecl) bool
}

func NewFuncDeclWalker(walk func(*ast.FuncDecl) bool) DeclWalker {
	return &FuncDeclWalker{walk: walk}
}
func (walker *FuncDeclWalker) Walk(decl ast.Decl) bool {
	funcDecl, ok := decl.(*ast.FuncDecl)
	if ok {
		return walker.walk(funcDecl)
	}
	return true
}

func WalkFile(file *File, walker DeclWalker) {
	for _, decl := range file.File.Decls {
		if walker.Walk(decl) != true {
			return
		}
	}
}

func WalkPackage(pkg *Package, walker DeclWalker) {
	for _, file := range pkg.Files {
		for _, decl := range file.File.Decls {
			if walker.Walk(decl) != true {
				return
			}
		}
	}
}
