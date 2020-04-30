package parser

import (
	"go/ast"
	"strconv"
	"strings"
)

type File struct {
	parser *Parser

	BelongTo *Package

	Name string
	File *ast.File

	dotImports  []*Package
	nameImports map[string]*Package
}

func nameOfImport(ispc *ast.ImportSpec) string {
	//compare declared name
	if ispc.Name != nil {
		return ispc.Name.Name
	}

	path, err := strconv.Unquote(ispc.Path.Value)
	if err != nil {
		panic("import path error")
	}

	//compare default name
	//"xxx/xxx/name"
	all := strings.Split(path, "/")
	return all[len(all)-1]
}

func (f *File) FindImportPath(name string) (path string, ok bool) {
	for _, ispc := range f.File.Imports {
		//find the path
		ok = false

		//compare declared name
		if ispc.Name != nil {
			if ispc.Name.Name != name {
				continue
			}
			ok = true
		}

		path, err := strconv.Unquote(ispc.Path.Value)
		if err != nil {
			panic("import path error")
		}

		//compare default name
		if !ok {
			//"xxx/xxx/name"
			all := strings.Split(path, "/")
			if name != all[len(all)-1] {
				continue
			}
			ok = true
		}

		return path, true
	}

	return "", false
}

func (file *File) AddImport(name string, pkg *Package) {
	if name == "." {
		// log.Printf("import dot package: %v", pkg.Path)
		file.dotImports = append(file.dotImports, pkg)
		return
	}

	if file.nameImports == nil {
		file.nameImports = map[string]*Package{}
	}

	// log.Printf(">>>>>>>add import: %v ===== %v(%v) ", file.Name, pkg.Name, pkg.Path)
	file.nameImports[name] = pkg
}

func (file *File) FindImport(name string) *Package {
	if name == "" {
		return file.BelongTo
	}

	return file.nameImports[name]
}

func (file *File) FindImportNameByPath(path string) string {
	if path == "" {
		return ""
	}

	for _, ispc := range file.File.Imports {
		imptPath, err := strconv.Unquote(ispc.Path.Value)
		if err != nil {
			panic("import path error")
		}

		if imptPath != path {
			continue
		}

		if ispc.Name != nil {
			return ispc.Name.Name
		}

		names := strings.Split(imptPath, "/")
		return names[len(names)-1]
	}

	return ""
}

// GenDecl
// FuncDecl
func (file *File) FindDecl(expr ast.Expr, finder DeclFinder) *DeclNode {

	pkgName := ""
	name := ""

	switch t := expr.(type) {
	case *ast.StarExpr:
		return file.FindDecl(t.X, finder)
	case *ast.SelectorExpr:
		//donnot support more nest selector
		pkgName = t.X.(*ast.Ident).Name
		name = t.Sel.Name
	case *ast.Ident:
		name = t.Name
		if checkNameInternal(name) {
			return nil
		}
	// case *ast.CallExpr:
	default:
		return nil
	}

	// log.Printf("find `%v`.`%v` in package(%v)", pkgName, name, p.pkg.Dir)
	//pkgName != ""
	//@TODO: Struct ?
	if pkgName != "" {
		//package
		path, ok := file.FindImportPath(pkgName)
		if !ok {
			// not in this file
			return nil
		}
		pkg := file.parser.ImportPackage(pkgName, path, file.BelongTo.Dir)
		// _ = p.ImportPackageByName(srcPkg, pkgName)
		file.AddImport(pkgName, pkg)

		return pkg.FindDecl(name, finder.WithName(name))
	}

	//pkgName == ""
	// in this file
	for _, node := range file.File.Decls {
		if result := finder.WithName(name).Find(node); result != nil {
			if result == nil {
				continue
			}

			if result.Alias == nil {
				// ast.Print(token.NewFileSet(), result)
				return result.InFile(file)
			}

			//type alias
			// ast.Print(token.NewFileSet(), changed)
			next := file.FindDecl(result.Alias, finder)
			return next.WithPre(result)
		}
	}

	return nil
}

func (file *File) DotImported() []*Package {
	return file.dotImports
}

func (f *File) FindAllDotImport() (paths []string) {
	for _, ispc := range f.File.Imports {
		//compare declared name
		if ispc.Name != nil && ispc.Name.Name == "." {
			path, err := strconv.Unquote(ispc.Path.Value)
			if err != nil {
				panic("import path error")
			}

			paths = append(paths, path)
		}
	}
	return
}

func (f *File) ImportAllDots() []*Package {
	pkgs := []*Package{}
	if paths := f.FindAllDotImport(); len(paths) > 0 {
		for _, path := range paths {
			pkg := f.parser.ImportPackage(f.BelongTo.Name, ".", path)
			pkgs = append(pkgs, pkg)
			f.AddImport(".", pkg)
		}
	}

	return pkgs
}
