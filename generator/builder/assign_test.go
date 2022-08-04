package builder

import (
	"fmt"
	"log"

	"github.com/lawrsp/pigo/generator/parser"
)

var code = fmt.Sprintf(`
package pkgintest


type X int
type Y string

type A struct {
	PX *X  %[1]spc:"x"%[1]s
	PY *Y  %[1]spc:"y"%[1]s
}

type B struct {
	TheX X  %[1]spc:"x"%[1]s
	TheY Y  %[1]spc:"y"%[1]s
}

type C struct {
	Y   %[1]spc:"y"%[1]s
	T    int  %[1]spc:"x"%[1]s
}
type D struct {
	 B
}
`, "`", "`")

func ExampleStructAssign() {

	// fmt.Println(code)
	p := parser.NewParser()

	file := p.ParseFileContent("test", code)
	pkg := parser.NewPackage(p, "fake", "./fake.go", "", []*parser.File{file})
	file = pkg.Files[0]

	fileBuilder := NewFile(nil, file)

	tx := file.ReduceTypeSrc("X")
	// ty := file.ReduceTypeSrc("Y")
	ta := file.ReduceTypeSrc("A")
	tb := file.ReduceTypeSrc("B")
	tc := file.ReduceTypeSrc("C")

	var fnt parser.Type
	var bd *FuncBuilder
	var src *Variable
	var dst *Variable
	var assignBuilder *StructAssignBuilder

	//A => B
	fnt = fakeFunctionType(ta, tb, "test1")
	bd = NewFunction(fileBuilder, nil, fnt, nil)
	src = GetVariable(bd, ta, READ_MODE, Scope_Function)
	assignBuilder = NewStructAssign(bd, "pc", nil, nil)
	dst = NewVariable(tb).AutoName().WriteOnly()
	_ = assignBuilder.TryAssign(src, dst)
	bd.Add(assignBuilder)
	AddSuccessReturn(bd)
	fmt.Printf("%s\n", string(bd.Bytes()))

	//A => C
	log.Printf("test2")
	fnt = fakeFunctionType(ta, tc, "test2")
	bd = NewFunction(fileBuilder, nil, fnt, nil)
	src = GetVariable(bd, ta, READ_MODE, Scope_Function)
	assignBuilder = NewStructAssign(bd, "pc", nil, nil)
	dst = NewVariable(tc).AutoName().WriteOnly()
	_ = assignBuilder.TryAssign(src, dst)
	bd.Add(assignBuilder)
	AddSuccessReturn(bd)
	fmt.Printf("%s\n", string(bd.Bytes()))

	//A => x
	log.Printf("test3")
	fnt = fakeFunctionType(ta, tx, "test3")
	bd = NewFunction(fileBuilder, nil, fnt, nil)
	src = GetVariable(bd, ta, READ_MODE, Scope_Function)
	assignBuilder = NewStructAssign(bd, "pc", nil, nil)
	dst = NewVariable(tx).AutoName().WriteOnly()
	_ = assignBuilder.TryAssign(src, dst)
	bd.Add(assignBuilder)
	AddSuccessReturn(bd)
	fmt.Printf("%s\n", string(bd.Bytes()))

	//x => A
	fnt = fakeFunctionType(tx, ta, "test4")
	bd = NewFunction(fileBuilder, nil, fnt, nil)
	src = GetVariable(bd, tx, READ_MODE, Scope_Function)
	assignBuilder = NewStructAssign(bd, "pc", nil, nil)
	dst = NewVariable(ta).AutoName().WriteOnly()
	_ = assignBuilder.TryAssign(src, dst)
	bd.Add(assignBuilder)
	AddSuccessReturn(bd)
	fmt.Printf("%s\n", string(bd.Bytes()))

	//x => **A
	srcType := parser.TypeWithPointer(parser.TypeWithPointer(ta))
	dstType := tx
	fnt = fakeFunctionType(dstType, srcType, "test5")
	bd = NewFunction(fileBuilder, nil, fnt, nil)
	src = GetVariable(bd, dstType, READ_MODE, Scope_Function)
	assignBuilder = NewStructAssign(bd, "pc", nil, nil)
	dst = NewVariable(srcType).AutoName().WriteOnly()
	_ = assignBuilder.TryAssign(src, dst)
	bd.Add(assignBuilder)
	AddSuccessReturn(bd)
	fmt.Printf("%s\n", string(bd.Bytes()))

	//*x => **B
	srcType = parser.TypeWithPointer(parser.TypeWithPointer(tb))
	dstType = parser.TypeWithPointer(tx)
	fnt = fakeFunctionType(dstType, srcType, "test6")
	bd = NewFunction(fileBuilder, nil, fnt, nil)
	src = GetVariable(bd, dstType, READ_MODE, Scope_Function)
	assignBuilder = NewStructAssign(bd, "pc", nil, nil)
	dst = NewVariable(srcType).AutoName().WriteOnly()
	_ = assignBuilder.TryAssign(src, dst)
	bd.Add(assignBuilder)
	AddSuccessReturn(bd)
	fmt.Printf("%s\n", string(bd.Bytes()))

	// Output:
	//func test1(t A) (B, error) {
	//	b := B{}
	//	if t.PX != nil {
	//		b.TheX = *t.PX
	//	}
	//	if t.PY != nil {
	//		b.TheY = *t.PY
	//	}
	//	return b, nil
	//}
	//func test2(t A) (C, error) {
	//	c := C{}
	//	if t.PY != nil {
	//		c.Y = *t.PY
	//	}
	//	if t.PX != nil {
	//		x := *t.PX
	//		c.T = int(x)
	//	}
	//	return c, nil
	//}
	//func test3(t A) (X, error) {
	//	var x X
	//	if t.PX != nil {
	//		x = *t.PX
	//	}
	//	return x, nil
	//}
	//func test4(t X) (A, error) {
	//	a := A{}
	//	a.PX = &t
	//	return a, nil
	//}
	//func test5(t X) (**A, error) {
	//	a := &A{}
	//	a.PX = &t
	//	a1 := &a
	//	return a1, nil
	//}
	//func test6(t *X) (**B, error) {
	//	b := &B{}
	//	if t != nil {
	//		b.TheX = *t
	//	}
	//	b1 := &b
	//	return b1, nil
	//}
}

