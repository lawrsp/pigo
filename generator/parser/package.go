package parser

import (
	"fmt"
	"go/ast"
	"log"
	"strings"
)

type Package struct {
	Scope         *Scope
	Name          string
	Dir           string
	Path          string
	CanonicalPath string
	PackageName   string
	Files         []*File
	parser        *Parser
}

func (pkg *Package) EqualTo(np *Package) bool {
	return pkg.Scope == np.Scope
}

func NewPackage(p *Parser, name string, dir string, path string, files []*File) *Package {
	pkg := &Package{}
	astFiles := map[string]*ast.File{}
	pkgFiles := []*File{}
	for _, f := range files {
		cf := &File{}
		*cf = *f
		cf.BelongTo = pkg
		pkgFiles = append(pkgFiles, cf)
		astFiles[f.Name] = cf.File
	}

	canonicalPath, scope := p.ImportScope(path, dir)
	if scope == nil {
		scope = NewScope()
		p.AddScope(canonicalPath, scope)
	}

	if scope == nil {
		log.Fatalf("cannot create package %s:%s", name, path)
	}

	pkg.Scope = scope
	pkg.Name = name
	pkg.Dir = dir
	pkg.Path = path
	pkg.CanonicalPath = canonicalPath
	pkg.Files = pkgFiles
	pkg.parser = p
	if len(pkgFiles) > 0 {
		pkg.PackageName = pkgFiles[0].File.Name.Name
	} else {
		pkg.PackageName = name
	}

	return pkg
}

func (pkg *Package) InsertFile(file *File, index int) {
	var result []*File

	file.BelongTo = pkg
	file.parser = pkg.parser

	if index >= len(pkg.Files) {
		result = append(pkg.Files, file)
	} else if index == 0 {
		result = append([]*File{file}, pkg.Files...)
	} else {
		for i, f := range pkg.Files {
			if i == index {
				result = append(result, file)
			}
			result = append(result, f)
		}
	}

	pkg.Files = result

	return
}

func (this *Package) mergePackage(pkg *Package) {
	if pkg == this {
		return
	}

	if pkg.Name != this.Name {
		log.Fatalf("different package name")
	}

	for _, f := range pkg.Files {
		pkg.mergeFile(f)
	}

	return
}

func (pkg *Package) mergeFile(f *File) {
	for i, ef := range pkg.Files {
		if f.Name == ef.Name {
			f.BelongTo = pkg
			pkg.Files[i] = f
			return
		}
	}

	pkg.Files = append(pkg.Files, f)
}

func (pkg *Package) GetFile(name string) *File {
	for _, f := range pkg.Files {
		fnames := strings.Split(f.Name, "/")
		if name == fnames[len(fnames)-1] {
			return f
		}
	}
	return nil
}

//FindDecl
func (pkg *Package) FindDecl(name string, finder DeclFinder) *DeclNode {
	idents := strings.Split(name, ".")
	if len(idents) > 2 {
		log.Fatalf("not support name: %s", name)
	}

	//find in file
	expr, err := ParseExpr(name)
	if err != nil {
		log.Fatalf("parse expr(%s) failed: %v", name, err)
	}

	for _, f := range pkg.Files {
		if finded := f.FindDecl(expr, finder); finded != nil {
			return finded
		}
	}
	return nil
}

//findFuncDecl
func (pkg *Package) FindFuncDecl(name string) *DeclNode {
	idents := strings.Split(name, ".")
	if len(idents) == 1 {
		return pkg.FindDecl(name, NewFuncDeclFinder(""))
	}
	if len(idents) == 3 {
		stName := fmt.Sprintf("%s.%s", idents[0], idents[1])

		node := pkg.FindDecl(stName, NewValueSpecFinder())
		if node == nil {
			return nil
		}

		typ := ParseType(node.Node)
		receiverName := typ.Name()
		funName := idents[2]
		// log.Printf("find receiver: %s, funName: %s", receiverName, funName)

		return node.File.BelongTo.FindDecl(funName, NewFuncDeclFinder(receiverName))
	}

	if len(idents) == 2 {
		//as struct
		if node := pkg.FindDecl(idents[1], NewFuncDeclFinder(idents[0])); node != nil {
			return node
		}
		//as package
		if node := pkg.FindDecl(name, NewFuncDeclFinder("")); node != nil {
			return node
		}
	}

	log.Fatalf("not supported function name: %s", name)
	return nil
}

