package builder

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/lawrsp/pigo/generator/parser"
)

func expect(a interface{}, b interface{}, logger func(string, ...interface{})) {
	if !reflect.DeepEqual(b, a) {
		logger("Expected %#v (type %v) - Got %#v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	fmt.Println("Test started")

	os.Exit(m.Run())
}

func fakeFunctionType(source, target parser.Type, name string) parser.Type {
	fnt := &parser.FuncType{}
	fnt.Params = []*parser.Field{parser.NewField(source, "t", "")}
	fnt.Results = []*parser.Field{
		parser.NewField(target, "", ""),
		parser.NewField(parser.ErrorType(), "", ""),
	}
	funcType := parser.TypeWithName(fnt, name)
	return funcType
}

func fakeFunctionTypeNoError(source, target parser.Type, name string) parser.Type {
	fnt := &parser.FuncType{}
	fnt.Params = []*parser.Field{parser.NewField(source, "t", "")}
	fnt.Results = []*parser.Field{
		parser.NewField(target, "", ""),
	}
	funcType := parser.TypeWithName(fnt, name)
	return funcType
}

//T1 => T1
func TestFollowPath1(t *testing.T) {
	code := `
package tt

`
	p := parser.NewParser()

	file := p.ParseFileContent("test", code)
	pkg := parser.NewPackage(p, "fake", "./fake.go", "", []*parser.File{file})
	file = pkg.Files[0]

	fileBuilder := NewFile(nil, file)

	t1 := parser.TypeWithName(nil, "T1")
	tpaths := []*parser.TPath{
		parser.NewTPath(t1, t1).WithDirection(parser.D_Self),
	}

	fnt := fakeFunctionType(t1, t1, "test0")
	bd := NewFunction(fileBuilder, nil, fnt, nil)
	src := GetVariable(bd, t1, READ_MODE, Scope_Function)
	expect(src != nil, true, t.Errorf)
	log.Printf("src is %s(%d)", src.Type, src.mode)

	tbd := NewTPathBuilder(bd, nil, false)
	tbd.Follow(src, nil, tpaths)
	bd.Add(tbd)
	AddSuccessReturn(bd)
	fmt.Println(string(bd.Bytes()))
}

//T1 => *T1 => *T2 => T2
func TestFollowPath2(t *testing.T) {
	code := `
package tt

`
	p := parser.NewParser()

	file := p.ParseFileContent("test", code)
	pkg := parser.NewPackage(p, "fake", "./fake.go", "", []*parser.File{file})
	file = pkg.Files[0]

	fileBuilder := NewFile(nil, file)

	t1 := parser.TypeWithName(nil, "T1")
	pt1 := parser.TypeWithPointer(t1)

	t2 := parser.TypeWithName(nil, "T2")
	pt2 := parser.TypeWithPointer(t2)
	tpaths := []*parser.TPath{
		parser.NewTPath(t1, t1).WithDirection(parser.D_Self),
		parser.NewTPath(t1, pt1).WithDirection(parser.D_AddPointer),
		parser.NewTPath(pt1, pt2).WithFunction(fakeFunctionType(pt1, pt2, "pt1_pt2")),
		parser.NewTPath(pt2, t2).WithDirection(parser.D_SkipPointer),
	}

	fnt := fakeFunctionType(t1, t2, "test1")
	bd := NewFunction(fileBuilder, nil, fnt, nil)
	src := GetVariable(bd, t1, READ_MODE, Scope_Function)
	expect(src != nil, true, t.Errorf)

	tbd := NewTPathBuilder(bd, nil, false)
	tbd.Follow(src, nil, tpaths)
	bd.Add(tbd)
	AddSuccessReturn(bd)
	fmt.Println(string(bd.Bytes()))
}

//[]T1 => T1 => *T1 => *T2 => T2 => []T2 => *[]T2
func TestFollowPath3(t *testing.T) {
	code := `
package tt

`
	p := parser.NewParser()

	file := p.ParseFileContent("test", code)
	pkg := parser.NewPackage(p, "fake", "./fake.go", "", []*parser.File{file})
	file = pkg.Files[0]

	fileBuilder := NewFile(nil, file)

	t1 := parser.TypeWithName(nil, "T1")
	st1 := parser.TypeWithSlice(t1)
	pt1 := parser.TypeWithPointer(t1)

	t2 := parser.TypeWithName(nil, "T2")
	st2 := parser.TypeWithSlice(t2)
	pt2 := parser.TypeWithPointer(t2)
	pst2 := parser.TypeWithPointer(st2)
	tpaths := []*parser.TPath{
		parser.NewTPath(st1, st1).WithDirection(parser.D_Self),
		parser.NewTPath(st1, t1).WithDirection(parser.D_SkipBracket),
		parser.NewTPath(t1, pt1).WithDirection(parser.D_AddPointer),
		parser.NewTPath(pt1, pt2).WithFunction(fakeFunctionType(pt1, pt2, "pt1_pt2")),
		parser.NewTPath(pt2, t2).WithDirection(parser.D_SkipPointer),
		parser.NewTPath(t2, st2).WithDirection(parser.D_AddBracket),
		parser.NewTPath(t2, pst2).WithDirection(parser.D_AddPointer),
	}

	fnt := fakeFunctionType(st1, pst2, "test2")
	bd := NewFunction(fileBuilder, nil, fnt, nil)
	src := GetVariable(bd, st1, READ_MODE, Scope_Function)
	expect(src != nil, true, t.Errorf)

	tbd := NewTPathBuilder(bd, nil, false)
	tbd.Follow(src, nil, tpaths)
	bd.Add(tbd)
	AddSuccessReturn(bd)
	fmt.Println(string(bd.Bytes()))
}

//[]T1 => T1 => *T1 => *T2 =>  []*T2
func TestFollowPath5(t *testing.T) {
	code := `
package tt

`
	p := parser.NewParser()

	file := p.ParseFileContent("test", code)
	pkg := parser.NewPackage(p, "fake", "./fake.go", "", []*parser.File{file})
	file = pkg.Files[0]

	fileBuilder := NewFile(nil, file)

	t1 := parser.TypeWithName(nil, "T1")
	st1 := parser.TypeWithSlice(t1)
	pt1 := parser.TypeWithPointer(t1)

	t2 := parser.TypeWithName(nil, "T2")
	pt2 := parser.TypeWithPointer(t2)
	spt2 := parser.TypeWithSlice(pt2)
	tpaths := []*parser.TPath{
		parser.NewTPath(st1, st1).WithDirection(parser.D_Self),
		parser.NewTPath(st1, t1).WithDirection(parser.D_SkipBracket),
		parser.NewTPath(t1, pt1).WithDirection(parser.D_AddPointer),
		parser.NewTPath(pt1, pt2).WithFunction(fakeFunctionType(pt1, pt2, "pt1_pt2")),
		parser.NewTPath(pt2, spt2).WithDirection(parser.D_AddBracket),
	}

	fnt := fakeFunctionType(st1, spt2, "test5")
	bd := NewFunction(fileBuilder, nil, fnt, nil)
	src := GetVariable(bd, st1, READ_MODE, Scope_Function)
	dst := NewVariable(spt2).WithName("dst").WriteOnly()
	expect(src != nil, true, t.Errorf)
	tbd := NewTPathBuilder(bd, nil, false)
	tbd.Follow(src, dst, tpaths)
	bd.Add(tbd)
	AddSuccessReturn(bd)
	fmt.Println(string(bd.Bytes()))
}

//[][]T1 => []T1 => T1 => *T1 => *T2 => T2 => []T2 => *[]T2
func TestFollowPath6(t *testing.T) {
	code := `
package tt

`
	p := parser.NewParser()

	file := p.ParseFileContent("test", code)
	pkg := parser.NewPackage(p, "fake", "./fake.go", "", []*parser.File{file})
	file = pkg.Files[0]

	fileBuilder := NewFile(nil, file)

	t1 := parser.TypeWithName(nil, "T1")
	st1 := parser.TypeWithSlice(t1)
	pt1 := parser.TypeWithPointer(t1)
	sst1 := parser.TypeWithSlice(st1)

	t2 := parser.TypeWithName(nil, "T2")
	st2 := parser.TypeWithSlice(t2)
	pt2 := parser.TypeWithPointer(t2)
	pst2 := parser.TypeWithPointer(st2)
	tpaths := []*parser.TPath{
		parser.NewTPath(sst1, sst1).WithDirection(parser.D_Self),
		parser.NewTPath(sst1, st1).WithDirection(parser.D_SkipBracket),
		parser.NewTPath(st1, t1).WithDirection(parser.D_SkipBracket),
		parser.NewTPath(t1, pt1).WithDirection(parser.D_AddPointer),
		parser.NewTPath(pt1, pt2).WithFunction(fakeFunctionType(pt1, pt2, "pt1_pt2")),
		parser.NewTPath(pt2, t2).WithDirection(parser.D_SkipPointer),
		parser.NewTPath(t2, st2).WithDirection(parser.D_AddBracket),
		parser.NewTPath(t2, pst2).WithDirection(parser.D_AddPointer),
	}

	fnt := fakeFunctionType(sst1, pst2, "test4")
	bd := NewFunction(fileBuilder, nil, fnt, nil)
	src := GetVariable(bd, sst1, READ_MODE, Scope_Function)
	expect(src != nil, true, t.Errorf)
	tbd := NewTPathBuilder(bd, nil, false)
	tbd.Follow(src, nil, tpaths)
	bd.Add(tbd)
	AddSuccessReturn(bd)
	fmt.Println(string(bd.Bytes()))
}

// T1 => *T1 => *T2 => T2 => []T2 => *[]T2
func TestFollowPath7(t *testing.T) {
	code := `
package tt

`
	p := parser.NewParser()

	file := p.ParseFileContent("test", code)
	pkg := parser.NewPackage(p, "fake", "./fake.go", "", []*parser.File{file})
	file = pkg.Files[0]

	fileBuilder := NewFile(nil, file)

	t1 := parser.TypeWithName(nil, "T1")
	pt1 := parser.TypeWithPointer(t1)

	t2 := parser.TypeWithName(nil, "T2")
	st2 := parser.TypeWithSlice(t2)
	pt2 := parser.TypeWithPointer(t2)
	pst2 := parser.TypeWithPointer(st2)
	tpaths := []*parser.TPath{
		parser.NewTPath(t1, t1).WithDirection(parser.D_Self),
		parser.NewTPath(t1, pt1).WithDirection(parser.D_AddPointer),
		parser.NewTPath(pt1, pt2).WithFunction(fakeFunctionType(pt1, pt2, "pt1_pt2")),
		parser.NewTPath(pt2, t2).WithDirection(parser.D_SkipPointer),
		parser.NewTPath(t2, st2).WithDirection(parser.D_AddBracket),
		parser.NewTPath(t2, pst2).WithDirection(parser.D_AddPointer),
	}

	fnt := fakeFunctionType(t1, pst2, "test5")
	bd := NewFunction(fileBuilder, nil, fnt, nil)
	src := GetVariable(bd, t1, READ_MODE, Scope_Function)
	expect(src != nil, true, t.Errorf)
	tbd := NewTPathBuilder(bd, nil, false)
	tbd.Follow(src, nil, tpaths)
	bd.Add(tbd)
	AddSuccessReturn(bd)
	fmt.Println(string(bd.Bytes()))
}

//[]*T1 => *T1 =>(call) => T2 => *T2 => []*T2
func ExampleFollowPath8() {
	code := `
package tt

`
	p := parser.NewParser()

	file := p.ParseFileContent("test", code)
	pkg := parser.NewPackage(p, "fake", "./fake.go", "", []*parser.File{file})
	file = pkg.Files[0]

	fileBuilder := NewFile(nil, file)

	t1 := parser.TypeWithName(nil, "T1")
	pt1 := parser.TypeWithPointer(t1)
	spt1 := parser.TypeWithSlice(pt1)

	t2 := parser.TypeWithName(nil, "T2")
	pt2 := parser.TypeWithPointer(t2)
	spt2 := parser.TypeWithSlice(pt2)

	tpaths := []*parser.TPath{
		parser.NewTPath(spt1, spt1).WithDirection(parser.D_Self),
		parser.NewTPath(spt1, pt1).WithDirection(parser.D_SkipBracket),
		parser.NewTPath(pt1, t2).WithFunction(fakeFunctionType(pt1, t2, "pt1_t2")),
		parser.NewTPath(t2, pt2).WithDirection(parser.D_AddPointer),
		parser.NewTPath(pt2, spt2).WithDirection(parser.D_AddBracket),
	}

	srcType := spt1
	dstType := spt2

	fnt := fakeFunctionType(srcType, dstType, "test8")
	bd := NewFunction(fileBuilder, nil, fnt, nil)
	src := GetVariable(bd, srcType, READ_MODE, Scope_Function)

	tbd := NewTPathBuilder(bd, nil, false)
	tbd.Follow(src, nil, tpaths)
	bd.Add(tbd)
	AddSuccessReturn(bd)

	fmt.Println(string(bd.Bytes()))
	// Output:
	// func test8(t []*T1) ([]*T2, error) {
	//	var t2List []*T2
	//	for _, t1 := range t {
	//		if t1 != nil {
	//			t2, err := pt1_t2(t1)
	//			if err != nil {
	//				return nil, err
	//			}
	//			t3 := &t2
	//			t2List = append(t2List, t3)
	//		}
	//	}
	//	return t2List, nil
	//}
}

//T1 => T1
func ExampleFollowPath9() {
	code := `
package tt

`
	p := parser.NewParser()

	file := p.ParseFileContent("test", code)
	pkg := parser.NewPackage(p, "fake", "./fake.go", "", []*parser.File{file})
	file = pkg.Files[0]

	fileBuilder := NewFile(nil, file)

	t1 := parser.TypeWithName(nil, "T1")
	tpaths := []*parser.TPath{
		parser.NewTPath(t1, t1).WithDirection(parser.D_Self),
	}

	fnt := fakeFunctionType(t1, t1, "test9")
	bd := NewFunction(fileBuilder, nil, fnt, nil)
	src := GetVariable(bd, t1, READ_MODE, Scope_Function)

	mapType := parser.MapType(parser.NewBasicType("string"), parser.NewBasicType("interface"))
	updateV := NewVariable(mapType).WithName("updated").ReadOnly()
	updateV = AddVariableDecl(bd, updateV)

	tbd := NewTPathBuilder(bd, nil, true)
	tbd.Follow(src, nil, tpaths)
	bd.Add(tbd)
	AddSuccessReturn(bd)
	fmt.Println(string(bd.Bytes()))
	// Output:
	//func test9(t T1) (T1, error) {
	//	var updated map[string]interface{}
	//	return t, nil
	//}
}

//[]*T1 => *T1 =>(call) => T2 => []T2
func ExampleFollowPath10() {
	code := `
package tt

`
	p := parser.NewParser()

	file := p.ParseFileContent("test", code)
	pkg := parser.NewPackage(p, "fake", "./fake.go", "", []*parser.File{file})
	file = pkg.Files[0]

	fileBuilder := NewFile(nil, file)

	t1 := parser.TypeWithName(nil, "T1")
	pt1 := parser.TypeWithPointer(t1)
	spt1 := parser.TypeWithSlice(pt1)

	t2 := parser.TypeWithName(nil, "T2")
	st2 := parser.TypeWithSlice(t2)

	tpaths := []*parser.TPath{
		parser.NewTPath(spt1, spt1).WithDirection(parser.D_Self),
		parser.NewTPath(spt1, pt1).WithDirection(parser.D_SkipBracket),
		parser.NewTPath(pt1, t2).WithFunction(fakeFunctionTypeNoError(pt1, t2, "pt1_t2_noerror")),
		parser.NewTPath(t2, st2).WithDirection(parser.D_AddBracket),
	}

	srcType := spt1
	dstType := st2

	fnt := fakeFunctionType(srcType, dstType, "test10")
	bd := NewFunction(fileBuilder, nil, fnt, nil)
	src := GetVariable(bd, srcType, READ_MODE, Scope_Function)

	tbd := NewTPathBuilder(bd, nil, false)
	tbd.Follow(src, nil, tpaths)
	bd.Add(tbd)
	AddSuccessReturn(bd)

	fmt.Println(string(bd.Bytes()))
	// Output:
	// func test10(t []*T1) ([]T2, error) {
	//	var t2List []T2
	//	for _, t1 := range t {
	//		if t1 != nil {
	//			t2 := pt1_t2_noerror(t1)
	//			t2List = append(t2List, t2)
	//		}
	//	}
	//	return t2List, nil
	//}
}
