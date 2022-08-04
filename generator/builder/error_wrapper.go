package builder

import (
	"go/ast"
	"log"

	"github.com/lawrsp/pigo/generator/parser"
)

type ErrorWrapper interface {
	WrapError(b Builder, ev *Variable) *VariableList
	Wrap(b Builder, ev *Variable) *Variable
}

type funcErrorWrapper struct {
	Type parser.Type
}

func (wrapper *funcErrorWrapper) WrapError(b Builder, ev *Variable) *VariableList {

	file := b.File()
	nev := ev.Copy()
	var function *FuncBuilder
	for x := b; function == nil; x = x.Outer() {
		function, _ = x.(*FuncBuilder)
	}

	args := NewVariableList()
	args.Add(nev)
	args = args.Concat(function.Variables())

	underFunc := wrapper.Type.Underlying().(*parser.FuncType)
	exprs, ok := exprListFromFields(underFunc.Params, args.Getter(READ_MODE))
	if !ok {
		log.Fatalf("cannot wrapp error with func %s(arguments not enough)", wrapper.Type)
	}
	expr := callExpr(parser.TypeExprInFile(wrapper.Type, file), exprs)
	nev.SetName("")
	nev.SetExpr(expr)

	vl := NewVariableList()
	vl.Add(nev)

	return vl
}
func (wrapper *funcErrorWrapper) Wrap(b Builder, ev *Variable) *Variable {

	file := b.File()
	nev := ev.Copy()
	var function *FuncBuilder
	for x := b; function == nil; x = x.Outer() {
		function, _ = x.(*FuncBuilder)
	}

	args := NewVariableList()
	args.Add(nev)
	args = args.Concat(function.Variables())

	underFunc := wrapper.Type.Underlying().(*parser.FuncType)
	exprs, ok := exprListFromFields(underFunc.Params, args.Getter(READ_MODE))
	if !ok {
		log.Fatalf("cannot wrapp error with func %s(arguments not enough)", wrapper.Type)
	}
	expr := callExpr(parser.TypeExprInFile(wrapper.Type, file), exprs)
	nev.SetName("")
	nev.SetExpr(expr)
	return nev
}

type exprErrorWrapper struct {
	expr ast.Expr
}

func (wrapper *exprErrorWrapper) WrapError(b Builder, ev *Variable) *VariableList {
	nev := ev.Copy()
	nev.SetExpr(wrapper.expr)
	nev.SetName("")
	vl := NewVariableList()
	vl.Add(nev)

	return vl
}
func (wrapper *exprErrorWrapper) Wrap(b Builder, ev *Variable) *Variable {
	nev := ev.Copy()
	nev.SetExpr(wrapper.expr)
	nev.SetName("")
	return nev
}

func NewErrorWrapper(input interface{}) ErrorWrapper {
	switch i := input.(type) {
	case parser.Type:
		if _, ok := i.Underlying().(*parser.FuncType); ok {
			return &funcErrorWrapper{i}
		}
	case ast.Expr:
		return &exprErrorWrapper{i}
	case nil:
		return nil
	default:
		log.Fatalf("doonot support input %v", i)
		return nil
	}
	log.Fatalf("doonot support input %v", input)
	return nil
}
