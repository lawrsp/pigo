package builder

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"

	"github.com/lawrsp/pigo/generator/parser"
)

type Mode int

const (
	ANY_MODE = iota
	READ_MODE
	WRITE_MODE
	RW_MODE
)

func (t Mode) Contains(m Mode) bool {
	return int(t)&int(m) != 0
}

type Variable struct {
	Type      parser.Type
	isVisible bool
	name      string
	mode      Mode
	expr      ast.Expr
}

func NewVariable(pt parser.Type) *Variable {
	v := &Variable{
		Type: pt,
	}

	//@NOTICE: may some type need a expr not a Name

	return v
}

func (v *Variable) Copy() *Variable {
	nv := &Variable{}
	*nv = *v
	nv.Type = v.Type.Copy()
	return nv
}

func (v *Variable) AutoName() *Variable {
	if v.name != "" {
		return v
	}
	if v.Type != nil {
		if name := v.Type.Name(); name != "" {
			v.name = VariableName(name)
		} else {
			v.name = VariableName(v.Type.String())
		}
	}
	if v.name == "" {
		v.name = "var1"
	}
	return v
}

func (v *Variable) WithName(name string) *Variable {
	v.name = name
	return v
}
func (v *Variable) Name() string {
	return v.name
}
func (v *Variable) SetName(name string) {
	v.name = name
}

func (v *Variable) ReadOnly() *Variable {
	v.mode = READ_MODE
	return v
}
func (v *Variable) WriteOnly() *Variable {
	v.mode = WRITE_MODE
	return v
}
func (v *Variable) ReadWrite() *Variable {
	v.mode = RW_MODE
	return v
}
func (v *Variable) WithMode(mode Mode) *Variable {
	v.mode = mode
	return v
}
func (v *Variable) Mode() Mode {
	return v.mode
}
func (v *Variable) SetMode(mode Mode) {
	v.mode = mode
}

func (v *Variable) IsAnonymous() bool {
	return v.name == "" && v.expr != nil
}
func (v *Variable) IsReadable() bool {
	return (v.mode & READ_MODE) != 0
}
func (v *Variable) IsWriteable() bool {
	return (v.mode & WRITE_MODE) != 0
}
func (v *Variable) IsReadOnly() bool {
	return v.mode == READ_MODE
}
func (v *Variable) IsWriteOnly() bool {
	return v.mode == WRITE_MODE
}
func (v *Variable) IsVisible() bool {
	if v.IsAnonymous() {
		return true
	}
	return v.isVisible
}

func (v *Variable) Ident() ast.Expr {
	if v.name != "" {
		return NewAstIdent(v.name)
	}
	if v.expr != nil {
		return v.expr
	}

	// panic("variable has no name and no expr")
	log.Fatalf("variable has no name and no expr")
	return nil
}

func (v *Variable) Expr() ast.Expr {
	return v.expr
}
func (v *Variable) WithExpr(expr ast.Expr) *Variable {
	v.expr = expr
	return v
}

func (v *Variable) SetExpr(expr ast.Expr) {
	v.expr = expr
}
func (v *Variable) SetExprSrc(str string) {
	expr, err := parser.ParseExpr(str)
	if err != nil {
		log.Fatalf("expr %s parse error: %v", str, err)
	}

	v.expr = expr
}

func (v *Variable) IncreaseName() {
	v.name = IncreaseName(v.name)
}

func (v *Variable) AddressExpr() ast.Expr {
	return &ast.UnaryExpr{
		Op: token.AND,
		X:  v.Ident(),
	}
}
func (v *Variable) TypeConversionExpr(typ ast.Expr) ast.Expr {
	return TypeConversionExpr(typ, v.Ident())
}
func (v *Variable) PointerValueExpr() ast.Expr {
	return &ast.StarExpr{
		X: v.Ident(),
	}
}

func (v *Variable) CheckEqualExpr(isEqual bool, y ast.Expr) ast.Expr {
	var op token.Token
	if isEqual {
		op = token.EQL
	} else {
		op = token.NEQ
	}

	return &ast.BinaryExpr{
		X:  v.Ident(),
		Op: op,
		Y:  y,
	}
}

func (v *Variable) CheckNilExpr(isNil bool) ast.Expr {
	return v.CheckEqualExpr(isNil, ast.NewIdent("nil"))
}

func (v *Variable) IndexExpr(index ast.Expr) ast.Expr {
	return &ast.IndexExpr{
		X:     v.Ident(),
		Index: index,
	}
}

func (v *Variable) StringKeyItemExpr(key string) ast.Expr {
	return v.IndexExpr(&ast.BasicLit{
		Kind:  token.STRING,
		Value: fmt.Sprintf("\"%s\"", key),
	})
}

