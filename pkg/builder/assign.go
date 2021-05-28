package builder

import (
	"fmt"
	"go/ast"
	"log"

	"github.com/lawrsp/pigo/pkg/parser"
)

type StmtListBuilder interface {
	Builder
	StmtList() []ast.Stmt
}

type StructAssignBuilder struct {
	*BlockBuilder
	knownTpaths []*parser.TPath
	tagName     string
	ev          *Variable
}

func NewStructAssign(outer Builder, tag string, ev *Variable, paths []*parser.TPath) *StructAssignBuilder {
	blockStmt := &ast.BlockStmt{}
	block := newBlockBuilder(outer, blockStmt)
	b := &StructAssignBuilder{BlockBuilder: block}
	if paths != nil {
		b.knownTpaths = paths
	} else {
		b.knownTpaths = []*parser.TPath{}
	}
	b.tagName = tag
	b.ev = ev

	return b
}

func (b *StructAssignBuilder) Stmts() []ast.Stmt {
	return b.BlockBuilder.Body.List
}

func (b *StructAssignBuilder) TryAssign(srcV, dstV *Variable) bool {
	if srcV == nil || dstV == nil {
		err := fmt.Errorf("src == nil ? %v ; dst == nil ? %v", srcV == nil, dstV == nil)
		panic(err)
	}
	if srcV.IsVisible() == false {
		err := fmt.Errorf("src %s is not visible and not anonymous", srcV.Name())
		panic(err)
	}

	//try struct
	srcSt := NewFieldList(b.tagName)
	dstSt := NewFieldList(b.tagName)
	srcOk := parser.InspectUnderlyingStruct(srcV.Type, srcSt.SpreadInspector)
	dstOk := parser.InspectUnderlyingStruct(dstV.Type, dstSt.SpreadInspector)
	// srcSt.Print()
	// dstSt.Print()
	srcStHolderType := GetStructHolder(srcV.Type)
	dstStHolderType := GetStructHolder(dstV.Type)
	var srcStHolderV *Variable
	var dstStHolderV *Variable

	//prepare variables:
	if srcOk {
		if srcStHolderType.EqualTo(srcV.Type) {
			srcStHolderV = srcV
		} else {
			srcStHolderV = NewVariable(srcStHolderType).AutoName().WriteOnly()
			srcStHolderV = AddVariableDecl(b, srcStHolderV)
			paths, _ := parser.TypeToType(srcV.Type, srcStHolderType, []*parser.TPath{})
			FollowTPaths(b, srcV, srcStHolderV, b.ev, paths)
		}
	}

	if dstOk {
		if dstStHolderType.EqualTo(dstV.Type) {
			dstStHolderV = dstV
		} else {
			dstStHolderV = NewVariable(dstStHolderType).WithName(dstV.Name()).WriteOnly()
		}
		if !dstStHolderV.IsVisible() {
			var valueExpr ast.Expr
			if holder, ok := dstStHolderType.Underlying().(*parser.PointerType); ok {
				if x := parser.NotNilPointerValue(holder, b.File()); x != nil {
					valueExpr = x
				}
			}
			dstStHolderV = AddVariableAssign(b, dstStHolderV, valueExpr)
		}
	}

	//assign:
	//dst = src.Field
	if srcOk {
		if field, paths, ok := srcSt.ContainPathToType(dstV.Type); ok {
			srcFieldV := NewVariable(field.Field.Type).WithExpr(srcV.DotExpr(field.Field.Name())).ReadOnly()
			FollowTPaths(b, srcFieldV, dstV, b.ev, paths)
			return true
		}
	}
	//dst.Field = src
	if dstOk {
		if field, _, ok := dstSt.ContainPathToType(srcV.Type); ok {
			dstFieldV := NewVariable(field.Field.Type).WithExpr(dstStHolderV.DotExpr(field.Field.Name()))
			if ok := TryDirectAssign(b, srcV, dstFieldV, b.ev, b.knownTpaths); !ok {
				return false
			}
			// FollowTPaths(b, srcV, dstFieldV, b.ev, paths)
			if dstStHolderV != dstV {
				if ok := TryDirectAssign(b, dstStHolderV, dstV, b.ev, b.knownTpaths); !ok {
					return false
				}

			}
			return true
		}
	}
	//struct => struct
	if srcOk && dstOk {
		//prepare variables:
		log.Printf("struct 2 struct")
		StructAssign(b, srcStHolderV, dstStHolderV, srcSt, dstSt, b.knownTpaths)
		if dstStHolderV != dstV {
			if ok := TryDirectAssign(b, dstStHolderV, dstV, b.ev, b.knownTpaths); !ok {
				return false
			}

		}

		return true
	}

	return false
}

