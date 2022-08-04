package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io/ioutil"
	"log"
	"os"

	"github.com/lawrsp/pigo/pkg/builder"
	"github.com/lawrsp/pigo/pkg/parser"
	"golang.org/x/tools/imports"
)

type Generator struct {
	Parser *parser.Parser

	File *parser.File
	Pkg  *parser.Package

	workFile *parser.File
}

func (g *Generator) WorkFile() *parser.File {
	if g.workFile != nil {
		return g.workFile
	}
	return g.File
}

func (g *Generator) ReduceType(expr ast.Expr) parser.Type {
	if file := g.WorkFile(); file != nil {
		if t := file.ReduceType(expr); t != nil {
			return t
		}
	}
	_, t := g.Pkg.ReduceType(expr)
	return t
}
func (g *Generator) ReduceTypeSrc(src string) parser.Type {
	expr, err := parser.ParseExpr(src)
	if err == nil {
		return g.ReduceType(expr)
	}
	log.Fatalf("expr error: %s:%v", src, err)
	return nil
}

// Bytes return the generated bytes
func (g *Generator) Bytes() []byte {

	var buf bytes.Buffer
	// fmt.Fprintf(&buf, "// Code generated by \"%s\"; DO NOT EDIT.\n", strings.Join(os.Args[0:], " "))
	if err := format.Node(&buf, token.NewFileSet(), g.File.File); err != nil {
		log.Fatalf("generate code error: %v", err)
	}

	return buf.Bytes()
}

func (g *Generator) GetExprString(expr ast.Expr) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, token.NewFileSet(), expr); err != nil {
		log.Fatalf("generate code error: %v", err)
	}

	return buf.String()
}

func (g *Generator) GetExprValueString(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.BasicLit:
		if x.Kind == token.STRING {
			return fmt.Sprintf("\"%s\"", x.Value)
		} else {
			return x.Value
		}
	}

	return g.GetExprString(expr)
}

// format returns the gofmt-ed contents of the Generator's buffer.
func (g *Generator) Format() ([]byte, error) {

	options := &imports.Options{
		Fragment:  false,
		AllErrors: true,

		TabWidth:  2,
		TabIndent: true,
		Comments:  true,
	}
	res, err := imports.Process("", g.Bytes(), options)
	// src, err := format.Source(g.buf.Bytes())
	if err != nil {
		// Should never happen, but can arise when developing this code.
		// The user can compile the output to see the error.
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return nil, err
	}
	return res, nil
}

func (g *Generator) PrepareParser() {
	g.Parser = parser.NewParser()
}
func (g *Generator) PreparePackage(dir string, output string) {
	p := g.Parser
	pkg := p.ParsePackageDir(dir)

	var file *parser.File

	if output != "" {
		file = pkg.GetFile(output)
	}

	if file == nil {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "package %v\n", pkg.Name)
		// add custom imports
		fmt.Fprintf(&buf, "import ()\n")
		file = p.ParseFileContent("_prepare", buf.Bytes())
		p.InsertFileToPackage(pkg, file, 0)
	}

	g.File = file
	g.Pkg = pkg
}

func (g *Generator) Prepare(dir string, files []string, output string) {
	g.PrepareParser()

	p := g.Parser
	var pkg *parser.Package

	if len(files) == 0 {
		pkg = p.ParsePackageDir(dir)
	} else {
		if output != "" {
			if ok, err := PathExists(output); err == nil && ok {
				files = append(files, output)
			}
		}

		pkg = p.ParsePackageFiles(files)
	}
	var file *parser.File

	if output != "" {
		file = pkg.GetFile(output)
	}

	if file == nil {
		log.Printf("=======new file %s", output)
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "package %v\n", pkg.Name)
		// add custom imports
		fmt.Fprintf(&buf, "import ()\n")
		file = p.ParseFileContent("_prepare", buf.Bytes())
		p.InsertFileToPackage(pkg, file, 0)
	}

	g.workFile = nil
	g.File = file
	g.Pkg = pkg
}

func (g *Generator) PrepareImports(imports map[string]string) {
	bd := builder.NewFile(nil, g.File)
	if len(imports) > 0 {
		for name, path := range imports {
			bd.AddImport(name, path)
		}
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (g *Generator) PrepareWithFile(fileName string, output string) {
	g.PrepareParser()
	p := g.Parser

	names := []string{fileName}
	var pkg *parser.Package

	outputExists := false

	if output != "" {
		exists, err := PathExists(output)
		if err != nil {
			log.Fatalf("check output(%s) exists failed %s", output, err)
		}
		if exists {
			names = append(names, output)
			outputExists = true
		}
	}
	pkg = p.ParsePackageFiles(names)

	var file *parser.File
	if outputExists {
		file = pkg.GetFile(output)
	}

	if file == nil {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "package %v\n", pkg.Name)
		// add custom imports
		fmt.Fprintf(&buf, "import ()\n")
		file = p.ParseFileContent("_prepare", buf.Bytes())
		p.InsertFileToPackage(pkg, file, 0)
	}

	g.workFile = pkg.GetFile(fileName)
	g.File = file
	g.Pkg = pkg

}

func (g *Generator) Output(output string) {
	// fmt.Printf("%s: %s", output, string(g.Bytes()))
	// Format the output.
	result, err := g.Format()
	if err != nil {
		log.Fatalf("Foramt Failed: %v", err)
	}

	// Write to stdout / file
	if len(output) == 0 {
		fmt.Println(string(result))
	} else {
		err = ioutil.WriteFile(output, result, 0644)
		if err != nil {
			log.Fatalf("writing output: %s", err)
		}
	}
}
