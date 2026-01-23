package analyzer

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const (
	analyzerName = "forbiddencalls"
	analyzerDoc  = "reports usage of panic, log.Fatal, and os.Exit outside main function"
)

// Analyzer checks for forbidden function calls (panic, log.Fatal, os.Exit) in the code.
var Analyzer = &analysis.Analyzer{
	Name:     analyzerName,
	Doc:      analyzerDoc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	insp.Preorder(nodeFilter, func(node ast.Node) {
		callExpr := node.(*ast.CallExpr)
		checkCall(pass, callExpr)
	})

	return nil, nil
}

func checkCall(pass *analysis.Pass, callExpr *ast.CallExpr) {
	switch fn := callExpr.Fun.(type) {
	case *ast.Ident:
		if fn.Name == "panic" {
			pass.Reportf(callExpr.Pos(), "panic is forbidden")
		}
	case *ast.SelectorExpr:
		checkSelectorExpr(pass, fn, callExpr)
	}
}

func checkSelectorExpr(pass *analysis.Pass, selectorExpr *ast.SelectorExpr, callExpr *ast.CallExpr) {
	if ident, ok := selectorExpr.X.(*ast.Ident); ok {
		fn := selectorExpr.Sel.Name

		if pass.TypesInfo == nil {
			return
		}

		obj := pass.TypesInfo.Uses[ident]
		if obj == nil {
			return
		}

		pkgName, ok := obj.(*types.PkgName)
		if !ok {
			return
		}

		pkgPath := pkgName.Imported().Path()

		switch {
		case pkgPath == "log" && fn == "Fatal":
			if !isInMainFunction(pass, callExpr) {
				pass.Reportf(callExpr.Pos(), "log.Fatal is forbidden outside main function")
			}
		case pkgPath == "os" && fn == "Exit":
			if !isInMainFunction(pass, callExpr) {
				pass.Reportf(callExpr.Pos(), "os.Exit is forbidden outside main function")
			}
		}
	}
}

func isInMainFunction(pass *analysis.Pass, node ast.Node) bool {
	for _, f := range pass.Files {
		for _, decl := range f.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				if funcDecl.Name.Name == "main" && isNodeInsideFunc(node, funcDecl) {
					return true
				}
			}
		}
	}
	return false
}

func isNodeInsideFunc(target ast.Node, funcDecl *ast.FuncDecl) bool {
	found := false
	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if n == target {
			found = true
			return false
		}
		return true
	})
	return found
}
