package parser

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"strings"
)

type Parser struct {
	FileSet       *token.FileSet
	Scope         *ast.Scope
	ImportContext build.Context
	Scopes        map[string]*Scope
}

func NewParser() *Parser {
	return &Parser{
		FileSet:       token.NewFileSet(),
		Scope:         UniverseScope,
		ImportContext: build.Default,
		Scopes:        map[string]*Scope{},
	}
}

func (p *Parser) ImportScope(path string, dir string) (canonicalPath string, scope *Scope) {
	if path == "" {
		return "", nil
	}

	buildPkg, err := p.ImportContext.Import(path, dir, 0)
	if err != nil {
		log.Fatalf("cannot import package: %s, %s:%v", path, dir, err)
	}

	importPath := buildPkg.ImportPath //path

	//history equal
	if importPath == "golang.org/x/net/context" {
		importPath = "context"
	}

	if buildPkg.ImportComment != "" && buildPkg.ImportComment != buildPkg.ImportPath {
		importPath = buildPkg.ImportComment
	}

	if alt := p.Scopes[importPath]; alt != nil {
		return importPath, alt
	}

	return importPath, nil
}

func (p *Parser) AddScope(canonicalPath string, scope *Scope) {
	// log.Printf("!=====add %s: %p", canonicalPath, scope)
	p.Scopes[canonicalPath] = scope
}

func (p *Parser) PrintNode(node ast.Node) {
	ast.Print(p.FileSet, node)
}

func (p *Parser) ParsePackageDir(directory string) *Package {
	pkg, err := build.Default.ImportDir(directory, 0)
	if err != nil {
		log.Fatalf("cannot process directory %s: %s", directory, err)
	}
	var names []string
	names = append(names, pkg.GoFiles...)
	// TODO: Need to think about constants in test files. Maybe write type_string_test.go
	// in a separate pass? For later.
	// names = append(names, pkg.TestGoFiles...) // These are also in the "foo" package.
	//	names = append(names, pkg.SFiles...)
	names = prefixDirectory(directory, names)
	return p.ParsePackage("", directory, names)
}

// parsePackageFiles parses the package occupying the named files.
func (p *Parser) ParsePackageFiles(names []string) *Package {
	return p.ParsePackage("", ".", names)
}

func (p *Parser) ParseFileContent(name string, content interface{}) *File {
	parsedFile, err := parser.ParseFile(p.FileSet, name, content, parser.ParseComments)
	if err != nil {
		log.Fatalf("parsing package: %s: %s", name, err)
	}

	return &File{
		Name: name,
		File: parsedFile,
	}
}

func (p *Parser) InsertFileToPackage(pkg *Package, file *File, index int) {
	pkg.InsertFile(file, index)
	astFiles := make(map[string]*ast.File, len(pkg.Files))
	for _, f := range pkg.Files {
		astFiles[f.File.Name.Name] = f.File
	}

	return
}

// prefixDirectory places the directory name on the beginning of each name in the list.
func prefixDirectory(directory string, names []string) []string {
	if directory == "." {
		return names
	}
	ret := make([]string, len(names))
	for i, name := range names {
		ret[i] = filepath.Join(directory, name)
	}
	return ret
}

// parsePackage analyzes the single package constructed from the named files.
func (p *Parser) ParsePackage(path string, directory string, names []string) *Package {

	mod := parser.ParseComments
	var files = []*File{}
	for _, name := range names {
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		parsedFile, err := parser.ParseFile(p.FileSet, name, nil, mod)
		if err != nil {
			log.Fatalf("parsing package: %s: %s", name, err)
		}
		// astFiles[name] = parsedFile
		files = append(files, &File{
			parser: p,
			Name:   name,
			File:   parsedFile,
			// BelongTo: pkg,
		})
	}
	if len(files) == 0 {
		log.Fatalf("%s: no buildable Go files", directory)
	}

	name := files[0].File.Name.Name
	pkg := NewPackage(p, name, directory, path, files)

	return pkg
}

func (p *Parser) ImportPackage(name string, path string, srcDir string) *Package {
	buildPkg, err := build.Default.Import(path, srcDir, 0)
	if err != nil {
		panic(err)
	}
	dir := buildPkg.Dir
	names := []string{}
	names = append(names, buildPkg.GoFiles...)

	pkg := p.ParsePackage(path, dir, prefixDirectory(dir, names))
	pkg.Name = name
	return pkg
}

func ParseExpr(x string) (ast.Expr, error) {
	return parser.ParseExprFrom(token.NewFileSet(), "", []byte(x), 0)
}