/*
//FindDecl
func (pkg *Package) FindName(idents []string) *ast.Object {

	name := idents[0]
	//find in file
	expr, err := ParseExpr(idents[0])
	if err != nil {
		log.Fatalf("parse expr(%s) failed: %v", name, err)
	}

	// (), [], {}, *, &
	isFunction := false

	for it := expr; it != nil; {
		switch t := expr.(type) {
		case *ast.Ident:
			name = t.Name
			it = nil
		case *ast.StarExpr:
			//*Atype
			it = t.X
		case *ast.CallExpr:
			//func(x)
			it = t.Fun
			isFunction = true
		case *ast.IndexExpr:
			//A[10]
			it = t.X //array map ==> element
		case *ast.CompositeLit:
			it = t.Type
		}
	}
	if isFunction {
		log.Printf("isFunction")
	}

	return nil
}

//type decl
func (pkg *Package) ParseType(src string) Type {
	idents := strings.Split(name, ".")

	switch len(idents) {
	case 1:
		//name
	case 2:
		//package.name
		//struct.field
		//receiver.func  (receiver == variable || recevier == struct)
	case 3:
		//package.receiver.func  (receiver == variable || recevier == struct)
		//package.struct.field
	}

}


//FindName

func findTypeMethod(scope *ast.Scope, name string, typeName string) (exists bool, result *ast.Object) {
	result = scope.Lookup(name)
	if result == nil {
		return
	}

	exists = true
	if result.Kind == ast.Fun {
		if fnt, ok := result.Type.(*ast.FuncDecl); ok {
			if fdlist := fnt.Recv; fdlist != nil && len(fdlist.List) == 1 {
				if idents := fdlist.List[0].Names; len(idents) == 1 {
					if idents[0].Name == typeName {
						return
					}
				}
			}
		}
	}

	result = nil

	return
}

func findTypeField(st *ast.StructType, name string) (exists bool, result *ast.Object) {
	anonymousFields := []*ast.Field{}
	for _, fd := range st.Fields.List {
		if fd.Names == nil {
			anonymousFields = append(anonymousFields, fd)
		}

		if fd.Names[0].Name == name {
			return
		}
	}

	return
}

func typeFromValueSpec(spec *ast.ValueSpec) *ast.TypeSpec {
	// if spec.Type != nil {
	// 	typ := ParseType(spec.Type)
	// } else {
	// 	typ := ParseType(spec.Values[0])
	// }

	return nil
}

func findTypeFieldOrMethod(scope *ast.Scope, spec *ast.TypeSpec, name string) (exists bool, result *ast.Object) {
	if spec.Assign.IsValid() {
		//type Alias
		return
	}

	if st, ok := spec.Type.(*ast.StructType); ok {
		//type.field
		if exists, result = findTypeField(st, name); exists {
			log.Printf("   field: %s, finded: %v ", name, result != nil)
			return
		}
		//type.method
		if exists, result = findTypeMethod(scope, name, spec.Name.Name); exists {
			log.Printf("   method: %s, finded: %v", name, result != nil)
			return
		}

	}

	return
}

// Pkg                // package
// Con                // constant
// Typ                // type
// Var                // variable
// Fun                // function or method

func getType(expr ast.Expr) ast.Expr {
	return nil
}

func (pkg *Package) findName(name string) *ast.Object {

	idents := strings.Split(name, ".")

	// obj := pkg.Scope.Lookup(idents[0])
	// idents = idents[1:]
	// if len(idents) == 0 || obj == nil {
	// 	return obj
	// }

	// var typeSpec *ast.TypeSpec

	// if obj.Kind == ast.Var {
	// 	spec := obj.Data.(*ast.ValueSpec)
	// 	var typeExpr ast.Expr
	// 	if spec.Type != nil {
	// 		typeExpr = getType(spec.Type)
	// 	} else if len(spec.Values) > 0 {
	// 		typeExpr = getType(spec.Values[0])
	// 	}
	// } else if obj.Kind == ast.Typ {
	// 	typeSpec = obj.Data.(*ast.TypeSpec)
	// }

	switch len(idents) {
	case 1:
		//name
		obj := pkg.Scope.Lookup(name)
		return obj
	case 2:
		obj := pkg.Scope.Lookup(idents[0])
		if obj != nil {
			if obj.Kind == ast.Typ {
				//type.method, type.field
				log.Printf("find type %s", obj.Name)
				utils.PrintNodef(obj.Decl, "")
				if st, ok := obj.Decl.(*ast.TypeSpec); ok {
					_, result := findTypeFieldOrMethod(pkg.Scope, st, idents[1])
					return result
				}
				return nil
			} else if obj.Kind == ast.Var {
				//value.method, value.field
				if spec, ok := obj.Data.(*ast.ValueSpec); ok {
					if st := typeFromValueSpec(spec); st != nil {
						_, result := findTypeFieldOrMethod(pkg.Scope, st, idents[1])
						return result
					}
				}
			}
		}
	}

	return nil
}
*/
