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
func main() {
	var analyzers []*analysis.Analyzer

	analyzers = append(analyzers,
		atomic.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
	)

	for _, v := range stylecheck.Analyzers {
		name := v.Analyzer.Name
		if len(name) >= 2 && name[0] == 'S' && name[1] == 'A' {
			analyzers = append(analyzers, v.Analyzer)
		}
		if name == "ST1000" {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	analyzers = append(analyzers,
		errcheck.Analyzer, // Проверяет непроверенные ошибки
	)
	analyzers = append(analyzers, Analyzer)
	multichecker.Main(analyzers...)

}
