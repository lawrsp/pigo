package parser

import (
	"go/ast"
	"go/token"
	"log"

	"github.com/lawrsp/pigo/generator/printutil"
)

func checkFuncDecl(decl ast.Decl, name string) *ast.FuncDecl {
	if funDecl, ok := decl.(*ast.FuncDecl); ok {
		if funDecl.Name != nil {
			// log.Printf("decl Name %s ==? name: %s", funDecl.Name.Name, name)
			if funDecl.Name.Name == name {
				if funDecl.Recv != nil && len(funDecl.Recv.List) > 0 {
					// log.Printf("finded method %s", name)
					return funDecl
				}
			}
		}
	}
	return nil
}

//method need lookup
func (file *File) LookupMethod(name string) ([]*File, []*ast.FuncDecl) {
	resultFuncs := []*ast.FuncDecl{}
	resultFiles := []*File{}
	for _, decl := range file.File.Decls {
		if funcDecl := checkFuncDecl(decl, name); funcDecl != nil {
			resultFiles = append(resultFiles, file)
			resultFuncs = append(resultFuncs, funcDecl)
		}
	}

	for _, of := range file.BelongTo.Files {
		if of == file {
			continue
		}
		for _, decl := range of.File.Decls {
			if funcDecl := checkFuncDecl(decl, name); funcDecl != nil {
				resultFiles = append(resultFiles, file)
				resultFuncs = append(resultFuncs, funcDecl)
			}
		}
	}
	return resultFiles, resultFuncs
}

func (file *File) LookupName(name string) (*File, *ast.Object) {
	//file level Scope, type and function are in scope
	if obj := file.File.Scope.Lookup(name); obj != nil {
		return file, obj
	}
	return nil, nil
}

func (file *File) LookupImport(name string) (*File, *ast.Object) {
	//file level import:
	if imptPath, ok := file.FindImportPath(name); ok {
		impt := file.FindImport(name)
		if impt == nil {
			impt = file.parser.ImportPackage(name, imptPath, file.BelongTo.Dir)
			if impt != nil {
				file.AddImport(name, impt)
			}
		}
		if impt != nil {
			obj := ast.NewObj(ast.Pkg, name)
			obj.Data = impt.Scope
			return file, obj
		}
	}
	return nil, nil
}

func (file *File) lookup(name string) (*File, *ast.Object) {
	//file level Scope and funcDecl:
	if f, obj := file.LookupName(name); obj != nil {
		return f, obj
	}

	//file level import:
	if f, obj := file.LookupImport(name); obj != nil {
		return f, obj
	}

	//package level Scope and funDecl:
	for _, f := range file.BelongTo.Files {
		if f == file {
			//skip self
			continue
		}
		if of, obj := f.LookupName(name); obj != nil {
			return of, obj
		}
	}

	return nil, nil
}

func funcTypeFromFuncType(file *File, fnType *ast.FuncType) *FuncType {
	ft := &FuncType{}
	paramList := []*Field{}
	if fnType.Params != nil && fnType.Params.List != nil && len(fnType.Params.List) > 0 {
		for _, param := range fnType.Params.List {
			paramType := &Field{}
			if param.Names != nil {
				paramType.SetName(param.Names[0].Name)
			}
			paramType.SetType(file.ReduceType(param.Type))
			//NewUnknownType(file, param.Type))
			paramList = append(paramList, paramType)
		}
	}
	resultList := []*Field{}
	if fnType.Results != nil && fnType.Results.List != nil && len(fnType.Results.List) > 0 {
		for _, result := range fnType.Results.List {
			resType := &Field{}
			if result.Names != nil {
				resType.SetName(result.Names[0].Name)
			}
			resType.SetType(file.ReduceType(result.Type))
			//NewUnknownType(file, result.Type))
			resultList = append(resultList, resType)
		}
	}
	ft.Params = paramList
	ft.Results = resultList
	return ft
}

