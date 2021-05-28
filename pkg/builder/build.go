package builder

import (
	"go/ast"
	"go/token"
	"log"

	"github.com/lawrsp/pigo/pkg/parser"
)

func getAllVariables(b Builder) *VariableList {
	vl := NewVariableList()
	//oder: inner to outer
	for ob := b; ob != nil; ob = ob.Outer() {
		vl = vl.Concat(ob.Variables())
	}

	return vl
}

func GetVariable(b Builder, t parser.Type, mode Mode, scope ScopeLevel) *Variable {
	return GetVariableWithSkip(b, t, mode, scope, 0)
}

func GetVariableWithSkip(b Builder, t parser.Type, mode Mode, scope ScopeLevel, skip int) *Variable {
	vl := NewVariableList()
	for x := Builder(b.Block()); x != nil; x = x.Outer() {
		vl = vl.Concat(x.Variables())
		if scope == Scope_Local {
			break
		}
		if scope == Scope_Function {
			if _, ok := x.(*FuncBuilder); ok {
				break
			}
		}
		if scope == Scope_File {
			if _, ok := x.(*FileBuilder); ok {
				break
			}
		}
	}

	for _, v := range vl.List {
		if t == nil {
			goto FINDED
		}
		if !v.Type.EqualTo(t) {
			continue
		}
		if mode == ANY_MODE {
			goto FINDED
		}
		if !v.mode.Contains(mode) {
			continue
		}

	FINDED:
		if skip == 0 {
			return v
		}

		skip--
		continue
	}

	return nil
}

func AddCallStmtWithIgnores(b Builder, params *VariableList, fnt parser.Type, ignores []int) *VariableList {
	stbd := NewCallStmt(b, fnt)
	results := stbd.Call(params, ignores)
	b.Block().Add(stbd)
	return results
}

func AddCallStmt(b Builder, params *VariableList, fnt parser.Type) *VariableList {
	stbd := NewCallStmt(b, fnt)
	results := stbd.Call(params, []int{})
	b.Block().Add(stbd)
	return results
}

func AddAppendStmt(b Builder, v *Variable, element *Variable) *Variable {
	callExpr := &ast.CallExpr{}
	callExpr.Fun = ast.NewIdent("append")
	callExpr.Args = []ast.Expr{v.Ident(), element.Ident()}

	return AddVariableAssign(b, v, callExpr)
}

func AddSuccessReturn(b Builder) {
	errType := parser.ErrorType()
	vl := getAllVariables(b.Block())

	filterd := NewVariableList()
	vl.ForEach(func(i int, v *Variable) {
		if v.Type.EqualTo(errType) {
			return
		}
		filterd.Add(v)
	})

	AddReturnStmt(b, filterd)
}

func AddErrorReturn(b Builder, ev *Variable) {
	vl := NewVariableList()
	vl.Add(ev)
	AddReturnStmt(b, vl)
}

func AddReturnStmt(b Builder, results *VariableList) {
	returnStmt := NewReturnStmt(b, results)
	b.Block().Add(returnStmt)
}

func AddCheckReturn(b Builder, cond ast.Expr, ev *Variable) {
	block := b.Block()
	builder := NewIfStmt(block)
	builder.SetInitCond(nil, cond)
	//
	var vl *VariableList = NewVariableList()
	if ev != nil {
		vl.Add(ev)
	}
	AddReturnStmt(builder, vl)
	b.Block().Add(builder)
}

//use exists variable assign to var
func AddAutoAssignExists(b Builder, lhs *Variable) {
	block := b.Block()
	vl := getAllVariables(block)
	valueType := lhs.Type
	rhs := vl.GetByType(valueType, READ_MODE)
	if rhs == nil {
		log.Printf("cannot find var of type %s", valueType)
		return
	}

	stmt := assignStmt(token.ASSIGN, lhs.Ident(), rhs.Ident())
	body := block.Body
	body.List = append(body.List, stmt)
}

//Variable Declaration
func AddVariableDecl(b Builder, v *Variable) *Variable {
	file := b.File()
	block := b.Block()
	vl := getAllVariables(block)
	for vl.Check(v.Name()) {
		v.IncreaseName()
	}

	spec := &ast.ValueSpec{}
	spec.Names = []*ast.Ident{ast.NewIdent(v.Name())}
	spec.Type = parser.TypeExprInFile(v.Type, file)

	decl := &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok:   token.VAR,
			Specs: []ast.Spec{spec},
		},
	}
	block.Body.List = append(block.Body.List, decl)
	block.addVariable(v)
	return v
}

//add assignment
func AddVariableAssign(b Builder, v *Variable, value ast.Expr) *Variable {
	block := b.Block()
	file := b.File()
	if value == nil {
		value = parser.TypeZeroValue(v.Type, file)
	}
	if v.IsAnonymous() {
		//X.y = ....
		stmt := assignStmt(token.ASSIGN, v.Ident(), value)
		body := block.Body
		body.List = append(body.List, stmt)
		v.SetMode(READ_MODE)
		return v
	} else {
		if v.IsVisible() {
			//old variable
			stmt := assignStmt(token.ASSIGN, v.Ident(), value)
			body := block.Body
			body.List = append(body.List, stmt)
			v.SetMode(READ_MODE)
		} else {
			variables := getAllVariables(block)
			// log.Printf("assign: %s", v.Type)
			// variables.debug()
			//new variable can assign a value or make ge declaration
			for variables.Check(v.Name()) || !block.addVariable(v) {
				v.IncreaseName()
			}

			stmt := assignStmt(token.DEFINE, v.Ident(), value)
			body := block.Body
			body.List = append(body.List, stmt)
			v.SetMode(READ_MODE)
		}
		return v
	}
}
