package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strings"

	"github.com/lawrsp/pigo/pkg/printutil"
)

func NewType(name string) Type {
	switch name {
	case "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64":
		fallthrough
	case "float32", "float64":
		fallthrough
	case "complex64", "complex128":
		fallthrough
	case "byte", "rune":
		fallthrough
	case "string":
		fallthrough
	case "error", "interface":
		fallthrough
	case "bool":
		return &BasicType{name: name}
	default:
		nt := &namedType{name: name}
		nt.SetUnderlying(NewUnknownType(nil, nil))
		return nt
	}
}

var errorType = &BasicType{name: "error"}

func ErrorType() Type {
	return errorType
}

func NewUnknownType(file *File, expr ast.Expr) *UnknownType {
	return &UnknownType{
		expr: expr,
		file: file,
	}
}

func NewField(t Type, name string, tag string) *Field {
	return &Field{Type: t, Tag: tag, name: name}
}

func typeWithPointer(x Type) Type {
	if pt, ok := x.Underlying().(*PointerType); ok {
		pt.Stars += 1
		return x
	}
	return &PointerType{
		Base:  x,
		Stars: 1,
	}
}
func TypeWithPointer(t Type) Type {
	x := t.Copy()
	return typeWithPointer(x)
}

func typeWithSlice(x Type) Type {
	if at, ok := x.(*ArrayType); ok {
		if at.Slices > 0 {
			at.Slices += 1
			return x
		}
	}

	return &ArrayType{
		Element: x,
		Slices:  1,
	}
}
func TypeWithSlice(t Type) Type {
	x := t.Copy()
	return typeWithSlice(x)
}

func typeWithArray(x Type, len int) Type {
	return &ArrayType{
		Element: x,
		Len:     len,
	}
}
func TypeWithArray(t Type, len int) Type {
	x := t.Copy()
	return typeWithArray(x, len)
}

// func TypeWithStar(base Type) Type {
// 	//TODO: check? named ?
// 	res := base
// 	var file *File = nil
// 	if x, ok := res.(*filedType); ok {
// 		res = x.Type
// 		file = x.file
// 	}
// 	if x, ok := res.(*PointerType); ok {
// 		res = &PointerType{x.Base, x.Stars + 1}
// 	} else {
// 		res = &PointerType{base, 1}
// 	}
// 	if file != nil {
// 		return TypeWithFile(res, file)
// 	}

// 	return res
// }

func TypeSkipPointer(t Type, n int) Type {
	//TODO: check? named ?
	m := t.Copy()
	ptr, ok := m.Underlying().(*PointerType)
	if !ok {
		return m
	}

	ptr.Stars -= n

	if ptr.Stars == 0 {
		return ptr.Base
	}

	return m
}

func TypeSkipBracket(t Type, n int) Type {
	//TODO: check? named ?
	m := t.Copy()
	for i := 0; i < n; i++ {
		switch x := m.Underlying().(type) {
		case *ArrayType:
			if x.Slices > 0 {
				x.Slices -= 1
			}
			if x.Slices == 0 {
				m = x.Element
			}
		default:
			return m
		}
	}

	return m
}

// func TypeWithSlice(t Type) Type {
// 	res := t
// 	var file *File = nil
// 	if x, ok := res.(*filedType); ok {
// 		res = x.Type
// 		file = x.file
// 	}
// 	if x, ok := res.(*sliceType); ok {
// 		res = &sliceType{x.element, x.brackets + 1}
// 	} else {
// 		res = &sliceType{t, 1}
// 	}
// 	if file != nil {
// 		return TypeWithFile(res, file)
// 	}
// 	return res
// }

func TypeWrapStruct(t Type, name string) Type {
	fd := NewField(t, name, "")
	return &StructType{
		Fields: []*Field{fd},
	}
}

func GetInterfaceFuncByName(t Type, name string) Type {
	if ift, ok := t.Underlying().(*InterfaceType); ok {
		return ift.GetFuncByName(name)
	}
	return nil
}

func TypeHasUnknown(t Type) (*UnknownType, bool) {
	for udt := t.Underlying(); udt != nil; {
		switch rt := udt.(type) {
		case *PointerType:
			udt = rt.Base.Underlying()
		case *ArrayType:
			udt = rt.Element.Underlying()
			// log.Printf("?????rt.Element: %s, underlying: %s", rt, udt)
		case *UnknownType:
			return rt, true
		default:
			if udt.Underlying() == udt {
				return nil, false
			}
			udt = udt.Underlying()
		}
	}

	// log.Printf("==========no unkown, %s", t)
	return nil, false
}

