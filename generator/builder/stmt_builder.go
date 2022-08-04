package builder

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
	"log"

	"github.com/lawrsp/pigo/generator/parser"
)

type StmtBuilder interface {
	Builder
	Stmt() ast.Stmt
}

type ForRangeBuilder struct {
	*baseBuilder
	block *BlockBuilder
	stmt  *ast.RangeStmt
}

func (b *ForRangeBuilder) Block() *BlockBuilder {
	return b.block
}
func (b *ForRangeBuilder) Bytes() []byte {
	var buf bytes.Buffer
	if err := format.Node(&buf, token.NewFileSet(), b.block); err != nil {
		log.Fatalf("generate code error: %v", err)
	}
	return buf.Bytes()
}
func (b *ForRangeBuilder) Stmt() ast.Stmt {
	return b.stmt
}

func NewForRange(outer Builder, key *Variable, value *Variable, over *Variable) *ForRangeBuilder {
	pkg := outer.Package()
	file := outer.File()
	base := newbaseBuilder(outer.Block(), pkg, file)
	builder := &ForRangeBuilder{base, nil, nil}

	//check key, value exists:
	var isNewKey = false
	var isNewValue = false
	if key != nil && !key.IsVisible() {
		isNewKey = true
	}
	if value != nil && !value.IsVisible() {
		isNewValue = true
	}

	vl := getAllVariables(outer.Block())

	tok := token.ASSIGN
	if isNewKey || isNewValue {
		tok = token.DEFINE
		if key != nil {
			key = NewVariable(key.Type).WithName(key.Name()).AutoName().ReadOnly()
			for vl.Check(key.Name()) {
				key.IncreaseName()
			}
		}
		if value != nil {
			value = NewVariable(value.Type).WithName(value.Name()).AutoName().ReadOnly()
			for vl.Check(value.Name()) {
				value.IncreaseName()
			}
		}

	}

	var keyIdent ast.Expr
	var valIdent ast.Expr
	if key != nil {
		keyIdent = key.Ident()
	} else {
		keyIdent = ast.NewIdent("_")
	}
	if value != nil {
		valIdent = value.Ident()
	} else {
		valIdent = ast.NewIdent("_")
	}

	stmt := rangeStmt(keyIdent, valIdent, tok, over.Ident())
	builder.stmt = stmt
	block := newBlockBuilder(builder, stmt.Body)
	builder.block = block

	if key != nil {
		key.SetMode(READ_MODE)
		if !key.IsVisible() {
			block.addVariable(key)
		}
	}
	if value != nil {
		value.SetMode(READ_MODE)
		if !value.IsVisible() {
			block.addVariable(value)
		}
	}

	return builder
}

type DefineStmtBuilder struct {
	*baseBuilder
	stmt *ast.AssignStmt
}

func (b *DefineStmtBuilder) Stmt() ast.Stmt {
	return b.stmt
}

func NewDefine(outer Builder, vl []*Variable, value ast.Expr) *DefineStmtBuilder {
	base := newBaseFromOuter(outer.Block())
	tok := token.DEFINE

	//make result variables
	resultList := []ast.Expr{}
	for _, v := range vl {
		resultList = append(resultList, v.Ident())
	}
	stmt := multiAssginStmt(tok, resultList, value)

	return &DefineStmtBuilder{base, stmt}
}

type CallStmtBuilder struct {
	*baseBuilder
	stmt ast.Stmt
	fnt  parser.Type
}

func (b *CallStmtBuilder) Stmt() ast.Stmt {
	return b.stmt
}
func (b *CallStmtBuilder) Bytes() []byte {
	var buf bytes.Buffer
	if err := format.Node(&buf, token.NewFileSet(), b.stmt); err != nil {
		log.Fatalf("generate code error: %v", err)
	}

	return buf.Bytes()
}

func (b *CallStmtBuilder) Call(vl *VariableList, ignores []int) *VariableList {
	fnt := b.fnt
	file := b.File()

	underFunc, ok := fnt.Underlying().(*parser.FuncType)
	if !ok {
		log.Fatalf("not a function type: %s", fnt)
	}

	//oder: inner to outer
	args := vl.Concat(getAllVariables(b))
	//args.debug()
	paramExprList, ok := exprListFromFields(underFunc.Params, args.Getter(READ_MODE))
	if !ok {
		log.Fatalf("call expr arguments not enough: %s", fnt)
	}
	callExpr := callExpr(parser.TypeExprInFile(fnt, file), paramExprList)

	//make result variables
	var resultList []ast.Expr
	var isNew = false
	getter := args.ResultGetter(ignores)
	resultList, ok = exprListFromFields(underFunc.Results, getter)
	if !ok {
		getter = VarGetterWithIgnore(args.CreateGetter(WRITE_MODE), ignores)
		resultList, _ = exprListFromFields(underFunc.Results, getter)
		isNew = true
	}

	resultVars := getter.Getted()
	resultVars.ForEach(func(idx int, v *Variable) {
		//after the call, result variables will be read only
		//except the error
		if v.Type.EqualTo(parser.ErrorType()) {
			v.SetMode(RW_MODE)
		} else {
			v.SetMode(READ_MODE)
		}
		if !v.IsVisible() {
			b.addVariable(v)
		}
	})

	// utils.PrintNodef(callExpr.Fun, "CallExpr.Fun:")
	var buildStmt ast.Stmt
	if len(resultList) > 0 {
		stmt := &ast.AssignStmt{}
		if isNew {
			stmt.Tok = token.DEFINE
		} else {
			stmt.Tok = token.ASSIGN
		}
		stmt.Lhs = resultList
		stmt.Rhs = []ast.Expr{callExpr}
		buildStmt = stmt
	} else {
		stmt := &ast.ExprStmt{}
		stmt.X = callExpr
		buildStmt = stmt
	}
	b.stmt = buildStmt
	return resultVars
}