func StructAssign(bd Builder, srcV, dstV *Variable, srcFields, dstFields *FieldList, knownTpaths []*parser.TPath) {
	for _, dstFd := range dstFields.Fields {
		if srcFd := srcFields.GetFieldByName(dstFd.Name); srcFd != nil {
			if paths, ok := parser.TypeToType(srcFd.Field.Type, dstFd.Field.Type, knownTpaths); ok {
				srcV := NewVariable(srcFd.Field.Type).WithExpr(srcV.DotExpr(srcFd.Field.Name())).ReadOnly()
				dstV := NewVariable(dstFd.Field.Type).WithExpr(dstV.DotExpr(dstFd.Field.Name())).WriteOnly()
				// log.Printf("%s = %s, dstV anonymous: %v", srcV.Name(), dstV.Name(), dstV.IsAnonymous())
				FollowTPaths(bd, srcV, dstV, nil, paths)
			} else {
				log.Fatalf("cannot assgin field %s(%s) to %s(%s) knownTpaths(%v)",
					srcFd.Name, srcFd.Field, dstFd.Name, dstFd.Field, knownTpaths)
			}
		}
	}
}

func TryDirectAssign(bd Builder, srcV, dstV *Variable, ev *Variable, knownTpaths []*parser.TPath) bool {
	if srcV == nil || !srcV.IsVisible() {
		log.Fatalf("some bugs: source variable is nil")
		return false
	}
	if paths, ok := parser.TypeToType(srcV.Type, dstV.Type, knownTpaths); ok {
		FollowTPaths(bd, srcV, dstV, ev, paths)
		return true
	}

	return false
}

type StructSetterBuilder struct {
	*BlockBuilder
	tagName     string
	knownTpaths []*parser.TPath
	ev          *Variable

	checkDiff bool

	callFunc func(*Field) parser.Type
	assigner func(*Field) func(Builder, *Variable, ast.Expr)
}

func NewStructSetter(outer Builder, tag string, ev *Variable, paths []*parser.TPath) *StructSetterBuilder {
	blockStmt := &ast.BlockStmt{}
	block := newBlockBuilder(outer, blockStmt)
	b := &StructSetterBuilder{BlockBuilder: block}
	if paths != nil {
		b.knownTpaths = paths
	} else {
		b.knownTpaths = []*parser.TPath{}
	}
	b.tagName = tag
	b.ev = ev

	return b
}

func (b *StructSetterBuilder) Stmts() []ast.Stmt {
	return b.BlockBuilder.Body.List
}
func (b *StructSetterBuilder) WithCheckDiff() *StructSetterBuilder {
	b.checkDiff = true
	return b
}
func (b *StructSetterBuilder) WithAssigner(assigner func(*Field) func(Builder, *Variable, ast.Expr)) *StructSetterBuilder {
	b.assigner = assigner
	return b
}
func (b *StructSetterBuilder) WithCallFunc(callFunc func(*Field) parser.Type) *StructSetterBuilder {
	b.callFunc = callFunc
	return b
}

/*
func (b *StructSetterBuilder) customSetter(srcV, dstV *Variable, srcField, dstField *Field, fnt parser.Type) {
	cbd := NewCallStmt(b, fnt)

	arg := NewVariable(srcField.Field.Type).ReadOnly().WithExpr(srcV.DotExpr(srcField.Field.Name()))
	vl := NewVariableList()
	vl.Add(arg)
	results := cbd.Call(vl, nil)
	if ev := results.GetByType(parser.ErrorType(), ANY_MODE); ev != nil {
		AddCheckReturn(b, ev.CheckNilExpr(false), ev)
	}

	underFunc, ok := fnt.Underlying().(*parser.FuncType)
	if !ok {
		log.Fatalf("custom setter is not a function")
	}

	paramField := underFunc.Params[0]

	if srcV.Type.EqualTo(paramField.Type) {
		//call(srcV.Field)

	} else if paths, ok := parser.TypeToType(srcV.Type, paramField.Type, nil); ok {
		//call(*srcV.Field)
	}
}
*/

