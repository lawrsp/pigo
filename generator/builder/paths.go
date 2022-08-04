package builder

import (
	"go/ast"
	"go/token"
	"log"

	"github.com/lawrsp/pigo/generator/parser"
)

func testIgnores(fnt parser.Type, wanted []parser.Type) []int {
	underFunc := fnt.Underlying().(*parser.FuncType)
	ignores := []int{}
	for i, r := range underFunc.Results {
		ifWant := false
		for _, wt := range wanted {
			if r.Type.EqualTo(wt) {
				ifWant = true
				break
			}
		}
		if !ifWant {
			ignores = append(ignores, i)
		}
	}
	return ignores
}

type TPathBuilder struct {
	*BlockBuilder

	checkDiff bool
	assigner  func(Builder, *Variable, ast.Expr)
	ev        *Variable
	neverNils []*Variable
}

func NewTPathBuilder(outer Builder, ev *Variable, checkDiff bool) *TPathBuilder {
	blockStmt := &ast.BlockStmt{}

	block := newBlockBuilder(outer, blockStmt)
	b := &TPathBuilder{BlockBuilder: block}
	b.ev = ev
	b.checkDiff = checkDiff
	b.assigner = func(inBuilder Builder, dstV *Variable, value ast.Expr) {
		AddVariableAssign(inBuilder, dstV, value)
	}

	return b
}

func (b *TPathBuilder) WithAssigner(assigner func(Builder, *Variable, ast.Expr)) *TPathBuilder {
	b.assigner = assigner
	return b
}

func (b *TPathBuilder) WithTargetSetter(fnt parser.Type) *TPathBuilder {
	return b
}
func (b *TPathBuilder) Stmts() []ast.Stmt {
	return b.BlockBuilder.Body.List
}
func (b *TPathBuilder) checkNeverNil(v *Variable) bool {
	if v == nil {
		return false
	}
	for _, nv := range b.neverNils {
		if nv == v {
			return true
		}
	}
	return false
}
func (b *TPathBuilder) addNeverNil(v *Variable) {
	b.neverNils = append(b.neverNils, v)
}

func (b *TPathBuilder) assignTargetVar(inBuilder Builder, dstV *Variable, midV *Variable, isPointerValue bool) {

	var value ast.Expr

	if isPointerValue {
		// log.Printf(" %v = %v", dstV.Name(), midV.Name())
		if b.ev != nil {
			//if midV == nil {
			//  return ..err
			//}
			AddCheckReturn(inBuilder, midV.CheckNilExpr(true), b.ev)
			if b.checkDiff {
				//if v != *midV {
				//  v = *midV
				//}
				ifB := NewIfStmt(inBuilder).SetInitCond(nil, dstV.CheckEqualExpr(false, midV.PointerValueExpr()))
				inBuilder.Block().Add(ifB)
				inBuilder = ifB
			}
			// else {
			// 	// v = *midV
			// }
		} else {
			var cond ast.Expr
			if b.checkDiff {
				// if midV != nil && v != *midV{
				//		v = *midV
				// }
				cond = condExpr(token.LAND, midV.CheckNilExpr(false), dstV.CheckEqualExpr(false, midV.PointerValueExpr()))
			} else {
				// if midV != nil {
				//		v = *midV
				// }
				cond = midV.CheckNilExpr(false)
			}

			ifB := NewIfStmt(inBuilder).SetInitCond(nil, cond)
			inBuilder.Block().Add(ifB)
			inBuilder = ifB
		}

		value = midV.PointerValueExpr()
	} else {
		if b.checkDiff {
			// if  v != midV {
			//		v = midV
			// }
			cond := dstV.CheckEqualExpr(false, midV.Ident())
			ifB := NewIfStmt(inBuilder).SetInitCond(nil, cond)
			inBuilder.Block().Add(ifB)
			inBuilder = ifB
		}
		// else {
		// 	// v = midV
		// }
		value = midV.Ident()
	}

	b.assigner(inBuilder, dstV, value)
}