// A => D
// D => A
func ExampleStructAssign2() {

	// fmt.Println(code)
	p := parser.NewParser()

	file := p.ParseFileContent("test", code)
	pkg := parser.NewPackage(p, "fake", "./fake.go", "", []*parser.File{file})
	file = pkg.Files[0]

	fileBuilder := NewFile(nil, file)

	ta := file.ReduceTypeSrc("A")
	td := file.ReduceTypeSrc("D")

	srcType := ta
	dstType := td
	fnt := fakeFunctionType(dstType, srcType, "test21")
	bd := NewFunction(fileBuilder, nil, fnt, nil)
	src := GetVariable(bd, dstType, READ_MODE, Scope_Function)
	assignBuilder := NewStructAssign(bd, "pc", nil, nil)
	dst := NewVariable(srcType).AutoName().WriteOnly()
	assignBuilder.TryAssign(src, dst)
	bd.Add(assignBuilder)
	AddSuccessReturn(bd)
	fmt.Printf("%s\n", string(bd.Bytes()))

	srcType = td
	dstType = ta
	fnt = fakeFunctionType(dstType, srcType, "test22")
	bd = NewFunction(fileBuilder, nil, fnt, nil)
	src = GetVariable(bd, dstType, READ_MODE, Scope_Function)
	assignBuilder = NewStructAssign(bd, "pc", nil, nil)
	dst = NewVariable(srcType).AutoName().WriteOnly()
	assignBuilder.TryAssign(src, dst)
	bd.Add(assignBuilder)
	AddSuccessReturn(bd)
	fmt.Printf("%s\n", string(bd.Bytes()))

	//Output:
	//func test21(t D) (A, error) {
	//	a := A{}
	//	a.PX = &t.TheX
	//	a.PY = &t.TheY
	//	return a, nil
	//}
	//func test22(t A) (D, error) {
	//	d := D{}
	//	if t.PX != nil {
	//		d.TheX = *t.PX
	//	}
	//	if t.PY != nil {
	//		d.TheY = *t.PY
	//	}
	//	return d, nil
	//}
}

func ExampleStructAssign3() {

	// fmt.Println(code)
	p := parser.NewParser()

	file := p.ParseFileContent("test", code)
	pkg := parser.NewPackage(p, "fake", "./fake.go", "", []*parser.File{file})
	file = pkg.Files[0]

	fileBuilder := NewFile(nil, file)

	ta := file.ReduceTypeSrc("A")
	td := file.ReduceTypeSrc("A")

	srcType := ta
	dstType := td
	fnt := fakeFunctionType(dstType, srcType, "test31")
	bd := NewFunction(fileBuilder, nil, fnt, nil)
	src := GetVariable(bd, dstType, READ_MODE, Scope_Function)
	assignBuilder := NewStructAssign(bd, "pc", nil, nil)
	dst := NewVariable(srcType).AutoName().WriteOnly()
	assignBuilder.TryAssign(src, dst)
	bd.Add(assignBuilder)
	AddSuccessReturn(bd)
	fmt.Printf("%s\n", string(bd.Bytes()))

	//Output:
	//func test31(t A) (A, error) {
	//	a := A{}
	//	a.PX = t.PX
	//	a.PY = t.PY
	//	return a, nil
	//}
}
