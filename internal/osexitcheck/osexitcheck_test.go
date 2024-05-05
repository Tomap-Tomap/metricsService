package osexitcheck

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestOSExitCheckAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), OSExitAnalyzer, "./...")
}
