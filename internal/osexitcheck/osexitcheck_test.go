package osexitcheck

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestOSExitCheckAnalyzer(t *testing.T) {
	t.Run("positive test", func(t *testing.T) {
		fm := map[string]string{
			"a/main.go": `
			package main
	
			import (
				"fmt"
				"os"
			)
			
			func main() {
				os.Exit(0) // want "use os.Exit"
				fmt.Print("test") // want
			}
			`,
			"b/pkg.go": `
			package pkg1
	
			import (
				"fmt"
				"os"
			)
	
			func osExitCheckFunc() {
				os.Exit(0) // want
				fmt.Print("test") // want
			}
			`,
		}

		dir, cleanup, err := analysistest.WriteFiles(fm)

		require.NoError(t, err)
		defer cleanup()

		analysistest.Run(t, dir, OSExitAnalyzer, "./...")
	})

	t.Run("negative test", func(t *testing.T) {
		fm := map[string]string{
			"a/main.go": `
			package main
	
			import (
				"fmt"
				"os"
			)
			
			func main() {
				os.Exit(0) // want
				fmt.Print("test") // want "error diagnostic"
			}
			`,
			"b/pkg.go": `
			package pkg1
	
			import (
				"fmt"
				"os"
			)
	
			func osExitCheckFunc() {
				os.Exit(0) // want "error diagnostic"
				fmt.Print("test") // want "error diagnostic"
			}
			`,
		}

		dir, cleanup, err := analysistest.WriteFiles(fm)

		require.NoError(t, err)
		defer cleanup()

		// a fake *testing.T https://github.com/golang/tools/blob/master/go/analysis/analysistest/analysistest_test.go#L131
		var got []string
		t2 := errorfunc(func(s string) { got = append(got, s) })
		analysistest.Run(t2, dir, OSExitAnalyzer, "./...")

		want := []string{
			"a/main.go:10:5: unexpected diagnostic: use os.Exit",
			"a/main.go:11: no diagnostic was reported matching `error diagnostic`",
			"b/pkg.go:10: no diagnostic was reported matching `error diagnostic`",
			"b/pkg.go:11: no diagnostic was reported matching `error diagnostic`",
		}

		require.Equal(t, want, got)
	})
}

type errorfunc func(string)

func (f errorfunc) Errorf(format string, args ...interface{}) {
	f(fmt.Sprintf(format, args...))
}
