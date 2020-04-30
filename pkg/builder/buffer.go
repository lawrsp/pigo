package builder

import (
	"bytes"
	"fmt"
	"go/ast"

	"github.com/lawrsp/pigo/pkg/parser"
)

type Printer interface {
	Printf(format string, args ...interface{})
}

// 	Package() *parser.Package
// 	File() *parser.File
// 	addVariable(*Variable) bool
// }

// type Builder interface {
// 	builderInternal
// 	Outer() Builder
// 	Variables() *VariableList
// 	Block() *BlockBuilder
// 	Add(Builder)
// 	Bytes() []byte
// 	String() string

type BufferBuilder interface {
	Builder
}

type FuncBufferBuilder struct {
	*baseBuilder
	buf  bytes.Buffer
	name string
}

func NewFuncBuffer(outer Builder, name string) *FuncBufferBuilder {
	base := newbaseBuilder(outer, outer.Package(), outer.File())
	fn := &FuncBufferBuilder{baseBuilder: base}
	fn.name = name
	return fn
}
func (b *FuncBufferBuilder) String() string {
	return fmt.Sprintf("buffer.func")
}
func (b *FuncBufferBuilder) Block() *BlockBuilder {
	return nil
}
func (b *FuncBufferBuilder) Bytes() []byte {
	return b.buf.Bytes()
}

func (b *FuncBufferBuilder) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&b.buf, format, args...)
}

func (b *FuncBufferBuilder) Decl() *ast.FuncDecl {
	p := parser.NewParser()
	file := p.ParseFileContent("_buffer", fmt.Sprintf("package buffer\n %s", b.Bytes()))
	var decl *ast.FuncDecl
	parser.WalkFile(file, parser.NewFuncDeclWalker(func(fd *ast.FuncDecl) bool {
		if fd.Name.Name == b.name {
			decl = fd
			return false
		}
		return true
	}))

	return decl
}
