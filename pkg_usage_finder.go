package watchman

import (
	"go/build"
	"go/parser"
	"go/token"
	"strings"
)

/*
getImportPaths returns the list of packages imported by the package present at 'path'
*/
func getImportPaths(path string) ([]string, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	imports := make([]string, 0)
	for _, pkg := range pkgs {
		for _, pkgFile := range pkg.Files {
			for _, fileImport := range pkgFile.Imports {
				importPath := fileImport.Path.Value
				imports = append(imports, importPath)
			}
		}
	}

	return imports, nil
}

/*
getPathsForNonRootPkgs filters the given package list and returns the complete dir paths
for all non-root (non std-lib) pkgs
*/
func getPathsForNonRootPkgs(importPaths []string) []string {
	nonRootPkgs := make([]string, 0)

	for _, path := range importPaths {
		cleanedPath := strings.Trim(path, "\"")

		pkg, err := build.Import(cleanedPath, "", build.FindOnly)
		if err != nil {
			continue
		}

		if !pkg.Goroot {
			nonRootPkgs = append(nonRootPkgs, pkg.Dir)
		}
	}

	return nonRootPkgs
}
