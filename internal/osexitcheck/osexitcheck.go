// Package osexitcheck define Analyzer, which check for call os.Exit
package osexitcheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var OSExitAnalyzer = &analysis.Analyzer{
	Name: "osexitcheck",
	Doc:  "check for call os.Exit",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.SelectorExpr:
				if i, ok := x.X.(*ast.Ident); ok && i.Name == "os" && x.Sel.Name == "Exit" {
					pass.Reportf(x.Pos(), "use os.Exit")
				}
			}
			return true
		})
	}
	return nil, nil
}
