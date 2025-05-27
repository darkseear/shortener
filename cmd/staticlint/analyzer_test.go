package staticlint

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestNoOsExitInMain(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "main")
}
