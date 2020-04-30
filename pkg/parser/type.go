package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"strings"
)

type Type interface {
	Copy() Type

	File() *File
	Package() *Package
	Name() string

	AssignableTo(Type) bool
	EqualTo(Type) bool

	Underlying() Type
	String() string
}

func IsAtomicType(t Type) bool {
	switch t.(type) {
	case *BasicType, *mapType, *ArrayType:
		return true
	default:
		return false
	}
}

//Define:
//basicType: for golang pre defeined type
type BasicType struct {
	name string
}

func (t *BasicType) Copy() Type {
	return &BasicType{name: t.name}
}
func (t *BasicType) File() *File {
	return nil
}
func (t *BasicType) Package() *Package {
	return nil
}
func (t *BasicType) Name() string {
	return t.name
}
func (t *BasicType) AssignableTo(Type) bool {
	return false
}
func (t *BasicType) EqualTo(tt Type) bool {
	if ttb, ok := tt.(*BasicType); ok {
		return t.name == ttb.name
	}
	return false
}
func (t *BasicType) Underlying() Type {
	return t
}
func (t *BasicType) String() string {
	return t.name
}
func (t *BasicType) DefaultValue() ast.Expr {
	switch t.name {
	case "int", "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64":
		return &ast.BasicLit{Kind: token.INT, Value: "0"}
	case "float32", "float64":
		return &ast.BasicLit{Kind: token.FLOAT, Value: "0"}
	case "complex64", "complex128":
		return &ast.BasicLit{Kind: token.IMAG, Value: "0"}
	case "byte", "rune":
		return &ast.BasicLit{Kind: token.CHAR, Value: "0"}
	case "string":
		return &ast.BasicLit{Kind: token.STRING, Value: ""}
	case "bool":
		return &ast.Ident{Name: "false"}
	case "error", "interface":
		return &ast.Ident{Name: "nil"}
	default:
		log.Fatalf("donot support basic type: %s.DefaultValue()", t.name)
		return nil
	}
}
func NewBasicType(name string) Type {
	return &BasicType{name}
}

//Define:
//PointerType
type PointerType struct {
	Base  Type
	Stars int
}

func (t *PointerType) Copy() Type {
	return &PointerType{
		Base:  t.Base.Copy(),
		Stars: t.Stars,
	}
}
func (t *PointerType) BaseType() Type {
	return t.Base
}
func (t *PointerType) File() *File {
	return t.Base.File()
}
func (t *PointerType) Package() *Package {
	return t.Base.Package()
}
func (t *PointerType) DefaultValue() ast.Expr {
	return &ast.Ident{Name: "nil"}
}
func (t *PointerType) Underlying() Type {
	return t
}
func (t *PointerType) String() string {
	stars := ""
	for i := 0; i < t.Stars; i++ {
		stars = stars + "*"
	}
	return fmt.Sprintf("%s%s", stars, t.Base.String())
}
func (t *PointerType) Name() string {
	return ""
}
func (t *PointerType) AssignableTo(tt Type) bool {
	return t.Base.AssignableTo(tt)
}
func (t *PointerType) EqualTo(tt Type) bool {
	if x, ok := tt.(*PointerType); ok {
		return t.Stars == x.Stars && t.Base.EqualTo(x.Base)
	}
	return false
}

//Define:
//sliceType

type ArrayType struct {
	Element Type
	Slices  int
	Len     int
	lenExpr ast.Expr
}