func (bd *TPathBuilder) do(inBuilder Builder, v, midV, dstV *Variable, fp *parser.TPath) (Builder, *Variable) {

	switch fp.D {
	case parser.D_Self: //do nothing
		v = midV
	case parser.D_SkipPointer: //pointer assign
		//// 2 choices:
		//// 1:
		// if midV != nil {
		//		v = *midV
		// }
		//// or 2:
		// if midV == nil {
		//  return ..err
		// }
		// v = *midV

		// log.Printf("--2222222-------%v != %v: %v", v.Type, dstV.Type, dstV.IsVisible())
		if v == dstV && !v.IsVisible() {
			// log.Printf("%s is not visible", v.Name())

			v = AddVariableDecl(inBuilder, v)
		}

		if bd.ev != nil {
			//has nil check-return
			AddCheckReturn(inBuilder, midV.CheckNilExpr(true), bd.ev)
			v = AddVariableAssign(inBuilder, v, midV.PointerValueExpr())
		} else {
			cond := midV.CheckNilExpr(false)
			ifB := NewIfStmt(inBuilder).SetInitCond(nil, cond)
			v = AddVariableAssign(ifB, v, midV.PointerValueExpr())
			inBuilder.Block().Add(ifB)
			inBuilder = ifB
		}
	case parser.D_SkipBracket: //for range
		//for _, v := range midV {
		//   ...inBuilder
		//}
		fr := NewForRange(inBuilder, nil, v, midV)
		v = GetVariable(fr, fp.Target, READ_MODE, Scope_Function)
		inBuilder.Block().Add(fr)
		inBuilder = fr
		// FollowTPaths(inBuilder, v, dstV, ev, paths[i+1:])
	case parser.D_SpreadFields:
		//v := midV.field
		field := fp.Arg.(*parser.Field)
		v = AddVariableAssign(inBuilder, v, midV.DotExpr(field.Name()))
	case parser.D_CallFunction:
		// v, err := convert(midV)
		// or:
		// v := convert(midV)

		if _, ok := midV.Type.Underlying().(*parser.PointerType); ok {
			//if midV != nil {
			//   var v
			//   ...callInBuilder
			//}
			if !bd.checkNeverNil(midV) {
				ifB := NewIfStmt(inBuilder)
				ifB.SetInitCond(nil, midV.CheckNilExpr(false))
				inBuilder.Block().Add(ifB)
				inBuilder = ifB
			}
		}

		fnt := fp.Arg.(parser.Type)
		params := NewVariableList()
		params.Add(midV)
		if v.IsVisible() {
			params.Add(v)
		}

		callStmt := NewCallStmt(inBuilder, fnt)
		ignores := testIgnores(fnt, []parser.Type{fp.Target, parser.ErrorType()})
		results := callStmt.Call(params, ignores)

		inBuilder.Block().Add(callStmt)
		v = results.GetByType(fp.Target, READ_MODE)

		if errV := results.GetByType(parser.ErrorType(), READ_MODE); errV != nil {
			// if err != nil {
			//    return ..err
			// }
			ifB := NewIfStmt(inBuilder)
			ifB.SetInitCond(nil, errV.CheckNilExpr(false))
			AddErrorReturn(ifB, errV)
			inBuilder.Block().Add(ifB)
		}

	case parser.D_AddPointer:
		//v := &midV
		v = AddVariableAssign(inBuilder, v, midV.AddressExpr())
		bd.addNeverNil(v)
	case parser.D_AddBracket:
		// v = append(v, midV)
		//// or:
		//if midV != nil {
		// ..append
		//}
		appendInBuilder := inBuilder
		if _, ok := fp.Source.Underlying().(*parser.PointerType); ok {
			if !bd.checkNeverNil(midV) {
				ifB := NewIfStmt(inBuilder).SetInitCond(nil, midV.CheckNilExpr(false))
				appendInBuilder = ifB
				inBuilder.Block().Add(ifB)
			}
		}
		AddAppendStmt(appendInBuilder, v, midV)
		if inBuilder.Outer().Block() != nil {
			inBuilder = inBuilder.Outer()
		}

	case parser.D_TypeConversion:
		//v = type(midV)
		typ := fp.Arg.(parser.Type)
		typeExpr := parser.TypeExprInFile(typ, inBuilder.File())
		v = AddVariableAssign(inBuilder, v, midV.TypeConversionExpr(typeExpr))

		// return
	}

	return inBuilder, v
}

