package staticlint

import (
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/stylecheck"
)

// Добавляем стандартные анализаторы из golang.org/x/tools и добавляем свой.
//
//go:generate go mod tidy
func main() {
	var analyzers []*analysis.Analyzer

	analyzers = append(analyzers,
		atomic.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
	)

	for _, v := range stylecheck.Analyzers {
		if v.Analyzer.Name[0] == 'S' && v.Analyzer.Name[1] == 'A' {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	for _, v := range stylecheck.Analyzers {
		if v.Analyzer.Name == "ST1000" {
			analyzers = append(analyzers, v.Analyzer)
			break
		}
	}
	analyzers = append(analyzers,
		errcheck.Analyzer, // Проверяет непроверенные ошибки
	)
	analyzers = append(analyzers, Analyzer)
	multichecker.Main(analyzers...)

}