func (t *ArrayType) Copy() Type {
	return &ArrayType{
		Element: t.Element.Copy(),
		Slices:  t.Slices,
		Len:     t.Len,
		lenExpr: t.lenExpr,
	}
}
func (t *ArrayType) File() *File {
	return t.Element.File()
}
func (t *ArrayType) Package() *Package {
	return t.Element.Package()
}
func (t *ArrayType) DefaultValue() ast.Expr {
	return &ast.Ident{Name: "nil"}
}
func (t *ArrayType) Name() string {
	return ""
}
func (t *ArrayType) Underlying() Type {
	return t
}
func (t *ArrayType) String() string {
	if t.Slices > 0 {
		bkts := ""
		for i := 0; i < t.Slices; i++ {
			bkts = bkts + "[]"
		}

		return fmt.Sprintf("%s%s", bkts, t.Element.String())
	}
	return fmt.Sprintf("[%d]%s", t.Len, t.Element.String())
}
func (t *ArrayType) AssignableTo(tt Type) bool {
	ttu := tt.Underlying()
	if x, ok := ttu.(*ArrayType); ok {
		return t.Slices == x.Slices && t.Element.EqualTo(x.Element)
	}

	return false
}
func (t *ArrayType) EqualTo(tt Type) bool {
	if x, ok := tt.(*ArrayType); ok {
		return t.Slices == x.Slices && t.Element.EqualTo(x.Element)
	}
	return false
}
func (t *ArrayType) LenExpr() ast.Expr {
	if t.lenExpr != nil {
		return t.lenExpr
	}
	return &ast.BasicLit{
		Kind:  token.INT,
		Value: fmt.Sprintf("%d", t.Len),
	}
}

//Define:
//mapType
type mapType struct {
	key Type
	val Type
}

func (t *mapType) Copy() Type {
	return &mapType{
		key: t.key.Copy(),
		val: t.val.Copy(),
	}
}
func (t *mapType) File() *File {
	return nil
}
func (t *mapType) Package() *Package {
	return nil
}
func (t *mapType) Name() string {
	return ""
}
func MapType(key Type, val Type) Type {
	return &mapType{key, val}
}
func (t *mapType) Underlying() Type {
	return t
}
func (t *mapType) String() string {
	return fmt.Sprintf("map[%s]%s", t.key.String(), t.val.String())
}
func (t *mapType) AssignableTo(tt Type) bool {
	if t.EqualTo(tt) {
		return true
	}
	if t.EqualTo(tt.Underlying()) {
		return true
	}

	return false
}
func (t *mapType) EqualTo(tt Type) bool {
	if x, ok := tt.(*mapType); ok {

		return t.key.EqualTo(x.key) && t.val.EqualTo(x.val)
	}
	return false
}

//Define:
//FuncType
type FuncType struct {
	Receiver *Field
	Params   []*Field
	Results  []*Field
}

func (t *FuncType) Copy() Type {
	cp := &FuncType{}
	if t.Receiver != nil {
		cp.Receiver = t.Receiver.Copy().(*Field)
	}

	if t.Params != nil && len(t.Params) > 0 {
		params := make([]*Field, len(t.Params))
		for i, p := range t.Params {
			params[i] = p.Copy().(*Field)
		}
		cp.Params = params
	}
	if t.Results != nil && len(t.Results) > 0 {
		results := make([]*Field, len(t.Results))
		for i, r := range t.Results {
			results[i] = r.Copy().(*Field)
		}
		cp.Results = results
	}

	return cp
}
func (*FuncType) Name() string {
	return ""
}
func (t *FuncType) File() *File {
	return nil
}
func (t *FuncType) Package() *Package {
	return nil
}
func (t *FuncType) AssignableTo(Type) bool {
	return false
}
func (t *FuncType) EqualTo(Type) bool {
	return false
}
func (t *FuncType) Underlying() Type {
	return t
}
func (t *FuncType) String() string {
	var paramStr, resultStr string
	params := []string{}
	if t.Receiver != nil {
		params = append(params, t.Receiver.Type.String())
	}
	if t.Params != nil {
		for _, p := range t.Params {
			params = append(params, p.Type.String())
		}
	}

	paramStr = strings.Join(params, ",")

	results := []string{}
	if t.Results != nil {
		for _, r := range t.Results {
			results = append(results, r.Type.String())
		}
	}
	resultStr = strings.Join(results, ",")
	if len(results) > 1 {
		resultStr = fmt.Sprintf("(%s)", resultStr)
	}

	return fmt.Sprintf("(%s)%s", paramStr, resultStr)
}
func (t *FuncType) DefaultValue() ast.Expr {
	return nil
}

//Define:
//Field
type Field struct {
	Type
	Tag  string
	name string
}