func (bd *TPathBuilder) Follow(srcV, dstV *Variable, paths []*parser.TPath) {

	//src should not be nil
	if srcV == nil {
		log.Fatalf("srcV is nil")
	}

	// for i, fp := range paths {
	// 	log.Printf("%d %s => %s %d", i, fp.Source, fp.Target, fp.D)
	// }
	if len(paths) == 1 {
		if dstV != nil {
			// AddVariableAssign(bd, dstV, srcV.Ident())

			bd.assignTargetVar(bd, dstV, srcV, false)
		}
		return
	}

	if dstV == nil {
		dstType := paths[len(paths)-1].Target
		dstV = GetVariable(bd, dstType, WRITE_MODE, Scope_Function)
		if dstV == nil {
			dstV = NewVariable(dstType).AutoName().WriteOnly()
			// dstV = AddVariableDecl(bd, dstV)
		}
	}

	//prepare variable:
	addBracketTPath := []*parser.TPath{}
	var skipBracket, addBracket int
	for _, fp := range paths[1:] {
		if fp.D == parser.D_SkipBracket {
			skipBracket++
		}
		if fp.D == parser.D_AddBracket {
			addBracketTPath = append(addBracketTPath, fp)
			addBracket++
		}
	}

	midV := srcV
	inBuilder := Builder(bd)

	if skipBracket > 0 {
		if !dstV.IsVisible() {
			dstV = AddVariableDecl(bd, dstV)

			// if _, ok := dstV.Type.Underlying().(*parser.ArrayType); ok {
			// 	valueExpr := parser.TypeInitValue(dstV.Type, bd.File())
			// 	dstV = AddVariableAssign(bd, dstV, valueExpr)
			// } else {
			// 	dstV = AddVariableDecl(bd, dstV)
			// }
		}
	}

	//loop:
	for _, fp := range paths {
		for addBracket >= skipBracket && addBracket > 0 {
			abtp := addBracketTPath[0]
			addBracketTPath = addBracketTPath[1:]
			addBracket--

			if v := GetVariable(inBuilder, abtp.Target, ANY_MODE, Scope_Function); v == nil {
				AddVariableDecl(inBuilder, NewVariable(abtp.Target).AutoName().WriteOnly())
			}
		}
		if fp.D == parser.D_SkipBracket {
			skipBracket--
		}

		// if inBuilder.Block() == nil {
		// 	log.Printf("block is nil =======")
		// }

		var v *Variable = GetVariable(inBuilder, fp.Target, WRITE_MODE, Scope_Function)
		if v == nil {
			if dstV != nil && dstV.Type.EqualTo(fp.Target) {
				v = dstV
			} else {
				v = NewVariable(fp.Target).AutoName().WriteOnly()
			}
		}

		if v == dstV && (fp.D == parser.D_Self || fp.D == parser.D_SkipPointer) {
			if !v.IsVisible() {
				v = AddVariableDecl(inBuilder, v)
			}
			bd.assignTargetVar(inBuilder, dstV, midV, fp.D == parser.D_SkipPointer)
			return
		}

		inBuilder, midV = bd.do(inBuilder, v, midV, dstV, fp)
	}

	//if not assigned to dstV:
	if dstV != nil && dstV != midV {
		bd.assignTargetVar(inBuilder, dstV, midV, false)
	}
}

