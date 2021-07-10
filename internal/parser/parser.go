package parser

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"

	"golang.org/x/mod/modfile"
)

func LoadPackageNames(filename string) (string, error) {
	tok, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
	if err != nil {
		return "", err
	}

	return tok.Name.Name, nil
}

func LoadStructNames(filename string) ([]string, error) {
	tok, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
	if err != nil {
		return nil, err
	}

	var structs []string
	for _, decl := range tok.Decls {
		if v, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range v.Specs {
				if w, ok := spec.(*ast.TypeSpec); ok {
					if _, ok := w.Type.(*ast.StructType); ok {
						structs = append(structs, w.Name.Name)
					}
				}
			}
		}
	}

	return structs, nil
}

func LoadStructName(path string) ([]string, error) {
	tokMap, err := parser.ParseDir(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, err
	}

	var structs []string
	for _, tok := range tokMap {
		for _, file := range tok.Files {
			for _, decl := range file.Decls {
				if v, ok := decl.(*ast.GenDecl); ok {
					for _, spec := range v.Specs {
						if w, ok := spec.(*ast.TypeSpec); ok {
							if _, ok := w.Type.(*ast.StructType); ok {
								structs = append(structs, w.Name.Name)
							}
						}
					}
				}
			}
		}
	}

	return structs, nil
}

func LoadModuleName() (string, error) {
	if _, err := os.Stat("../go.mod"); os.IsNotExist(err) {
		return "", errors.New("go.mod required")
	}

	b, err := os.ReadFile("../go.mod")
	if err != nil {
		return "", err
	}

	modf, err := modfile.Parse("../go.mod", b, nil)
	if err != nil {
		return "", err
	}

	return modf.Module.Mod.String(), nil
}