func (v *Variable) DotExpr(name string) ast.Expr {
	if v.name == "" {
		return nil
	}

	dot, err := parser.ParseExpr(name)
	if err != nil {
		log.Fatalf("%s dot %s error: %v", v.name, name, err)
	}

	return DotExpr(ast.NewIdent(v.name), dot)
}

func (v *Variable) DotTypeInFile(dotSrc string, file *parser.File) parser.Type {
	dot, err := parser.ParseExpr(dotSrc)
	if err != nil {
		log.Fatalf("%s dot %s error: %v", v.name, dotSrc, err)
	}

	expr := parser.TypeExprInFile(v.Type, file)
	expr = DotExpr(expr, dot)

	typ := file.ReduceType(expr)

	return typ
}

func (v *Variable) DotVariable(dotString string, file *parser.File) *Variable {
	expr := v.DotExpr(dotString)
	typ := v.DotTypeInFile(dotString, file)

	return NewVariable(typ).WithExpr(expr).ReadOnly()
}

type VariableList struct {
	List []*Variable
}

func NewVariableList() *VariableList {
	return &VariableList{
		List: []*Variable{},
	}
}

func (vl *VariableList) ForEach(run func(int, *Variable)) {
	for i, v := range vl.List {
		run(i, v)
	}
}

func (vl *VariableList) IsEmpty() bool {
	return vl.List == nil || len(vl.List) == 0
}

func (vl *VariableList) Concat(vl2 *VariableList) *VariableList {
	newVl := NewVariableList()
	for _, v := range vl.List {
		newVl.Add(v)
	}
	for _, v := range vl2.List {
		newVl.Add(v)
	}
	return newVl
}

func (vl *VariableList) GetByType(pt parser.Type, mode Mode) *Variable {
	if vl.List == nil {
		return nil
	}
	for _, x := range vl.List {
		// log.Printf(">>> compare: %s, %s : %v", x.Type, pt, x.Type.EqualTo(pt))
		if x.Type.EqualTo(pt) {
			if x.mode.Contains(mode) {
				return x
			}
		}
	}

	return nil
}

func (vl *VariableList) GetByTypeAndName(pt parser.Type, name string) *Variable {
	if vl.List == nil {
		return nil
	}
	for _, x := range vl.List {
		// log.Printf(">>> compare: %s, %s : %v", x.Type, pt, x.Type.EqualTo(pt))
		if x.Type.EqualTo(pt) {
			if x.name == name {
				return x
			}
		}
	}

	return nil

}

func (vl *VariableList) Get(name string) *Variable {
	for _, v := range vl.List {
		if v.name == name {
			return v
		}
	}

	return nil
}

func (vl *VariableList) Check(name string) bool {
	for _, v := range vl.List {
		if v.name == name {
			return true
		}
	}

	return false
}

func (vl *VariableList) Add(v *Variable) bool {
	if !v.IsAnonymous() {
		if exists := vl.Check(v.name); exists {
			return false
		}
	}

	vl.List = append(vl.List, v)
	return true
}

func (vl *VariableList) Insert(v *Variable, pos int) bool {
	if !v.IsAnonymous() {
		if exists := vl.Check(v.name); exists {
			return false
		}
	}
	list := append([]*Variable{}, vl.List[:pos]...)
	list = append(list, v)
	list = append(list, vl.List[pos:]...)
	vl.List = list
	return true
}

func (vl *VariableList) Names() []string {
	names := make([]string, len(vl.List))
	for i, v := range vl.List {
		names[i] = v.name
	}
	return names
}

func (vl *VariableList) Getter(mode Mode) VarGetter {
	return NewVarListGetter(vl, mode)
}
func (vl *VariableList) ResultGetter(ignores []int) VarGetter {
	return VarGetterWithIgnore(NewVarListGetter(vl, WRITE_MODE), ignores)
}
func (vl *VariableList) CreateGetter(mode Mode) VarGetter {
	return NewCreateVarGetter(vl, mode)
}

func (vl *VariableList) debug() {
	vl.ForEach(func(i int, t *Variable) {
		log.Printf("%d: %s, %s: %d", i, t.Name(), t.Type, t.mode)
	})

}

func (vl *VariableList) Print() {
	vl.ForEach(func(i int, t *Variable) {
		log.Printf("%d: %s, %s: %d", i, t.Name(), t.Type, t.mode)
	})
}

func NewAstIdent(name string) *ast.Ident {
	return &ast.Ident{Name: name}
}

func NewSelectorExpr(x ast.Expr, sel *ast.Ident) *ast.SelectorExpr {
	return &ast.SelectorExpr{X: x, Sel: sel}
}