func (t *Field) Copy() Type {
	nf := &Field{
		Type: t.Type.Copy(),
		Tag:  t.Tag,
		name: t.name,
	}
	if nf.Type != nil {
		nf.SetType(nf.Type)
	}
	return nf
}
func (t *Field) IsAnonymous() bool {
	return t.name == ""
}
func (t *Field) IsStruct() bool {
	// struct / *struct
	for x := t.Type.Underlying(); x != nil; x = x.Underlying() {
		if pt, ok := x.(*PointerType); ok {
			x = pt.Base
			continue
		}
		if at, ok := x.(*aliasType); ok {
			x = at.alias
			continue
		}
		if _, ok := x.(*StructType); ok {
			return true
		}
		return false
	}
	return false
}
func (t *Field) Name() string {
	if t.name != "" {
		return t.name
	}

	// anonymous field name is complicated:
	target := t.Type
	for {
		switch udt := target.(type) {
		case *PointerType:
			target = udt.Base
		case *filedType:
			target = udt.Type
		case *namedType:
			return udt.name
		default:
			return target.Name()
		}
	}
}

func (t *Field) FieldName() string {
	return t.name
}
func (t *Field) SetName(name string) {
	t.name = name
}
func (t *Field) SetType(tt Type) {
	if unknown, ok := tt.(*UnknownType); ok {
		unknown.Parent = t
	}

	t.Type = tt
	return
}
func (t *Field) String() string {
	return fmt.Sprintf("%s(%s)", t.name, t.Type)
}

//Define:
//StructType
type StructType struct {
	Fields []*Field
}

func (t *StructType) Copy() Type {
	cp := &StructType{}
	var fields []*Field
	if t.Fields != nil && len(t.Fields) > 0 {
		fields = make([]*Field, len(t.Fields))
		for i, fd := range t.Fields {
			fields[i] = fd.Copy().(*Field)
		}
		cp.Fields = fields
	}

	return cp
}
func (t *StructType) File() *File {
	return nil
}
func (t *StructType) Package() *Package {
	return nil
}

// func (t *StructType) PackageName() string {
//	return ""
// }
func (t *StructType) Name() string {
	return ""
}
func (t *StructType) AssignableTo(tt Type) bool {
	return false
}
func (t *StructType) EqualTo(tt Type) bool {
	return false
}
func (t *StructType) Underlying() Type {
	return t
}
func (t *StructType) String() string {
	return "struct"
}
func (t *StructType) DefaultValue() ast.Expr {
	return nil
}

//Define:
//packageType
type packageType struct {
	file *File
	name string
	path string
}

func (t *packageType) Copy() Type {
	return &packageType{
		file: t.file,
		name: t.name,
		path: t.path,
	}
}
func (t *packageType) File() *File {
	return t.file
}
func (t *packageType) Package() *Package {
	return nil
}
func (t *packageType) Name() string {
	return t.name
}
func (t *packageType) AssignableTo(Type) bool {
	return false
}
func (t *packageType) EqualTo(tt Type) bool {
	if ptt, ok := tt.(*packageType); ok {
		return ptt.path == t.path
	}
	return false
}
func (t *packageType) Underlying() Type {
	return t
}
func (t *packageType) String() string {
	return t.path
}
func (t *packageType) DefaultValue() ast.Expr {
	return nil
}

//Define:
//InterfaceType
type InterfaceType struct {
	Methods []*Field
}

func (t *InterfaceType) Copy() Type {
	cp := &InterfaceType{}
	methods := []*Field{}
	if t.Methods != nil && len(t.Methods) > 0 {
		methods = make([]*Field, len(t.Methods))
		for i, m := range t.Methods {
			methods[i] = m.Copy().(*Field)
		}
	}

	cp.Methods = methods
	return cp
}
func (t *InterfaceType) Name() string {
	return ""
}
func (t *InterfaceType) AssignableTo(tt Type) bool {
	return t.EqualTo(tt)
}
func (t *InterfaceType) EqualTo(tt Type) bool {
	if target, ok := tt.(*InterfaceType); ok {
		//TODO
		if len(target.Methods) != len(t.Methods) {
			return false
		}

		for _, method := range t.Methods {
			finded := false
			for _, other := range target.Methods {
				if other.EqualTo(method) {
					finded = true
					break
				}
			}

			if !finded {
				return false
			}
		}

		return true
	}
	return false
}