func funcTypeFromFuncDecl(file *File, decl *ast.FuncDecl) Type {
	name := decl.Name.Name
	fnt := funcTypeFromFuncType(file, decl.Type)
	ft := fnt.Underlying().(*FuncType)
	if decl.Recv != nil {
		ft.Receiver = &Field{}
		if decl.Recv.List[0].Names != nil {
			ft.Receiver.SetName(decl.Recv.List[0].Names[0].Name)
		}
		if len(decl.Recv.List) > 1 {
			log.Fatalf("receiver list == %d", len(decl.Recv.List))
		}
		// printutil.PrintNodef(decl.Recv.List[0].Type, "%s decl.Recv.List[0].Type:", name)
		// ft.receiver.Type = NewUnknownType(decl.Recv.List[0].Type)
		t := file.ReduceType(decl.Recv.List[0].Type)
		ft.Receiver.SetType(t)
	}

	// nt := TypeWithFile(TypeWithName(ft, name), file)
	// nt := &namedType{name: name, file: file}
	// nt.SetUnderlying(ft)
	return TypeWithName(fnt, name)
}

func preDotType(t Type) (*File, Type) {
	file := t.File()
	for res := t; ; {
		switch utt := res.(type) {
		case *packageType:
			return utt.file, utt
		case *namedType:
			return file, utt
		case *StructType:
			for _, field := range utt.Fields {
				// ResolveUnknownField(utt.Fields[i])
				ResolveUnknownField(field)
			}
			return file, utt
		case *FuncType:
			if len(utt.Results) != 1 {
				log.Fatalf("cannot decide cause function result is not 1")
			}
			ResolveUnknownField(utt.Results[0])
			res = utt.Results[0].Type
		case *InterfaceType:
			return file, utt
		case *aliasType:
			res = utt.alias
		case *PointerType:
			res = utt.Base
		case *filedType:
			res = utt.Type
			file = utt.file
		default:
			printutil.PrintNodef(t, "cannot support pre dot:")

			return nil, nil
		}
	}
}

func (file *File) tryTypeField(res Type, name string) Type {
	if st, ok := res.Underlying().(*StructType); ok {
		for _, field := range st.Fields {
			if field.Name() == name {
				ResolveUnknownField(field)
				return field.Type
			}
		}
	}
	return nil
}

func (file *File) tryInterfaceFunc(res Type, name string) Type {
	if it, ok := res.Underlying().(*InterfaceType); ok {
		for _, field := range it.Methods {
			ResolveUnknownField(field)
			// log.Printf("field.Name %s, realType: %s", field.Name(), resType)
			if field.Name() == name {
				// nt := &namedType{name: t.Sel.Name, file: realFile}
				// nt.SetUnderlying(resType)
				nt := TypeWithFile(TypeWithName(field.Type, name), field.Type.File())
				// log.Printf("%s.%s: file: %v", res, nt, nt.File())
				return nt
			}
		}
	}
	return nil
}

func (file *File) tryTypeMethod(res Type, name string) Type {
	dfiles, decls := file.LookupMethod(name)
	if decls == nil || len(decls) == 0 {
		return nil
	}

	for i, f := range dfiles {
		// printutil.PrintNodef(decl, "func %s", t.Sel.Name)
		// log.Printf("file %s fined func %s", resfile.Name, t.Sel.Name)
		nt := funcTypeFromFuncDecl(f, decls[i])
		ft := nt.Underlying().(*FuncType)
		ResolveUnknownField(ft.Receiver)
		// log.Printf("%s === %s ? %v", ft.Receiver.Type, res, TypeEqualAsReceiver(ft.Receiver.Type, res))
		if TypeEqualAsReceiver(ft.Receiver.Type, res) {
			return TypeWithFile(nt, f)
		}
		// else {
		//	log.Printf("receiver not equal: %s != %s", recvType, res)
		// }

	}
	return nil
}

func (file *File) tryPackageName(res Type, expr ast.Expr) Type {
	if pt, ok := res.(*packageType); ok {
		impt := file.FindImport(pt.name)
		for _, imptFile := range impt.Files {
			if t := imptFile.ReduceType(expr); t != nil {
				return t
			}
		}
	}
	return nil
}