func TypeIsUnknown(t Type) bool {
	_, ok := t.(*UnknownType)
	return ok
}

func TypeEqualAsReceiver(t1 Type, t2 Type) bool {
	if ft, ok := t1.(*filedType); ok {
		t1 = ft
	}
	if ft, ok := t2.(*filedType); ok {
		t2 = ft
	}

	if star1, ok := t1.(*PointerType); ok {
		if star2, ok := t2.(*PointerType); ok {
			if star2.Base.EqualTo(star1.Base) {
				return true
			} else {
				return false
			}
		} else {
			return star1.Base.EqualTo(t2)
		}
	} else if star2, ok := t2.(*PointerType); ok {
		return star2.Base.EqualTo(t1)
	}
	return t1.EqualTo(t2)
}

func TypeWithName(underlying Type, name string) Type {
	// var file *File = nil

	nt := &namedType{name: name}
	nt.SetUnderlying(underlying)
	return nt
}

func TypeWithExpr(t Type, expr ast.Expr) Type {
	return &exprType{t, expr}
}

func TypeWithFile(t Type, file *File) Type {
	switch ft := t.(type) {
	case *BasicType:
		return t
	case *filedType:
		return &filedType{ft.Type, file}
	case *ArrayType:
		return t
	case *StructType:
		return t
	case *PointerType:
		return t
	case *mapType:
		return t
	case *FuncType:
		return t
	case *InterfaceType:
		return t
	case *UnknownType:
		return t
		// default:
		//	if t.Underlying() == t {
		//		return t
		//	}
	}

	// log.Printf("add t %s file", t)
	// if t.File() != nil {
	//	log.Printf("t %s, file: %s", t, t.File().Name)
	//	return t
	// }
	return &filedType{t, file}
}

func GetTypeStars(t Type) int {
	if x, ok := t.Underlying().(*PointerType); ok {
		return x.Stars
	}
	return 0
}

func TypeExprInFile(t Type, file *File) ast.Expr {
	if t == nil {
		return nil
	}
	if _, ok := t.(*UnknownType); ok {
		log.Fatalf("field is unknown, there is a bug in ReduceType")
		return nil
	}

	if _, ok := t.(*Field); ok {
		log.Fatalf("field Type %s cannot change to Expr", t)
		return nil
	}
	if et, ok := t.(*exprType); ok {
		return et.expr
	}
	if pt, ok := t.(*packageType); ok {
		return ast.NewIdent(pt.name)
	}
	var pkgName string
	var ot = t
	if ft, ok := t.(*filedType); ok && file != nil {
		pkgName = ft.PackageNameInFile(file)
		ot = ft.Type
	}

	switch tm := ot.(type) {
	case *BasicType:
		if tm.name == "interface" {
			return &ast.CompositeLit{
				Type: ast.NewIdent(tm.name),
			}
		}
		return ast.NewIdent(tm.name)
	case *PointerType:
		var result = &ast.StarExpr{}
		pointer := result
		for i := 1; i < tm.Stars; i++ {
			n := &ast.StarExpr{}
			pointer.X = n
			pointer = n
		}

		pointer.X = TypeExprInFile(tm.Base, file)
		return result
	case *ArrayType:
		if tm.Slices == 0 {
			return &ast.ArrayType{
				Len: tm.LenExpr(),
				Elt: TypeExprInFile(tm.Element, file),
			}
		}

		result := &ast.ArrayType{}
		var mid *ast.ArrayType = result

		for i := 1; i < tm.Slices; i++ {
			mid.Elt = &ast.ArrayType{}
			mid = mid.Elt.(*ast.ArrayType)
		}
		mid.Elt = TypeExprInFile(tm.Element, file)
		return result
	case *mapType:
		return &ast.MapType{
			Key:   TypeExprInFile(tm.key, file),
			Value: TypeExprInFile(tm.val, file),
		}
	case *FuncType:
		fnt := &ast.FuncType{}
		if tm.Params != nil && len(tm.Params) > 0 {
			params := &ast.FieldList{}
			params.List = make([]*ast.Field, len(tm.Params))
			for i, field := range tm.Params {
				ResolveUnknownField(field)
				astField := &ast.Field{}
				if field.name != "" {
					astField.Names = []*ast.Ident{ast.NewIdent(field.name)}
				}
				astField.Type = TypeExprInFile(field.Type, file)
				params.List[i] = astField
			}
			fnt.Params = params
		}
		if tm.Results != nil && len(tm.Results) > 0 {
			results := &ast.FieldList{}
			results.List = make([]*ast.Field, len(tm.Results))
			for i, field := range tm.Results {
				ResolveUnknownField(field)
				astField := &ast.Field{}
				if field.name != "" {
					astField.Names = []*ast.Ident{ast.NewIdent(field.name)}
				}
				astField.Type = TypeExprInFile(field.Type, file)
				results.List[i] = astField
			}
			fnt.Results = results
		}
		return fnt
	case *StructType:
		list := &ast.FieldList{}
		for _, field := range tm.Fields {
			ResolveUnknownField(field)
			fd := &ast.Field{}
			if field.name != "" {
				fd.Names = []*ast.Ident{ast.NewIdent(field.name)}
			}
			fd.Type = TypeExprInFile(field.Type, file)
		}
		return &ast.StructType{Fields: list}
	case *InterfaceType:
		expr := &ast.InterfaceType{}
		if tm.Methods != nil && len(tm.Methods) > 0 {
			methods := &ast.FieldList{}
			methods.List = make([]*ast.Field, len(tm.Methods))
			for i, md := range tm.Methods {
				ResolveUnknownField(md)
				if md.name != "" {
					methods.List[i].Names = []*ast.Ident{ast.NewIdent(md.name)}
				}
				methods.List[i].Type = TypeExprInFile(md.Type, file)
			}
			expr.Methods = methods
		}
		return expr
	case *namedType:
		if pkgName == "" {
			return ast.NewIdent(tm.name)
		}
		return &ast.SelectorExpr{
			X:   ast.NewIdent(pkgName),
			Sel: ast.NewIdent(tm.name),
		}

	case *aliasType:
		if pkgName == "" {
			return ast.NewIdent(tm.name)
		}
		return &ast.SelectorExpr{
			X:   ast.NewIdent(pkgName),
			Sel: ast.NewIdent(tm.name),
		}
	}

	log.Fatalf("donot support expr for type: %s", t)
	return nil
}