func (b *StructSetterBuilder) structAssign(srcV, dstV *Variable, srcFields, dstFields *FieldList) {
	knownTpaths := b.knownTpaths
	// log.Printf("assgin: %v(%s) = %v(%s)", srcV, srcV.Name(), dstV, dstV.Name())
	for _, srcFd := range srcFields.Fields {
		if dstFd := dstFields.GetFieldByName(srcFd.Name); dstFd != nil {
			if paths, ok := parser.TypeToType(srcFd.Field.Type, dstFd.Field.Type, knownTpaths); ok {
				srcV := NewVariable(srcFd.Field.Type).WithExpr(srcV.DotExpr(srcFd.Field.Name())).ReadOnly()
				dstV := NewVariable(dstFd.Field.Type).WithExpr(dstV.DotExpr(dstFd.Field.Name())).WriteOnly()

				// log.Printf("%s = %s", srcV.Name(), dstV.Name())
				// log.Printf("%v = %v", srcV, dstV)
				tpd := NewTPathBuilder(b, b.ev, b.checkDiff)
				if b.assigner != nil {
					tpd = tpd.WithAssigner(b.assigner(dstFd))
				}

				tpd.Follow(srcV, dstV, paths)
				b.Add(tpd)
			} else {
				log.Fatalf("cannot assgin field %s(%s) to %s(%s)", srcFd.Name, srcFd.Field, dstFd.Name, dstFd.Field)
			}
		}
	}
}

func (b *StructSetterBuilder) TryAssign(srcV, dstV *Variable) bool {
	if srcV == nil || dstV == nil {
		err := fmt.Errorf("src == nil ? %v ; dst == nil ? %v", srcV == nil, dstV == nil)
		panic(err)
	}
	if srcV.IsVisible() == false {
		err := fmt.Errorf("src %s is not visible and not anonymous", srcV.Name())
		panic(err)
	}
	// log.Printf("try assgin: %v(%s) = %v(%s)", srcV, srcV.Name(), dstV, dstV.Name())

	//try struct
	srcSt := NewFieldList(b.tagName)
	dstSt := NewFieldList(b.tagName)
	srcOk := parser.InspectUnderlyingStruct(srcV.Type, srcSt.SpreadInspector)
	dstOk := parser.InspectUnderlyingStruct(dstV.Type, dstSt.SpreadInspector)
	srcSt.Print()
	dstSt.Print()
	srcStHolderType := GetStructHolder(srcV.Type)
	dstStHolderType := GetStructHolder(dstV.Type)
	var srcStHolderV *Variable
	var dstStHolderV *Variable

	//prepare variables:
	if srcOk {
		if srcStHolderType.EqualTo(srcV.Type) {
			srcStHolderV = srcV
		} else {
			srcStHolderV = NewVariable(srcStHolderType).AutoName().WriteOnly()
			srcStHolderV = AddVariableDecl(b, srcStHolderV)
			paths, _ := parser.TypeToType(srcV.Type, srcStHolderType, []*parser.TPath{})
			FollowTPaths(b, srcV, srcStHolderV, b.ev, paths)
		}
	}

	if dstOk {
		if dstStHolderType.EqualTo(dstV.Type) {
			dstStHolderV = dstV
		} else {
			dstStHolderV = NewVariable(dstStHolderType).WithName(dstV.Name()).WriteOnly()
		}
		if !dstStHolderV.IsVisible() {
			var valueExpr ast.Expr = parser.TypeInitValue(dstStHolderType, b.File())
			// if holder, ok := dstStHolderType.Underlying().(*parser.PointerType); ok {
			// if x := parser.NotNilPointerValue(holder, b.File()); x != nil {
			// valueExpr = x
			// }
			// }
			dstStHolderV = AddVariableAssign(b, dstStHolderV, valueExpr)
		}
	}

	//struct => struct
	if srcOk && dstOk {
		//prepare variables:
		b.structAssign(srcStHolderV, dstStHolderV, srcSt, dstSt)
		if dstStHolderV != dstV {
			if ok := TryDirectAssign(b, dstStHolderV, dstV, b.ev, b.knownTpaths); !ok {
				return false
			}

		}

		return true
	}

	return false
}
