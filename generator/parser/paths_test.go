package parser

import (
	"fmt"
	"testing"
)

func TestTypeToType(t *testing.T) {
	p := NewParser()
	pkg := p.ParsePackageDir("./")
	code := `
package parser

import "github.com/lawrsp/pigo/generatorutils"
import "github.com/lawrsp/pigo/generatorbuilder"
import "strings"
import "go/ast"
import "github.com/golang/protobuf/ptypes/timestamp"
import "time"
import "github.com/golang/protobuf/ptypes"


type X int

type A struct {
	PX *X
}

type B struct {
	TheX X
}

type C struct {
	TheX X
	T    int
  Map  map[string]interface{}
}

type D struct {
	T  X
}

type E map[string]interface{}

func A2B(src *A) (*B, error) {
	b := &B{}
	if A.PX != nil {
		 b.TheX = *A.PX
	}
	return b, nil
}

func B2A(src *B) (*A, error) {
	A = &A{}
	t := &X{}
	*t = B.TheX
	A.PX = t
	return A, nil
}
`
	file := p.ParseFileContent("_test", code)
	p.InsertFileToPackage(pkg, file, 1)

	knowns := []*TPath{}
	{
		x := file.ReduceTypeSrc("*A")
		y := file.ReduceTypeSrc("*B")
		convert := file.ReduceTypeSrc("A2B")
		reverse := file.ReduceTypeSrc("B2A")

		knowns = append(knowns, NewTPath(x, y).WithFunction(convert))
		knowns = append(knowns, NewTPath(y, x).WithFunction(reverse))
	}

	//A >> A = 1
	aStr := "A"
	bStr := aStr
	fmt.Printf("paths:  %s  >>>>  %s\n", aStr, bStr)
	a := file.ReduceTypeSrc(aStr)
	b := file.ReduceTypeSrc(bStr)
	fps, ok := TypeToType(a, b, knowns)
	expect(t.Errorf, len(fps), 1)
	expect(t.Errorf, ok, true)
	expect(t.Errorf, fps[0].D, D_Self)
	fmt.Printf("paths: %s\n", fps)

	//x >> B.x (self)
	aStr = "X"
	bStr = "B.TheX"
	fmt.Printf("paths:  %s >>>>  %s\n", aStr, bStr)
	a = file.ReduceTypeSrc(aStr)
	expect(t.Errorf, a.String(), "parser.X")
	b = file.ReduceTypeSrc(bStr)
	expect(t.Errorf, b.String(), "parser.X")
	fps, ok = TypeToType(a, b, knowns)
	expect(t.Errorf, len(fps), 1)
	expect(t.Errorf, ok, true)
	fmt.Printf("paths: %s\n", fps)

	//B{x} >> x (self,  -struct)
	aStr = "B"
	bStr = "X"
	fmt.Printf("paths:  %s >>>>  %s\n", aStr, bStr)
	a = file.ReduceTypeSrc(aStr)
	b = file.ReduceTypeSrc(bStr)
	expect(t.Errorf, a.String(), "parser.B")
	expect(t.Errorf, b.String(), "parser.X")
	fps, ok = TypeToType(a, b, knowns)
	// expect(t.Errorf, len(fps), 2)
	if len(fps) != 2 {
		t.Errorf("expected: 2, get %d:\n%s", len(fps), fps)
	}
	expect(t.Errorf, ok, true)
	fmt.Printf("paths: %s\n", fps)

	aStr = "D"
	bStr = "int"
	fmt.Printf("paths:  %s >>>>  %s\n", aStr, bStr)
	a = file.ReduceTypeSrc(aStr)
	b = file.ReduceTypeSrc(bStr)
	expect(t.Errorf, a.String(), "parser.D")
	expect(t.Errorf, b.String(), "int")
	fps, ok = TypeToType(a, b, knowns)
	// expect(t.Errorf, len(fps), 2)
	if len(fps) != 3 {
		t.Errorf("expected: 3, get %d:\n%s", len(fps), fps)
	}
	expect(t.Errorf, ok, true)
	fmt.Printf("paths: %s\n", fps)

	//*B{x} >> x (self, -pointer, -struct)
	aStr = "*B"
	bStr = "X"
	fmt.Printf("paths:  %s >>>>  %s\n", aStr, bStr)
	a = file.ReduceTypeSrc(aStr)
	b = file.ReduceTypeSrc(bStr)
	expect(t.Errorf, a.String(), "*parser.B")
	expect(t.Errorf, b.String(), "parser.X")
	fps, ok = TypeToType(a, b, knowns)
	// expect(t.Errorf, len(fps), 2)
	if len(fps) != 3 {
		t.Errorf("expected: 3, get %d", len(fps))
	}
	expect(t.Errorf, ok, true)
	fmt.Printf("paths: %s\n", fps)

	//*A => *B (self, convert)
	aStr = "*A"
	bStr = "*B"
	fmt.Printf("paths:  %s >>>>  %s\n", aStr, bStr)
	a = file.ReduceTypeSrc(aStr)
	b = file.ReduceTypeSrc(bStr)
	expect(t.Errorf, a.String(), "*parser.A")
	expect(t.Errorf, b.String(), "*parser.B")
	fps, ok = TypeToType(a, b, knowns)
	expect(t.Errorf, len(fps), 2)
	expect(t.Errorf, ok, true)
	fmt.Printf("paths: %s\n", fps)

	//*A => B (self, convert,-pointer)
	aStr = "*A"
	bStr = "B"
	fmt.Printf("paths:  %s >>>>  %s\n", aStr, bStr)
	a = file.ReduceTypeSrc(aStr)
	b = file.ReduceTypeSrc(bStr)
	expect(t.Errorf, a.String(), "*parser.A")
	expect(t.Errorf, b.String(), "parser.B")
	fps, ok = TypeToType(a, b, knowns)
	expect(t.Errorf, len(fps), 3)
	expect(t.Errorf, ok, true)
	fmt.Printf("paths: %s\n", fps)

	//[]*A >> []*B (self, -slice, convert, +slice)
	aStr = "[]*A"
	bStr = "[]*B"
	fmt.Printf("paths:  %s >>>>  %s\n", aStr, bStr)
	a = file.ReduceTypeSrc(aStr)
	b = file.ReduceTypeSrc(bStr)
	expect(t.Errorf, a.String(), "[]*parser.A")
	expect(t.Errorf, b.String(), "[]*parser.B")
	fps, ok = TypeToType(a, b, knowns)
	expect(t.Errorf, len(fps), 4)
	expect(t.Errorf, ok, true)
	fmt.Printf("paths: %s\n", fps)

	//[]A >> []B (self, -slice, +pointer, convert, -pointer, +slice)
	aStr = "[]A"
	bStr = "[]B"
	fmt.Printf("paths:  %s >>>>  %s\n", aStr, bStr)
	a = file.ReduceTypeSrc(aStr)
	b = file.ReduceTypeSrc(bStr)
	expect(t.Errorf, a.String(), "[]parser.A")
	expect(t.Errorf, b.String(), "[]parser.B")
	fps, ok = TypeToType(a, b, knowns)
	expect(t.Errorf, len(fps), 6)
	expect(t.Errorf, ok, true)
	fmt.Printf("paths: %s\n", fps)

	// **A >> *B [*A => *B]
	aStr = "**A"
	bStr = "*B"
	fmt.Printf("paths:  %s >>>>  %s\n", aStr, bStr)
	a = file.ReduceTypeSrc(aStr)
	b = file.ReduceTypeSrc(bStr)
	expect(t.Errorf, a.String(), "**parser.A")
	expect(t.Errorf, b.String(), "*parser.B")
	fps, ok = TypeToType(a, b, knowns)
	expect(t.Errorf, ok, true)
	fmt.Printf("paths: %s\n", fps)

	// x >> B (false)
	aStr = "X"
	bStr = "B"
	fmt.Printf("paths:  %s >>>>  %s\n", aStr, bStr)
	a = file.ReduceTypeSrc(aStr)
	b = file.ReduceTypeSrc(bStr)
	expect(t.Errorf, a.String(), "parser.X")
	expect(t.Errorf, b.String(), "parser.B")
	fps, ok = TypeToType(a, b, knowns)
	expect(t.Errorf, ok, false)

	// fmt.Printf("paths:  %s >>>>  %s\n", bStr, aStr)
	// paths = TypeToType(b, a, knowns)
	// expect(t.Errorf, len(paths), 2)

	// E >> C (true)
	aStr = "E"
	bStr = "C.Map"
	fmt.Printf("paths:  %s >>>>  %s\n", aStr, bStr)
	a = file.ReduceTypeSrc(aStr)
	b = file.ReduceTypeSrc(bStr)
	expect(t.Errorf, a.String(), "parser.E")
	expect(t.Errorf, b.String(), "map[string]interface{}")
	fps, ok = TypeToType(a, b, knowns)
	if !ok {
		t.Errorf("%s != %s", a.String(), b.String())
	}

}