func TypeZeroValue(t Type, file *File) ast.Expr {
	if t == nil {
		return nil
	}
	if _, ok := t.(*UnknownType); ok {
		log.Fatalf("field is unknown, there is a bug in ReduceType")
		return nil
	}

	if _, ok := t.(*Field); ok {
		log.Fatalf("field Type %s cannot change to Expr", t)
		return nil
	}
	if et, ok := t.(*exprType); ok {
		return et.expr
	}
	if pt, ok := t.(*packageType); ok {
		return ast.NewIdent(pt.name)
	}
	ot := t
	if ft, ok := t.(*filedType); ok && file != nil {
		ot = ft.Type
	}

	switch tm := ot.(type) {
	case *BasicType:
		return tm.DefaultValue()
	case *ArrayType, *mapType, *PointerType, *FuncType, *InterfaceType:
		return ast.NewIdent("nil")
	case *namedType, *aliasType:
		if basic, ok := tm.Underlying().(*BasicType); ok {
			return &ast.CallExpr{
				Fun:  TypeExprInFile(t, file),
				Args: []ast.Expr{basic.DefaultValue()},
			}
		} else {
			return &ast.CompositeLit{
				Type: TypeExprInFile(t, file),
			}
		}
		// case *StructType:
		// 	log.Fatalf("tm is StructType")
		// default:
		// 	log.Printf("tm is :%v(%v)", tm)
	}

	log.Fatalf("donot support zero value for type: %s", t)
	return nil
}

