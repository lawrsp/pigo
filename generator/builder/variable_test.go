package builder

import (
	"fmt"
	"go/ast"
	"go/token"
	"testing"

	"github.com/lawrsp/pigo/generator/parser"
)

func ExampleVariableDot() {
	v := NewVariable(parser.ErrorType()).WithName("err")
	expr := v.DotExpr("Error")
	fs := token.NewFileSet()
	ast.Print(fs, expr)
	//Output:
	//0  *ast.SelectorExpr {
	//      1  .  X: *ast.Ident {
	//      2  .  .  NamePos: -
	//      3  .  .  Name: "err"
	//      4  .  }
	//      5  .  Sel: *ast.Ident {
	//      6  .  .  NamePos: -
	//      7  .  .  Name: "Error"
	//      8  .  .  Obj: *ast.Object {
	//      9  .  .  .  Kind: bad
	//     10  .  .  .  Name: ""
	//     11  .  .  }
	//     12  .  }
	//     13  }
}

func ExampleVariableDot2() {
	v := NewVariable(parser.ErrorType()).WithName("err")
	expr := v.DotExpr("Error()")
	fs := token.NewFileSet()
	ast.Print(fs, expr)
	//Output:
	//0  *ast.CallExpr {
	//      1  .  Fun: *ast.SelectorExpr {
	//      2  .  .  X: *ast.Ident {
	//      3  .  .  .  NamePos: -
	//      4  .  .  .  Name: "err"
	//      5  .  .  }
	//      6  .  .  Sel: *ast.Ident {
	//      7  .  .  .  NamePos: -
	//      8  .  .  .  Name: "Error"
	//      9  .  .  .  Obj: *ast.Object {
	//     10  .  .  .  .  Kind: bad
	//     11  .  .  .  .  Name: ""
	//     12  .  .  .  }
	//     13  .  .  }
	//     14  .  }
	//     15  .  Lparen: -
	//     16  .  Ellipsis: -
	//     17  .  Rparen: -
	//     18  }
}

func ExampleVariableDot3() {
	v := NewVariable(parser.ErrorType()).WithName("err")
	expr := v.DotExpr("a.b")
	ast.Print(token.NewFileSet(), expr)
	//Output:
	//0  *ast.SelectorExpr {
	//      1  .  X: *ast.SelectorExpr {
	//      2  .  .  X: *ast.Ident {
	//      3  .  .  .  NamePos: -
	//      4  .  .  .  Name: "err"
	//      5  .  .  }
	//      6  .  .  Sel: *ast.Ident {
	//      7  .  .  .  NamePos: -
	//      8  .  .  .  Name: "a"
	//      9  .  .  .  Obj: *ast.Object {
	//     10  .  .  .  .  Kind: bad
	//     11  .  .  .  .  Name: ""
	//     12  .  .  .  }
	//     13  .  .  }
	//     14  .  }
	//     15  .  Sel: *ast.Ident {
	//     16  .  .  NamePos: -
	//     17  .  .  Name: "b"
	//     18  .  }
	//     19  }

}

func TestVariableDotType(t *testing.T) {

	code := `
package tt

type B struct {
    Si int
}

type A struct {
   Bs B
}

`
	p := parser.NewParser()

	file := p.ParseFileContent("test", code)
	pkg := parser.NewPackage(p, "fake", "./fake.go", "", []*parser.File{file})
	file = pkg.Files[0]

	ta := file.ReduceType(ast.NewIdent("A"))
	tb := file.ReduceType(ast.NewIdent("B"))

	va := NewVariable(ta).WithName("ta").ReadOnly()
	t1 := va.DotTypeInFile("Bs.Si", file)

	vb := NewVariable(tb).WithName("tb").ReadOnly()
	t2 := vb.DotTypeInFile("Si", file)

	fmt.Println(t1)
	fmt.Println(t2)
	expect(t1, t2, t.Errorf)
}
