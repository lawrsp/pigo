package builder

import (
	"go/ast"
	"go/token"
)

//basic

func assignStmt(tok token.Token, lhs ast.Expr, rhs ast.Expr) *ast.AssignStmt {
	stmt := &ast.AssignStmt{}
	stmt.Lhs = []ast.Expr{lhs}
	stmt.Tok = tok
	stmt.Rhs = []ast.Expr{rhs}
	return stmt
}

func multiAssginStmt(tok token.Token, lhs []ast.Expr, rhs ast.Expr) *ast.AssignStmt {
	stmt := &ast.AssignStmt{}
	stmt.Lhs = lhs
	stmt.Tok = tok
	stmt.Rhs = []ast.Expr{rhs}
	return stmt
}

func rangeStmt(key ast.Expr, value ast.Expr, tok token.Token, x ast.Expr) *ast.RangeStmt {
	stmt := &ast.RangeStmt{
		Key:   key,
		Value: value,
		Tok:   tok,
		X:     x,
		Body:  &ast.BlockStmt{},
	}
	return stmt
}

func ifStmt(init ast.Stmt, cond ast.Expr) *ast.IfStmt {
	ifStmt := &ast.IfStmt{}
	ifStmt.Body = &ast.BlockStmt{}
	ifStmt.Cond = cond
	ifStmt.Init = init
	return ifStmt
}

func returnStmt(results []ast.Expr) *ast.ReturnStmt {
	return &ast.ReturnStmt{
		Results: results,
	}
}

//with variable