func (t *InterfaceType) Underlying() Type {
	return t
}
func (t *InterfaceType) String() string {
	if len(t.Methods) == 0 {
		return "interface{}"
	}
	return ""
}
func (t *InterfaceType) File() *File {
	return nil
}
func (t *InterfaceType) Package() *Package {
	return nil
}
func (t *InterfaceType) DefaultValue() ast.Expr {
	return nil
}
func (t *InterfaceType) GetFuncByName(name string) Type {
	if t.Methods == nil || len(t.Methods) == 0 {
		return nil
	}
	for _, mt := range t.Methods {
		if mt.name == name {
			ResolveUnknownField(mt)
			nt := TypeWithFile(TypeWithName(mt.Type, mt.name), mt.File())
			//&namedType{name: mt.name, file: t.file}
			//nt.SetUnderlying(mt.Type)
			// log.Printf("get func by name: %s : %s", nt.Name(), nt)
			return nt
		}
	}
	return nil
}

//@TODO:
type chanType struct {
	base     Type
	sender   bool
	Receiver bool
}

//wrapper types:

//Define wrapper:
//UnknownType
type UnknownType struct {
	file   *File
	Parent Type
	expr   ast.Expr
}

func (t *UnknownType) Copy() Type {
	return &UnknownType{
		file:   t.file,
		Parent: t.Parent,
		expr:   t.expr,
	}
}
func (t *UnknownType) WithFile(f *File) *UnknownType {
	t.file = f
	return t
}
func (unknown *UnknownType) Reduce(file *File) Type {
	resType := file.ReduceType(unknown.expr)
	unknown.ResetParent(resType)
	return resType
}
func (unknown *UnknownType) ResetParent(kn Type) {
	if unknown.Parent == nil {
		return
	}
	switch p := unknown.Parent.(type) {
	case *namedType:
		p.SetUnderlying(kn)
	case *aliasType:
		p.SetAlias(kn)
	case *Field:
		p.SetType(kn)
	default:
		log.Fatalf("unsupported parent %s", unknown.Parent)
	}
}
func (t *UnknownType) Name() string {
	return ""
}
func (t *UnknownType) AssignableTo(Type) bool {
	return false
}
func (t *UnknownType) EqualTo(Type) bool {
	return false
}
func (t *UnknownType) Underlying() Type {
	return t
}
func (t *UnknownType) String() string {
	return "!!Unknown!!"
}
func (t *UnknownType) File() *File {
	return t.file
}
func (t *UnknownType) Package() *Package {
	return t.file.BelongTo
}
func (t *UnknownType) DefaultValue() ast.Expr {
	return nil
}

//Define wrapper:
//namedType
type namedType struct {
	Type
	name string
}

func (t *namedType) Copy() Type {
	cp := &namedType{name: t.name}
	if t.Type != nil {
		tcp := t.Type.Copy()
		cp.SetUnderlying(tcp)
	}
	return cp
}
func (t *namedType) File() *File {
	if t.Type == nil {
		return nil
	}

	return t.Type.File()
}
func (t *namedType) Package() *Package {
	return nil
}
func (t *namedType) Name() string {
	return t.name
}
func (t *namedType) AssignableTo(tt Type) bool {
	if t.EqualTo(tt) {
		return true
	}
	if t.Type != nil && t.Type.EqualTo(tt) {
		return true
	}

	return false
}
func (t *namedType) EqualTo(tt Type) bool {
	var ot Type = tt
	if ft, ok := tt.(*filedType); ok {
		ot = ft.Type
	}
	if ttu, ok := ot.(*namedType); ok {
		return t.name == ttu.name
	}
	return false
}
func (t *namedType) Underlying() Type {
	if t.Type != nil {
		return t.Type.Underlying()
	}

	return nil
}
func (t *namedType) SetUnderlying(ut Type) {
	t.Type = ut
	if ut == nil {
		return
	}
	if unknown, ok := ut.(*UnknownType); ok {
		unknown.Parent = t
	}
}
func (t *namedType) String() string {
	name := t.name
	if ft, ok := t.Type.(*FuncType); ok {
		return fmt.Sprintf("%s%s", name, ft)
	}

	// return fmt.Sprintf("%s %s", name, t.Type)

	return name
}

//Define wrapper:
//aliasType
type aliasType struct {
	alias Type
	name  string
}

