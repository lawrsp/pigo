package printutil

import (
	"bytes"
	"go/ast"
	"go/token"
	"log"
)

func FatalNodef(node interface{}, format string, args ...interface{}) {
	var buf bytes.Buffer
	_ = ast.Fprint(&buf, token.NewFileSet(), node, ast.NotNilFilter)
	log.Printf(format, args...)
	log.Fatalf("%s", buf.String())
}

func PrintNodef(node interface{}, format string, args ...interface{}) {
	var buf bytes.Buffer
	_ = ast.Fprint(&buf, token.NewFileSet(), node, ast.NotNilFilter)
	log.Printf(format, args...)
	log.Printf("%s", buf.String())
}
