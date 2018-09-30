package fixturegenerator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// structType models a golang struct.
type structType struct {
	name   string
	fields []*structField
}

// structField models a golang struct field.
type structField struct {
	name   string
	goType ast.Expr
}

// GetImports extracts the imports from a golang package and returns a mapping between
// the package name and the path to that package.
func GetImports(dirPath string) (map[string]string, error) {
	fset := token.NewFileSet()

	packageASTMapping, err := parser.ParseDir(fset, dirPath, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	imports := map[string]string{}
	for _, packageAST := range packageASTMapping {
		for _, fileAST := range packageAST.Files {
			for _, imp := range fileAST.Imports {
				if imp.Name != nil {
					imports[imp.Name.Name] = strings.Trim(imp.Path.Value, `"`)
					continue
				}
				path := imp.Path.Value
				tokens := strings.Split(path, "/")
				imports[strings.Trim(tokens[len(tokens)-1], `"`)] = strings.Trim(imp.Path.Value, `"`)
			}
		}
	}

	return imports, nil
}

// GetStruct extracts a struct identified by name from a directory.
func GetStruct(dirPath, name string) (*structType, error) {
	fset := token.NewFileSet()

	packageASTMapping, err := parser.ParseDir(fset, dirPath, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	for _, packageAST := range packageASTMapping {
		for _, fileAST := range packageAST.Files {
			for _, s := range fileAST.Decls {
				genDecl, ok := s.(*ast.GenDecl)
				if !ok {
					// Not a generic declaration Node. It can't be a struct.
					continue
				}
				if genDecl.Tok != token.TYPE {
					// Not a type declaration so it can't be a struct.
					continue
				}
				// We know for sure (from docs) that this is a TypeSpec.
				spec := genDecl.Specs[0].(*ast.TypeSpec)

				_, ok = spec.Type.(*ast.StructType)
				if !ok {
					// Not a struct, we can ignore.
					continue
				}
				if spec.Name.Name != name {
					// Not the struct we want.
					continue
				}

				return typeSpecToStructType(spec)
			}
		}
	}

	return nil, fmt.Errorf("Struct %s does not exist", name)
}

// typeSpecToStructType transforms an internal go struct spec to our own representation.
func typeSpecToStructType(spec *ast.TypeSpec) (*structType, error) {
	expr, ok := spec.Type.(*ast.StructType)
	if !ok {
		// Not a struct, we can ignore.
		return nil, fmt.Errorf("TypeSpec is not a struct")
	}

	fields := []*structField{}
	for _, goField := range expr.Fields.List {
		fields = append(fields, &structField{name: goField.Names[0].Name, goType: goField.Type})
	}

	return &structType{name: spec.Name.Name, fields: fields}, nil
}