func (t *aliasType) Copy() Type {
	cp := &aliasType{name: t.name}
	if t.alias != nil {
		tcp := t.alias.Copy()
		cp.SetAlias(tcp)
	}
	return cp
}
func (t *aliasType) SetAlias(at Type) {
	t.alias = at
	if at == nil {
		return
	}
	if unknown, ok := at.(*UnknownType); ok {
		unknown.Parent = t
	}

}
func (t *aliasType) File() *File {
	return nil //t.file
}
func (t *aliasType) Package() *Package {
	return nil //t.file.BelongTo
}
func (t *aliasType) Name() string {
	return t.name
}
func (t *aliasType) AssignableTo(tt Type) bool {
	if utt, ok := tt.(*aliasType); ok {
		return t.alias.EqualTo(utt.alias)
	}
	return t.alias.EqualTo(tt)
}
func (t *aliasType) EqualTo(tt Type) bool {
	if ttu, ok := tt.(*aliasType); ok {
		return t.alias.EqualTo(ttu.alias)
	}
	return t.alias.EqualTo(tt)
}
func (t *aliasType) Underlying() Type {
	return t.alias.Underlying()
}
func (t *aliasType) String() string {
	return t.name
}

//Define wrapper:
//exprType
type exprType struct {
	Type
	expr ast.Expr
}

func (t *exprType) Copy() Type {
	cp := &exprType{expr: t.expr}
	if t.Type != nil {
		tcp := t.Type.Copy()
		cp.Type = tcp
	}
	return cp
}

//Define wrapper:
//filedType
type filedType struct {
	Type
	file *File
}

func (t *filedType) Copy() Type {
	cp := &filedType{file: t.file}
	if t.Type != nil {
		cp.Type = t.Type.Copy()
	}
	return cp
}
func (t *filedType) File() *File {
	return t.file
}
func (t *filedType) Package() *Package {
	return t.file.BelongTo
}
func (t *filedType) EqualTo(tt Type) bool {
	if IsAtomicType(t.Type) && IsAtomicType(tt) {
		return t.Type.EqualTo(tt)
	}
	var ot Type = tt
	if x, ok := tt.(*filedType); ok {

		// if basic,map,chan,array
		if IsAtomicType(t.Type) && IsAtomicType(x.Type) {
			return t.Type.EqualTo(x.Type)
		}

		tPkg := t.file.BelongTo
		xPkg := x.file.BelongTo

		//TODO: same package compare:
		if ok := xPkg.EqualTo(tPkg); !ok {
			// }|| xPkg.Dir != tPkg.Dir {
			return false
		}
		ot = x
	}
	if t.Type == nil {
		return false
	}

	return t.Type.EqualTo(ot)
}
func (t *filedType) String() string {
	if t.file == nil {
		return t.Type.String()
	}
	// log.Printf("======t:%s, package: %s", t.Type, t.file.BelongTo.Name)
	return fmt.Sprintf("%s.%s", t.file.BelongTo.PackageName, t.Type.String())
}
func (t *filedType) PackageNameInFile(file *File) string {
	tPkg := t.file.BelongTo
	fPkg := file.BelongTo
	if tPkg != nil && !tPkg.EqualTo(fPkg) {
		pkgName := file.FindImportNameByPath(tPkg.CanonicalPath)
		if pkgName == "" {
			pkgName = file.FindImportNameByPath(tPkg.Path)
		}
		if pkgName == "" {
			log.Fatalf("package %s(%s) not imported in file(%s)", tPkg.Path, tPkg.CanonicalPath, file.Name)
		}
		return pkgName
	}
	return ""
}

func TypeEqual(a, b Type) bool {
	if field, ok := a.(*Field); ok {
		return TypeEqual(field.Type, b)
	}

	if field, ok := b.(*Field); ok {
		return TypeEqual(a, field.Type)
	}
	return a.EqualTo(b)
}

func TypeAssignable(target, source Type) bool {
	if field, ok := target.(*Field); ok {
		return TypeAssignable(field.Type, source)
	}

	if field, ok := source.(*Field); ok {
		return TypeAssignable(target, field.Type)
	}

	if basic, ok := target.(*BasicType); ok && basic.name == "interface" {
		return true
	}
	if inf, ok := target.(*InterfaceType); ok && len(inf.Methods) == 0 {
		return true
	}

	// TODO: InterfaceType

	return target.EqualTo(source)
}
