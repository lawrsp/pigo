package builder

import (
	"go/ast"

	"github.com/lawrsp/pigo/pkg/parser"
)

type VarGetter interface {
	Get(int, parser.Type) *Variable
	Getted() *VariableList
}

type VarListGetter struct {
	vl     *VariableList
	chosed *VariableList
	mode   Mode
}

func NewVarListGetter(vl *VariableList, mode Mode) VarGetter {
	return &VarListGetter{vl, NewVariableList(), mode}
}
func (getter *VarListGetter) Get(i int, t parser.Type) *Variable {
	v := getter.vl.GetByType(t, getter.mode)
	if v != nil {
		getter.chosed.Add(v)
	}
	return v
}
func (getter *VarListGetter) Getted() *VariableList {
	return getter.chosed
}

type CreateVarGetter struct {
	exists  *VariableList
	NewVars *VariableList
	mode    Mode
}

func NewCreateVarGetter(exists *VariableList, mode Mode) VarGetter {
	return &CreateVarGetter{exists, NewVariableList(), mode}
}
func (getter *CreateVarGetter) Getted() *VariableList {
	return getter.NewVars
}
func (getter *CreateVarGetter) Get(i int, t parser.Type) *Variable {
	v := NewVariable(t).AutoName().WithMode(getter.mode)
	// log.Printf("new write only variable: %s", v.Name())
	for getter.exists.Check(v.Name()) == true || getter.NewVars.Add(v) != true {
		v.IncreaseName()
	}
	return v
}

type ignoresVarGetter struct {
	VarGetter
	ignores []int
}

func VarGetterWithIgnore(getter VarGetter, ignores []int) VarGetter {
	if ignores == nil {
		ignores = []int{}
	}
	return &ignoresVarGetter{getter, ignores}
}
func (g *ignoresVarGetter) Get(i int, t parser.Type) *Variable {
	for _, ignore := range g.ignores {
		if ignore == i {
			return NewVariable(t).WithExpr(ast.NewIdent("_")).WriteOnly()
		}
	}
	return g.VarGetter.Get(i, t)
}