func NewCallStmt(outer Builder, fnt parser.Type) *CallStmtBuilder {
	block := outer.Block()
	if block == nil {
		log.Fatalf("builder does not have a block")
	}
	return &CallStmtBuilder{baseBuilder: newBaseFromOuter(block), fnt: fnt}
}

type ReturnStmtBuilder struct {
	*baseBuilder
	stmt *ast.ReturnStmt
}

func (b *ReturnStmtBuilder) Stmt() ast.Stmt {
	return b.stmt
}
func NewReturnStmt(outer Builder, results *VariableList) *ReturnStmtBuilder {
	file := outer.File()
	var function *FuncBuilder
	for x := Builder(outer.Block()); function == nil; x = x.Outer() {
		function, _ = x.(*FuncBuilder)
	}
	if function == nil {
		log.Fatalf("context donnot support return")
	}

	resultExpr := []ast.Expr{}
	ew := function.returnError
	for _, field := range function.funcType.Results {
		res := field.Type

		var v *Variable

		if hasName := field.FieldName(); hasName != "" {
			// log.Println("has name: ", field.Name())
			v = results.GetByTypeAndName(res, hasName)
		} else {
			// log.Println("no name")
			v = results.GetByType(res, READ_MODE)
		}

		if v != nil {
			if ew != nil && v.Type.EqualTo(parser.ErrorType()) {
				resultExpr = append(resultExpr, ew.Wrap(outer, v).Ident())
			} else {
				resultExpr = append(resultExpr, v.Ident())
			}
		} else {
			resultExpr = append(resultExpr, parser.TypeZeroValue(res, file))
		}
	}

	stmt := returnStmt(resultExpr)
	return &ReturnStmtBuilder{newBaseFromOuter(outer), stmt}
}

type IfStmtBuilder struct {
	*baseBuilder
	stmt              *ast.IfStmt
	internalVariables *VariableList
	block             *BlockBuilder
	elseBlock         *BlockBuilder
}

func (b *IfStmtBuilder) Block() *BlockBuilder {
	return b.block
}
func (b *IfStmtBuilder) Bytes() []byte {
	var buf bytes.Buffer
	if err := format.Node(&buf, token.NewFileSet(), b.block); err != nil {
		log.Fatalf("generate code error: %v", err)
	}
	return buf.Bytes()
}
func (b *IfStmtBuilder) Stmt() ast.Stmt {
	stmt := &ast.IfStmt{}
	*stmt = *b.stmt
	if b.elseBlock == nil || b.elseBlock.Body == nil || len(b.elseBlock.Body.List) == 0 {
		stmt.Else = nil
	}
	return stmt
}

func NewIfStmt(outer Builder) *IfStmtBuilder {
	pkg := outer.Package()
	file := outer.File()
	base := newbaseBuilder(outer.Block(), pkg, file)
	ifStmt := ifStmt(nil, nil)
	builder := &IfStmtBuilder{base, ifStmt, NewVariableList(), nil, nil}
	block := newBlockBuilder(builder, ifStmt.Body)
	builder.block = block
	return builder
}

func (b *IfStmtBuilder) SetInitCond(init StmtBuilder, cond ast.Expr) *IfStmtBuilder {
	if cond == nil {
		return nil
	}
	b.stmt.Cond = cond
	if init != nil {
		b.stmt.Init = init.Stmt()
		b.internalVariables = b.variables.Concat(init.Variables())
		if b.block != nil {
			b.block.variables = init.Variables().Concat(b.block.variables)
		}
		if b.elseBlock != nil {
			b.elseBlock.variables = init.Variables().Concat(b.elseBlock.variables)
		}
	}

	return b
}
func (b *IfStmtBuilder) Else() *BlockBuilder {
	if b.elseBlock != nil {
		return b.elseBlock
	}
	blockStmt := &ast.BlockStmt{}
	b.stmt.Else = blockStmt
	block := newBlockBuilder(b.Outer().Block(), blockStmt)
	b.elseBlock = block
	b.elseBlock.variables = b.elseBlock.variables.Concat(b.internalVariables)

	return block
}

func NewIfBuilderWithSrc(outer Builder, src string) Builder {
	cond, err := parser.ParseExpr(src)
	if err != nil {
		log.Fatalf("if statement src error: %s", src)
		return nil
	}

	ifStmt := NewIfStmt(outer)
	ifStmt.SetInitCond(nil, cond)

	return ifStmt
}

type VarAssignBuilder struct {
	*baseBuilder
	stmt *ast.IfStmt
}