func (file *File) ReduceType(expr ast.Expr) Type {

	switch t := expr.(type) {
	//Exprs:
	case *ast.BadExpr:
		log.Fatalf("cannot support BadExpr")
	case *ast.Ident:
		name := t.Name
		if try := Universe.Lookup(name); try != nil {
			return try
		}

		nf, obj := file.lookup(name)
		if obj == nil {
			// log.Printf("cannot find name %s in File %s", name, file.Name)
			return nil
		}
		switch obj.Kind {
		case ast.Typ:
			//ast.TypeSpe
			// log.Printf("type object %s", name)
			spec := obj.Decl.(*ast.TypeSpec)
			if spec.Assign.IsValid() {
				//type alias:
				at := &aliasType{name: name}
				at.SetAlias(nf.ReduceType(spec.Type))
				return TypeWithFile(at, nf)
			} else {
				// nt := &namedType{name: name, file: nf}
				underlying := nf.ReduceType(spec.Type)
				// if t, ok := underlying.(*filedType); ok {
				//	log.Printf("filetype=======")
				//	if _, ok := t.Type.(*StructType); ok {
				//		log.Printf("structType========")
				//	}
				// }
				nt := TypeWithFile(TypeWithName(underlying, name), nf)
				return nt
			}
		case ast.Var:
			//ast.ValueSpec
			// log.Printf("var object %s", name)
			spec := obj.Decl.(*ast.ValueSpec)
			if spec.Type != nil {
				return nf.ReduceType(spec.Type)
			} else {
				return nf.ReduceType(spec.Values[0])
			}
		case ast.Fun:
			//ast.FuncSepc
			// log.Printf("fun object %s", name)
			ft := funcTypeFromFuncDecl(nf, obj.Decl.(*ast.FuncDecl))
			return TypeWithFile(ft, nf)
		case ast.Pkg:
			//ast.Scope
			// log.Printf("pkg object %s", name)
			impt := nf.FindImport(name)
			pt := &packageType{name: name, path: impt.Path, file: nf}
			return pt
		default:
			log.Fatalf("unsupported type of object %s", name)
		}

	case *ast.Ellipsis:
		log.Fatalln("cannot support Ellipsis")
		return nil
	case *ast.BasicLit:
		log.Fatalf("cannot support BasicLit")
	case *ast.FuncLit:
		log.Fatalf("cannot support FuncLit")
	case *ast.CompositeLit:
		//Type{}  will reduced by the Type
		return file.ReduceType(t.Type)
	case *ast.ParenExpr:
		log.Fatalf("cannot support ParenExpr")
	case *ast.SelectorExpr:
		//X.Sel
		//@TODO:
		xt := file.ReduceType(t.X)
		if xt == nil {
			log.Fatalf("reduct type(%v) failed", t.X)
		}
		resfile, res := preDotType(xt)
		if res == nil {
			return nil
		}

		//package.Name
		if pt := resfile.tryPackageName(res, t.Sel); pt != nil {
			return pt
		}
		//interface.func
		if ft := resfile.tryInterfaceFunc(res, t.Sel.Name); ft != nil {
			return ft
		}
		//type.Field
		if fd := resfile.tryTypeField(res, t.Sel.Name); fd != nil {
			return fd
		}
		//type.Method
		if md := resfile.tryTypeMethod(res, t.Sel.Name); md != nil {
			return md
		}

	case *ast.IndexExpr:
		//x[Index]
		res := file.ReduceType(t.X)
		switch ut := res.Underlying().(type) {
		case *ArrayType:
			switch t.Index.(type) {
			case *ast.BasicLit:
				//x[1]
				return ut.Element
			case *ast.KeyValueExpr:
				//x[1:2]
				return TypeWithFile(TypeWithSlice(TypeSkipBracket(ut, 1)), res.File())
			default:
				printutil.FatalNodef(t.Index, "index expr error:")
			}
		// case *sliceType:
		// 	switch t.Index.(type) {
		// 	case *ast.BasicLit:
		// 		if ut.brackets > 1 {
		// 			return TypeWithFile(&sliceType{element: ut.element, brackets: ut.brackets - 1}, res.File())
		// 		}
		// 		return ut.element
		// 	case *ast.KeyValueExpr:
		// 		return TypeWithFile(&sliceType{element: ut.element, brackets: ut.brackets}, res.File())
		// 	default:
		// 		printutil.FatalNodef(t.Index, "index expr error:%v")

		// 	}
		case *mapType:
			return ut.val
		default:
			log.Fatalf("type %s donot suppoort index", ut)
			return nil
		}

	case *ast.SliceExpr:
		//[]X
		res := file.ReduceType(t.X)
		return TypeWithFile(TypeWithSlice(res), res.File())

	case *ast.TypeAssertExpr:
		log.Fatalf("cannot support TypeAssertExpr")
	case *ast.CallExpr:
		res := file.ReduceType(t.Fun)
		if ft, ok := res.Underlying().(*FuncType); ok {
			if ft.Results == nil || len(ft.Results) != 1 {
				//parent ?
				log.Fatalf("results != 1")
				return NewUnknownType(res.File(), t.Fun)
			}
			ResolveUnknownField(ft.Results[0])
			return ft.Results[0].Type
		}

		log.Fatalf("cannot support CallExpr")
	case *ast.StarExpr:
		//*X
		res := file.ReduceType(t.X)
		if res == nil {
			log.Fatalf("cannot reduce type: %v", ExprToString(t.X))
		}
		// printutil.PrintNodef(t.X, "")
		return TypeWithFile(TypeWithPointer(res), res.File())
		// if srt, ok := res.Underlying().(*PointerType); ok {
		//	return TypeWithFile(&PointerType{base: srt.base, stars: srt.stars + 1}, res.File())
		// }
		// return TypeWithFile(&PointerType{base: typ, stars: 1}, res.File())

	case *ast.UnaryExpr:
		//*X  &X
		if t.Op != token.MUL && t.Op != token.AND {
			printutil.FatalNodef(t, "cannot support UnaryExpr:")
		}

		res := file.ReduceType(t.X)
		return TypeWithFile(TypeWithPointer(res), res.File())
	case *ast.BinaryExpr:
		log.Fatalf("cannot support BinaryExpr")
	case *ast.KeyValueExpr:
		log.Fatalf("cannot support KeyValueExpr")

		//Types:
	case *ast.ArrayType:
		element := file.ReduceType(t.Elt)
		if t.Len == nil {
			return TypeWithFile(TypeWithSlice(element), element.File())
		}

		typ := &ArrayType{}
		typ.Len = 1
		typ.lenExpr = t.Len
		typ.Element = element
		return TypeWithFile(typ, element.File())

	case *ast.StructType:
		typ := &StructType{}
		for _, fd := range t.Fields.List {
			field := &Field{}
			if fd.Names != nil && len(fd.Names) > 0 {
				field.SetName(fd.Names[0].Name)
			}
			field.SetType(NewUnknownType(file, fd.Type))
			if fd.Tag != nil {
				if len := len(fd.Tag.Value); len > 2 {
					field.Tag = fd.Tag.Value[1 : len-1]
				}
			}

			// log.Printf("field: %s", field)
			typ.Fields = append(typ.Fields, field)
		}
		return TypeWithFile(typ, file)

	case *ast.FuncType:
		return TypeWithFile(funcTypeFromFuncType(file, t), file)
	case *ast.InterfaceType:
		// printutil.FatalNodef(t, "InterfaceType:")
		typ := &InterfaceType{}
		// res := TypeWithFile(typ, file)
		fields := []*Field{}
		for _, method := range t.Methods.List {
			field := &Field{}
			// printutil.PrintNodef(method.Type, "interface method:")
			// fnt := file.ReduceType(method.Type)
			// resfile := fnt.File()
			// resType := fnt.Underlying().(*FuncType)
			// resType.Receiver.Type = res

			field.SetType(NewUnknownType(file, method.Type))
			if method.Names != nil {
				field.SetName(method.Names[0].Name)
			}

			fields = append(fields, field)
		}

		typ.Methods = fields
		return TypeWithFile(typ, file)
	case *ast.MapType:
		key := file.ReduceType(t.Key)
		val := file.ReduceType(t.Value)
		return TypeWithFile(&mapType{key, val}, file)

	case *ast.ChanType:
		log.Fatalf("cannot support ChanType")

	case nil:
		return nil
	default:
		printutil.FatalNodef(expr, "New expr in go")
	}

	return nil
}

func (pkg *Package) ReduceType(expr ast.Expr) (*File, Type) {
	for _, file := range pkg.Files {
		t := file.ReduceType(expr)
		if t != nil {
			return file, t
		}
		// log.Printf("cannot find %s in %s", expr, file.Name)
	}
	return nil, nil
}

func (file *File) ReduceTypeSrc(src string) Type {
	expr, err := ParseExpr(src)
	if err == nil {
		return file.ReduceType(expr)
	}
	log.Printf("expr error: %s:%v", src, err)
	return nil
}

func ResolveUnknownField(field *Field) {

	if unknown, ok := TypeHasUnknown(field.Type); ok {
		resType := unknown.file.ReduceType(unknown.expr)
		if unknown.Parent == nil {
			field.Type = resType
		} else {
			unknown.ResetParent(resType)
		}
	}

	return
}
