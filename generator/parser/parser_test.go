package parser

import (
	"fmt"
	"go/ast"
	"os"
	"reflect"
	"runtime"
	"testing"

	"github.com/lawrsp/pigo/generator/printutil"
)

func expect(logger func(string, ...interface{}), a interface{}, b interface{}) {
	_, _, line, _ := runtime.Caller(1)
	if !reflect.DeepEqual(b, a) {
		logger("%d: Expected %#v (type %v) - Got %#v (type %v)", line, b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

// func expect(t *testing.T, a interface{}, b interface{}) {
// 	if !reflect.DeepEqual(b, a) {
// 		_, file, line, _ := runtime.Caller(1)
// 		t.Errorf("%s:%d  Expected %#v (type %v) - Got %#v (type %v)", file, line, b, reflect.TypeOf(b), a, reflect.TypeOf(a))
// 	}
// }

func assert(t *testing.T, ok bool, format string, args ...interface{}) {
	if !ok {
		_, file, line, _ := runtime.Caller(1)
		extends := fmt.Sprintf(format, args...)
		t.Errorf("%s:%d Assert Fail: %v", file, line, extends)
	}
}

func refute(t *testing.T, a interface{}, b interface{}) {
	if a == b {

		t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	fmt.Println("Test started")

	os.Exit(m.Run())
}

/*
func TestType(t *testing.T) {

	fs := token.NewFileSet()
	src := "****abc.Node"
	expr, err := ParseExpr(src)
	expect(t, err, nil)
	ast.Print(fs, expr)

	typ := ParseType(expr)

	expect(t, typ.String(), src)
}*/

func TestExpr(t *testing.T) {

	var expr ast.Expr
	var err error
	var src string
	//== nil
	b, _ := expr.(*ast.Ident)
	expect(t.Errorf, b == nil, true)

	//a name
	src = "uint64"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	printutil.PrintNodef(expr, src)

	src = "error"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	printutil.PrintNodef(expr, src)

	src = "interface{}"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	printutil.PrintNodef(expr, src)

	src = "Aname"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	printutil.PrintNodef(expr, src)

	//composed types:
	src = "map[ta]tb"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	printutil.PrintNodef(expr, src)

	src = "*Pointer"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	printutil.PrintNodef(expr, src)

	src = "[]slice"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	printutil.PrintNodef(expr, src)

	src = "[10]Array"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	printutil.PrintNodef(expr, src)

	src = "A(st{10})"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	printutil.PrintNodef(expr, src)

	src = "B(args)()"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	printutil.PrintNodef(expr, src)

	src = "C(args)[1]"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	printutil.PrintNodef(expr, src)

	src = "D{name:1, hell: 2}"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	printutil.PrintNodef(expr, src)

	src = "E(impt.Arg)"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	printutil.PrintNodef(expr, src)

	src = "A.B[0].C(impt.Arg).D"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	printutil.PrintNodef(expr, src)

	src = "10"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	printutil.PrintNodef(expr, src)

	src = "false"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	printutil.PrintNodef(expr, src)

}

type X interface {
	Go(int) string
}

func TestParsePackage(t *testing.T) {
	p := NewParser()
	pkg := p.ParsePackageDir("./")
	code := `
package parser

import "github.com/lawrsp/pigo/generatorutils"
import "github.com/lawrsp/pigo/generatorbuilder"
import "strings"
import "go/ast"
`
	file := p.ParseFileContent("_test", code)
	p.InsertFileToPackage(pkg, file, 1)
	// fmt.Println("packageName:", pkg.Name)
	// scope := pkg.Package.Scope
	// fmt.Println(pkg.Package.Name, ":", len(scope.Objects))
	// for k, v := range scope.Objects {
	// fmt.Println(k, ":Kind(", v.Kind, "):", v.Name)
	// }

	// file = pkg.Files[0]

	src := "Parser"
	fmt.Printf("parse: %s\n", src)
	expr, err := ParseExpr(src)
	expect(t.Errorf, err, nil)
	typ := file.ReduceType(expr)
	expect(t.Errorf, typ != nil, true)
	expect(t.Errorf, typ.File() != nil, true)
	expect(t.Errorf, typ.File().Name, "parser.go")

	src = "Parser.FileSet"
	fmt.Printf("parse: %s\n", src)
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	typ = file.ReduceType(expr)
	expect(t.Errorf, typ.File() != nil, true)
	expect(t.Errorf, typ.File().Name != "", true)
	expect(t.Errorf, typ.String(), "*token.FileSet")

	src = "Parser.ImportScope"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	typ = file.ReduceType(expr)
	fmt.Printf("parse: %s => %s\n", src, typ.String())
	fmt.Println(typ.String())
	expect(t.Errorf, typ.File() != nil, true)
	expect(t.Errorf, typ.File().Name != "", true)
	expect(t.Errorf, typ.String(), "parser.ImportScope(*parser.Parser,string,string)(string,*parser.Scope)")

	src = "[]int"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	typ = file.ReduceType(expr)
	fmt.Printf("parse: %s => %s\n", src, typ.String())
	expect(t.Errorf, typ.File() == nil, true)
	expect(t.Errorf, typ.String(), "[]int")

	src = "[][]string"
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	typ = file.ReduceType(expr)
	fmt.Printf("parse: %s => %s\n", src, typ.String())
	expect(t.Errorf, typ.File() == nil, true)
	expect(t.Errorf, typ.String(), "[][]string")

	src = "[]Parser"
	fmt.Printf("parse: %s\n", src)
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	typ = file.ReduceType(expr)
	expect(t.Errorf, typ.File() != nil, true)
	expect(t.Errorf, typ.String(), "[]parser.Parser")

	src = "[]*Parser"
	fmt.Printf("parse: %s\n", src)
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	typ = file.ReduceType(expr)
	expect(t.Errorf, typ.File() != nil, true)
	expect(t.Errorf, typ.String(), "[]*parser.Parser")

	// src = "printutil.IncreaseName(\"abc\")"
	// fmt.Printf("parse: %s\n", src)
	// expr, err = ParseExpr(src)
	// expect(t.Errorf, err, nil)
	// typ = file.ReduceType(expr)
	// expect(t.Errorf, typ.String(), "string")

	src = "strings.Split(\"abc\", \".\")[2]"
	fmt.Printf("parse: %s\n", src)
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	typ = file.ReduceType(expr)
	assert(t, typ.File() == nil, "%s.File() == nil", typ)
	expect(t.Errorf, typ.String(), "string")

	src = "Parser.Scope.Outer.Outer"
	fmt.Printf("parse: %s\n", src)
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	typ = file.ReduceType(expr)
	assert(t, typ != nil, "%s != nil", typ)
	assert(t, typ.File() != nil, "%s.File() != nil", typ)
	expect(t.Errorf, typ.String(), "*ast.Scope")

	src = "Type"
	fmt.Printf("parse: %s\n", src)
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	typ = file.ReduceType(expr)
	assert(t, typ.File() != nil, "%s.File() != nil", typ)
	expect(t.Errorf, typ.String(), "parser.Type")

	src = "Type.Package"
	fmt.Printf("parse: %s\n", src)
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	typ = file.ReduceType(expr)
	assert(t, typ != nil, "%s not found", src)
	// expect(t.Errorf, typ.File() != nil, true)
	assert(t, typ.File() == nil, "%s.File() == nil", typ)
	expect(t.Errorf, typ.String(), "Package()*parser.Package")

	src = "Type.Package()"
	fmt.Printf("parse: %s\n", src)
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	typ = file.ReduceType(expr)
	assert(t, typ.File() != nil, "%s.File() == nil", typ)
	expect(t.Errorf, typ.String(), "*parser.Package")

	src = "ast.Scope.Lookup"
	fmt.Printf("parse: %s\n", src)
	expr, err = ParseExpr(src)
	expect(t.Errorf, err, nil)
	typ = file.ReduceType(expr)
	assert(t, typ.File() != nil, "%s.File() != nil", typ)
	expect(t.Errorf, typ.String(), "ast.Lookup(*ast.Scope,string)*ast.Object")

	// src = "builder.Builder.Variables"
	// fmt.Printf("parse: %s\n", src)
	// expr, err = ParseExpr(src)
	// expect(t.Errorf, err, nil)
	// typ = file.ReduceType(expr)
	// assert(t, typ.File() == nil, "%s.File() == nil", typ)
	// expect(t.Errorf, typ.String(), "Variables()*builder.VariableList")

}

func TestTypeToString(t *testing.T) {
	p := NewParser()
	pkg := p.ParsePackageDir("./")
	code := `
package parser

type TypeToStringStruct struct {
  A map[string]interface{}
}

var TypeToStringInterface interface{}

type CustomerInterface interface{}

var myinterface CustomerInterface
`
	file := p.ParseFileContent("_test", code)
	p.InsertFileToPackage(pkg, file, 1)

	src := "TypeToStringStruct"
	expr, err := ParseExpr(src)
	expect(t.Errorf, err, nil)
	typ := file.ReduceType(expr)
	expect(t.Errorf, typ.String(), "parser.TypeToStringStruct")
	expect(t.Errorf, typ != nil, true)
	expect(t.Errorf, typ.File() != nil, true)
	expect(t.Errorf, typ.File().Name, "_test")

	src = "TypeToStringStruct.A"
	expr, err = ParseExpr(src)
	if err != nil {
		t.Errorf("parse %s failed: %v", src, err)
	}
	expect(t.Errorf, err, nil)
	typ = file.ReduceType(expr)
	fmt.Printf("%s: %s\n", src, typ)

	src = "TypeToStringInterface"
	expr, err = ParseExpr(src)
	if err != nil {
		t.Errorf("parse %s failed: %v", src, err)
	}
	typ = file.ReduceType(expr)
	if _, ok := typ.(*InterfaceType); !ok {
		t.Errorf("is not InterfaceType")
	}
	fmt.Printf("%s: %s\n", src, typ)

	src = "myinterface"
	expr, err = ParseExpr(src)
	if err != nil {
		t.Errorf("parse %s failed: %v", src, err)
	}
	typ = file.ReduceType(expr)
	fmt.Printf("%s: %s\n", src, typ)

}
