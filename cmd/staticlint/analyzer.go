package staticlint

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer настраивает и возвращает анализатор для проверки вызовов os.Exit.
//
// Анализатор проверяет только функции main в пакете main. При обнаружении
// прямого вызова os.Exit в этих функциях выдает предупреждение.
var Analyzer = &analysis.Analyzer{
	Name:     "noosexit",
	Doc:      "reports direct os.Exit calls in main package",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// run реализует логику анализатора.
//
// Функция проверяет, что анализируемый пакет является main, находит
// функцию main и проверяет её тело на наличие вызовов os.Exit.
// При обнаружении таких вызовов генерирует диагностическое сообщение "os.Exit запрещен в функции main".
func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil)}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)
		if fn.Name.Name != "main" {
			return
		}

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			if id, ok := sel.X.(*ast.Ident); ok && id.Name == "os" && sel.Sel.Name == "Exit" {
				pass.Reportf(call.Pos(), "os.Exit запрещен в функции main")
			}
			return true
		})
	})

	return nil, nil
}
