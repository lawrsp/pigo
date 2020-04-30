package builder

import (
	"go/ast"
	"go/token"

	"github.com/lawrsp/pigo/pkg/parser"
	"github.com/lawrsp/pigo/pkg/printutil"
)

func callExpr(fun ast.Expr, args []ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  fun,
		Args: args,
	}
}

func exprListFromFields(fields []*parser.Field, getter VarGetter) ([]ast.Expr, bool) {
	results := []ast.Expr{}
	for i, item := range fields {
		v := getter.Get(i, item.Type)
		if v == nil {
			// log.Printf("argument with type(%s) not defined", item.Type)
			return nil, false
		}
		results = append(results, v.Ident())
	}
	return results, true
}

func LenExpr(a ast.Expr) ast.Expr {
	callExpr := &ast.CallExpr{}
	callExpr.Fun = ast.NewIdent("len")
	callExpr.Args = []ast.Expr{a}
	return callExpr
}

func TypeConversionExpr(typeExpr ast.Expr, val ast.Expr) ast.Expr {
	callExpr := &ast.CallExpr{}
	callExpr.Fun = typeExpr
	callExpr.Args = []ast.Expr{val}
	return callExpr
}

func AppendExpr(list ast.Expr, item ast.Expr) ast.Expr {
	callExpr := &ast.CallExpr{}
	callExpr.Fun = ast.NewIdent("append")
	callExpr.Args = []ast.Expr{list, item}
	return callExpr
}

func condExpr(op token.Token, exprs ...ast.Expr) ast.Expr {

	exprLen := len(exprs)

	if exprLen == 1 {
		return &ast.ParenExpr{X: exprs[0]}
	}

	result := &ast.BinaryExpr{
		Op: op,
	}

	for i, expr := range exprs {
		if i == 0 {
			result.X = &ast.ParenExpr{X: expr}
		} else if i == 1 {
			result.Y = &ast.ParenExpr{X: expr}
		} else {
			result = &ast.BinaryExpr{
				X:  result,
				Op: op,
				Y:  &ast.ParenExpr{X: expr},
			}
		}
	}

	return result
}

func CondExpr(op token.Token, exprs ...ast.Expr) ast.Expr {
	return condExpr(op, exprs...)
}

func DotExpr(x ast.Expr, dot ast.Expr) ast.Expr {
	switch expr := dot.(type) {
	case *ast.Ident:
		return &ast.SelectorExpr{X: x, Sel: expr}
	case *ast.SelectorExpr:
		expr.X = DotExpr(x, expr.X)
		return expr
	case *ast.CallExpr:
		expr.Fun = DotExpr(x, expr.Fun)
		return expr
	default:
		printutil.PrintNodef(expr, "expr:")
	}

	return nil
}
