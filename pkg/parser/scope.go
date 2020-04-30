package parser

import "go/ast"

type Scope struct {
	Objects map[string]Type
}

func NewScope() *Scope {
	return &Scope{
		Objects: map[string]Type{},
	}
}

func (s *Scope) DeclType(name string, t Type) {
	s.Objects[name] = t
}

func (s *Scope) Lookup(name string) Type {
	return s.Objects[name]
}

var Universe = NewScope()

var UniverseScope = ast.NewScope(nil)

func declObj(kind ast.ObjKind, name string) {
	// UniverseScope.Insert(ast.NewObj(kind, name))
	UniverseScope.Objects[name] = ast.NewObj(kind, name)
	Universe.DeclType(name, &BasicType{name: name})
}

func init() {
	declObj(ast.Typ, "bool")

	declObj(ast.Typ, "complex64")
	declObj(ast.Typ, "complex128")

	declObj(ast.Typ, "int")
	declObj(ast.Typ, "int8")
	declObj(ast.Typ, "int16")
	declObj(ast.Typ, "int32")
	declObj(ast.Typ, "int64")

	declObj(ast.Typ, "uint")
	declObj(ast.Typ, "uintptr")
	declObj(ast.Typ, "uint8")
	declObj(ast.Typ, "uint16")
	declObj(ast.Typ, "uint32")
	declObj(ast.Typ, "uint64")

	declObj(ast.Typ, "float")
	declObj(ast.Typ, "float32")
	declObj(ast.Typ, "float64")

	declObj(ast.Typ, "string")
	declObj(ast.Typ, "error")
	declObj(ast.Typ, "interface")

	// predeclared constants
	// TODO(gri) provide constant value
	declObj(ast.Con, "false")
	declObj(ast.Con, "true")
	declObj(ast.Con, "iota")
	declObj(ast.Con, "nil")

	// predeclared functions
	// TODO(gri) provide "type"
	declObj(ast.Fun, "append")
	declObj(ast.Fun, "cap")
	declObj(ast.Fun, "close")
	declObj(ast.Fun, "complex")
	declObj(ast.Fun, "copy")
	declObj(ast.Fun, "delete")
	declObj(ast.Fun, "imag")
	declObj(ast.Fun, "len")
	declObj(ast.Fun, "make")
	declObj(ast.Fun, "new")
	declObj(ast.Fun, "panic")
	declObj(ast.Fun, "panicln")
	declObj(ast.Fun, "print")
	declObj(ast.Fun, "println")
	declObj(ast.Fun, "real")
	declObj(ast.Fun, "recover")

	// byte is an alias for uint8, so cheat
	// by storing the same object for both name
	// entries
	UniverseScope.Objects["byte"] = UniverseScope.Objects["uint8"]
	Universe.Objects["byte"] = Universe.Objects["uint8"]

	// The same applies to rune.
	UniverseScope.Objects["rune"] = UniverseScope.Objects["uint32"]
	Universe.Objects["rune"] = Universe.Objects["uint32"]
}