func FollowTPaths(bd Builder, srcV, dstV *Variable, ev *Variable, paths []*parser.TPath) {

	// log.Printf("paths: %v, dstV: %p", paths, dstV)
	tpbuilder := NewTPathBuilder(bd, ev, false)
	tpbuilder.Follow(srcV, dstV, paths)

	bd.Add(tpbuilder)

	/*
		neverNils := []*Variable{}
		checkNeverNil := func(v *Variable) bool {
			if v == nil {
				return false
			}
			for _, nv := range neverNils {
				if nv == v {
					return true
				}
			}
			return false
		}

		//src should not be nil
		if srcV == nil {
			log.Fatalf("srcV is nil")
		}

		// for i, fp := range paths {
		// 	log.Printf("%d %s => %s %d", i, fp.Source, fp.Target, fp.D)
		// }

		if len(paths) == 1 {
			if dstV != nil {
				AddVariableAssign(bd, dstV, srcV.Ident())
			}
			return
		}

		if dstV == nil {
			dstType := paths[len(paths)-1].Target
			dstV = GetVariable(bd, dstType, WRITE_MODE, Scope_Function)
			if dstV == nil {
				dstV = NewVariable(dstType).AutoName().WriteOnly()
				// dstV = AddVariableDecl(bd, dstV)
			}
		}

		//prepare variable:
		addBracketTPath := []*parser.TPath{}
		var skipBracket, addBracket int
		for _, fp := range paths[1:] {
			if fp.D == parser.D_SkipBracket {
				skipBracket += 1
			}
			if fp.D == parser.D_AddBracket {
				addBracketTPath = append(addBracketTPath, fp)
				addBracket += 1
			}
		}

		midV := srcV
		inBuilder := bd

		if skipBracket > 0 {
			if dstV.IsVisible() == false {
				if _, ok := dstV.Type.Underlying().(*parser.ArrayType); ok {
					valueExpr := parser.TypeInitValue(dstV.Type, bd.File())
					dstV = AddVariableAssign(bd, dstV, valueExpr)
				} else {
					dstV = AddVariableDecl(bd, dstV)
				}
			}
		}

		//loop:
		for _, fp := range paths {
			for addBracket >= skipBracket && addBracket > 0 {
				abtp := addBracketTPath[0]
				addBracketTPath = addBracketTPath[1:]
				addBracket -= 1

				if !dstV.IsVisible() || !dstV.Type.EqualTo(abtp.Target) {
					if v := GetVariable(inBuilder, abtp.Target, WRITE_MODE, Scope_Function); v == nil {
						AddVariableDecl(inBuilder, NewVariable(abtp.Target).AutoName().WriteOnly())
					}
				}
			}
			if fp.D == parser.D_SkipBracket {
				skipBracket -= 1
			}

			// if inBuilder.Block() == nil {
			// 	log.Printf("block is nil =======")
			// }

			var v *Variable
			if dstV != nil && dstV.Type.EqualTo(fp.Target) {
				v = dstV
			} else {
				v = GetVariable(inBuilder, fp.Target, WRITE_MODE, Scope_Function)
			}
			if v == nil {
				v = NewVariable(fp.Target).AutoName().WriteOnly()
			}

			switch fp.D {
			case parser.D_Self: //do nothing
				v = midV
			case parser.D_SkipPointer: //pointer assign
				//// 2 choices:
				//// 1:
				// if midV != nil {
				//		v := *midV
				// }
				//// or 2:
				// if midV == nil {
				//  return ..err
				// }
				// v = *midV

				// else {
				// 	log.Printf("=====v %s(%s) has benn added", v.Name(), v.Type)
				// }

				if v == dstV && !v.IsVisible() {
					// log.Printf("%s is not visible: anonymous: %v", v.Name(), dstV.IsAnonymous())
					v = AddVariableDecl(inBuilder, v)
				}

				if ev != nil {
					//has nil check-return
					AddCheckReturn(inBuilder, midV.CheckNilExpr(true), ev)
					v = AddVariableAssign(inBuilder, v, midV.PointerValueExpr())
				} else {
					ifB := NewIfStmt(inBuilder).SetInitCond(nil, midV.CheckNilExpr(false))
					v = AddVariableAssign(ifB, v, midV.PointerValueExpr())
					inBuilder.Block().Add(ifB)
					inBuilder = ifB
				}
			case parser.D_SkipBracket: //for range
				//for _, v := range midV {
				//   ...inBuilder
				//}
				fr := NewForRange(inBuilder, nil, v, midV)
				v = GetVariable(fr, fp.Target, READ_MODE, Scope_Function)
				inBuilder.Block().Add(fr)
				inBuilder = fr
				// FollowTPaths(inBuilder, v, dstV, ev, paths[i+1:])
			case parser.D_SpreadFields:
				//v := midV.field
				field := fp.Arg.(*parser.Field)
				v = AddVariableAssign(inBuilder, v, midV.DotExpr(field.Name()))
			case parser.D_CallFunction:
				//v, err := convert(midV)

				if _, ok := midV.Type.Underlying().(*parser.PointerType); ok {
					//if midV != nil {
					//   var v
					//   ...callInBuilder
					//}
					if checkNeverNil(midV) == false {
						ifB := NewIfStmt(inBuilder)
						ifB.SetInitCond(nil, midV.CheckNilExpr(false))
						inBuilder.Block().Add(ifB)
						inBuilder = ifB
					}
				}

				fnt := fp.Arg.(parser.Type)
				params := NewVariableList()
				params.Add(midV)
				if v.IsVisible() {
					params.Add(v)
				}

				callStmt := NewCallStmt(inBuilder, fnt)
				ignores := testIgnores(fnt, []parser.Type{fp.Target, parser.ErrorType()})
				results := callStmt.Call(params, ignores)

				inBuilder.Block().Add(callStmt)
				v = results.GetByType(fp.Target, READ_MODE)

				if errV := results.GetByType(parser.ErrorType(), READ_MODE); errV != nil {
					// if err != nil {
					//    return ..err
					// }
					ifB := NewIfStmt(inBuilder)
					ifB.SetInitCond(callStmt, errV.CheckNilExpr(false))
					AddErrorReturn(ifB, errV)
					inBuilder.Block().Add(ifB)
				}

			case parser.D_AddPointer:
				//v := &midV
				v = AddVariableAssign(inBuilder, v, midV.AddressExpr())
				neverNils = append(neverNils, v)
			case parser.D_AddBracket:
				// v = append(v, midV)
				//// or:
				//if midV != nil {
				// ..append
				//}
				appendInBuilder := inBuilder
				if _, ok := fp.Source.Underlying().(*parser.PointerType); ok {
					if checkNeverNil(midV) == false {
						ifB := NewIfStmt(inBuilder).SetInitCond(nil, midV.CheckNilExpr(false))
						appendInBuilder = ifB
						inBuilder.Block().Add(ifB)
					}
				}
				AddAppendStmt(appendInBuilder, v, midV)
				if inBuilder.Outer().Block() != nil {
					inBuilder = inBuilder.Outer()
				}

			case parser.D_TypeConversion:
				//v = type(midV)
				typ := fp.Arg.(parser.Type)
				typeExpr := parser.TypeExprInFile(typ, inBuilder.File())
				v = AddVariableAssign(inBuilder, v, midV.TypeConversionExpr(typeExpr))

				// return
			}

			midV = v
		}
		if dstV != nil && dstV != midV {
			AddVariableAssign(bd, dstV, midV.Ident())
		}
	*/
}
