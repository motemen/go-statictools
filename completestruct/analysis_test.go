package completestruct

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestRun(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "a")
}