func TypeInitValue(t Type, file *File) ast.Expr {
	if t == nil {
		return nil
	}
	if _, ok := t.(*UnknownType); ok {
		log.Fatalf("field is unknown, there is a bug in ReduceType")
		return nil
	}

	if _, ok := t.(*Field); ok {
		log.Fatalf("field Type %s cannot change to Expr", t)
		return nil
	}
	if et, ok := t.(*exprType); ok {
		return et.expr
	}
	if pt, ok := t.(*packageType); ok {
		return ast.NewIdent(pt.name)
	}
	ot := t
	if ft, ok := t.(*filedType); ok && file != nil {
		ot = ft.Type
	}

	switch tm := ot.(type) {
	case *BasicType:
		return tm.DefaultValue()
	case *ArrayType, *mapType:
		return &ast.CompositeLit{
			Type: TypeExprInFile(t, file),
		}
	case *PointerType:
		if tm.Stars > 1 {
			return ast.NewIdent("nil")
		}

		if _, ok := tm.Base.Underlying().(*BasicType); ok {
			return ast.NewIdent("nil")
		}
		return &ast.UnaryExpr{
			Op: token.AND,
			X:  TypeZeroValue(tm.Base, file),
		}
	case *FuncType, *InterfaceType:
		return ast.NewIdent("nil")
	case *namedType, *aliasType:
		log.Printf("type: %s", t)
		if basic, ok := tm.Underlying().(*BasicType); ok {
			return &ast.CallExpr{
				Fun:  TypeExprInFile(t, file),
				Args: []ast.Expr{basic.DefaultValue()},
			}
		} else {
			return &ast.CompositeLit{
				Type: TypeExprInFile(t, file),
			}
		}
		// case *StructType:
	}

	log.Fatalf("donot support init value for type: %s", t)
	return nil
}

func NotNilPointerValue(t *PointerType, file *File) ast.Expr {
	if t.Stars != 1 {
		return nil
	}
	if _, ok := t.Base.Underlying().(*BasicType); ok {
		return nil
	}
	return &ast.UnaryExpr{
		Op: token.AND,
		X:  TypeZeroValue(t.Base, file),
	}
}

func fieldListToString(fieldList *ast.FieldList, isStruct bool) string {
	if fieldList == nil {
		return ""
	}
	fields := []string{}
	for _, field := range fieldList.List {
		idents := []string{}
		for _, ident := range field.Names {
			idents = append(idents, ExprToString(ident))
		}

		f := fmt.Sprintf("%s %s", strings.Join(idents, ","), ExprToString(field.Type))
		fields = append(fields, f)
	}

	if isStruct {
		return strings.Join(fields, "\n")
	}
	return strings.Join(fields, ",")
}

func ExprToString(expr ast.Expr) string {
	if expr == nil {
		return ""
	}

	switch v := expr.(type) {
	case *ast.BadExpr:
		return "BAD!!"
	case *ast.Ident:
		return v.Name
	case *ast.Ellipsis:
		return "..."
	case *ast.BasicLit:
		return v.Value
	case *ast.FuncLit:
		if v.Type != nil {
			return ExprToString(v.Type)
		}
		return "func()"
	case *ast.CompositeLit:
		elts := ""
		if len(v.Elts) > 0 {
			strs := []string{}
			for _, e := range v.Elts {
				str := ExprToString(e)
				strs = append(strs, str)
			}
			elts = strings.Join(strs, ",")
		}
		return fmt.Sprintf("%s{%s}", ExprToString(v.Type), elts)
	case *ast.ParenExpr:
		return fmt.Sprintf("(%s)", ExprToString(v.X))
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", ExprToString(v.X), ExprToString(v.Sel))
	case *ast.IndexExpr:
		return fmt.Sprintf("%s[%s]", ExprToString(v.X), ExprToString(v.Index))
	case *ast.SliceExpr:
		if v.Slice3 {
			return fmt.Sprintf("%s[%s:%s:%s]", ExprToString(v.X), ExprToString(v.Low), ExprToString(v.High), ExprToString(v.Max))
		} else {
			return fmt.Sprintf("%s[%s:%s]", ExprToString(v.X), ExprToString(v.Low), ExprToString(v.High))
		}
	case *ast.TypeAssertExpr:
		return fmt.Sprintf("%s.(%s)", ExprToString(v.X), ExprToString(v.Type))
	case *ast.CallExpr:
		args := ""
		if len(v.Args) > 0 {
			arglist := []string{}
			for _, a := range v.Args {
				as := ExprToString(a)
				arglist = append(arglist, as)
			}
			args = strings.Join(arglist, ",")
		}

		return fmt.Sprintf("%s(%s)", ExprToString(v.Fun), args)
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", ExprToString(v.X))
	case *ast.UnaryExpr:
		return fmt.Sprintf("%s%s", v.Op.String(), ExprToString(v.X))
	case *ast.BinaryExpr:
		return fmt.Sprintf("%s%s%s", ExprToString(v.X), v.Op.String(), ExprToString(v.Y))
	case *ast.KeyValueExpr:
		return fmt.Sprintf("%s:%s", ExprToString(v.Key), ExprToString(v.Value))
	case *ast.ArrayType:
		return fmt.Sprintf("[%s]%s", ExprToString(v.Len), ExprToString(v.Elt))
	case *ast.StructType:
		fieldList := v.Fields
		if fieldList == nil {
			return "struct{}"
		}
		fields := fieldListToString(v.Fields, true)
		return fmt.Sprintf("struct{%s}", fields)
	case *ast.FuncType:
		params := fieldListToString(v.Params, false)
		results := fieldListToString(v.Results, false)
		if v.Results != nil && v.Results.List != nil && len(v.Results.List) >= 2 {
			return fmt.Sprintf("func (%s) (%s)", params, results)
		} else {
			return fmt.Sprintf("func (%s) %s", params, results)
		}
	case *ast.InterfaceType:
		fields := fieldListToString(v.Methods, true)
		return fmt.Sprintf("interface{%s}", fields)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", ExprToString(v.Key), ExprToString(v.Value))
	case *ast.ChanType:
		if v.Dir == ast.SEND {
			return fmt.Sprintf("->%s", ExprToString(v.Value))
		} else {
			return fmt.Sprintf("<-%s", ExprToString(v.Value))
		}
	}

	return ""
}

/*










 */

func ParseType(node ast.Node) Type {
	var typ Type

	switch ident := node.(type) {
	//Exprs:
	case *ast.BadExpr:
		log.Fatalf("cannot support BadExpr")
	case *ast.Ident:
		typ = NewType(ident.Name)
	case *ast.Ellipsis:
		log.Fatalf("cannot support Ellipsis")
	case *ast.BasicLit:
		log.Fatalf("cannot support BasicLit")
	case *ast.FuncLit:
		log.Fatalf("cannot support FuncLit")
	case *ast.CompositeLit:
		typ = ParseType(ident.Type)
	case *ast.ParenExpr:
		log.Fatalf("cannot support ParenExpr")
	case *ast.SelectorExpr:
		typ = ParseType(ident.Sel)
		if _, ok := ident.X.(*ast.Ident); ok {
			// typ.SetPackageName(ni.Name)
		} else {
			printutil.FatalNodef(node, "cannot support parse type:")
		}
	case *ast.IndexExpr:
		log.Fatalf("cannot support IndexExpr")
	case *ast.SliceExpr:
		log.Fatalf("cannot support SliceExpr")
	case *ast.TypeAssertExpr:
		log.Fatalf("cannot support TypeAssertExpr")
	case *ast.CallExpr:
		log.Fatalf("cannot support CallExpr")
	case *ast.StarExpr:
		typ2 := ParseType(ident.X)
		typ = TypeWithPointer(typ2)
	case *ast.UnaryExpr:
		typ = ParseType(ident.X)
		if ident.Op == token.MUL || ident.Op == token.AND {
			typ = TypeWithPointer(typ)
		}
	case *ast.BinaryExpr:
		log.Fatalf("cannot support BinaryExpr")
	case *ast.KeyValueExpr:
		log.Fatalf("cannot support KeyValueExpr")
	case *ast.ArrayType:
		typ = ParseType(ident.Elt)
		typ = TypeWithSlice(typ)
		//Types:
	case *ast.StructType:
		log.Fatalf("cannot support StructType")
	case *ast.FuncType:
		log.Fatalf("cannot support FuncType")
	case *ast.InterfaceType:
		typ = NewType("interface")
	case *ast.MapType:
		key := ParseType(ident.Key)
		val := ParseType(ident.Value)
		typ = MapType(key, val)
	case *ast.ChanType:
		log.Fatalf("cannot support ChanType")

		//others:
	case *ast.TypeSpec:
		typ = TypeWithName(ParseType(ident.Type), ident.Name.Name)
	case *ast.ValueSpec:
		if ident.Type != nil {
			typ = ParseType(ident.Type)
		} else {
			typ = ParseType(ident.Values[0])
		}
	default:
		printutil.FatalNodef(node, "ParseType donnot Support:")
	}

	// ast.Print(token.NewFileSet(), expr)
	return typ
}

func IsErrorType(t Type) bool {
	if x, ok := t.(*namedType); ok {
		return x.name == "error"
	}
	return false
}

func ParseTypeString(str string) Type {
	expr, err := ParseExpr(str)
	if err != nil {
		log.Fatalf("parse type %s error: %v", str, err)
	}

	return ParseType(expr)
}
