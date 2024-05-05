// staticlint contains various analyzers.
//
// # Analyzers used
//
//   - https://golang.org/x/tools/go/analysis/multichecker
//   - https://golang.org/x/tools/go/analysis/passes/appends
//   - https://golang.org/x/tools/go/analysis/passes/asmdecl
//   - https://golang.org/x/tools/go/analysis/passes/assign
//   - https://golang.org/x/tools/go/analysis/passes/atomic
//   - https://golang.org/x/tools/go/analysis/passes/atomicalign
//   - https://golang.org/x/tools/go/analysis/passes/bools
//   - https://golang.org/x/tools/go/analysis/passes/buildssa
//   - https://golang.org/x/tools/go/analysis/passes/buildtag
//   - https://golang.org/x/tools/go/analysis/passes/cgocall
//   - https://golang.org/x/tools/go/analysis/passes/composite
//   - https://golang.org/x/tools/go/analysis/passes/copylock
//   - https://golang.org/x/tools/go/analysis/passes/ctrlflow
//   - https://golang.org/x/tools/go/analysis/passes/deepequalerrors
//   - https://golang.org/x/tools/go/analysis/passes/defers
//   - https://golang.org/x/tools/go/analysis/passes/directive
//   - https://golang.org/x/tools/go/analysis/passes/errorsas
//   - https://golang.org/x/tools/go/analysis/passes/fieldalignment
//   - https://golang.org/x/tools/go/analysis/passes/findcall
//   - https://golang.org/x/tools/go/analysis/passes/framepointer
//   - https://golang.org/x/tools/go/analysis/passes/httpmux
//   - https://golang.org/x/tools/go/analysis/passes/httpresponse
//   - https://golang.org/x/tools/go/analysis/passes/ifaceassert
//   - https://golang.org/x/tools/go/analysis/passes/inspect
//   - https://golang.org/x/tools/go/analysis/passes/loopclosure
//   - https://golang.org/x/tools/go/analysis/passes/lostcancel
//   - https://golang.org/x/tools/go/analysis/passes/nilfunc
//   - https://golang.org/x/tools/go/analysis/passes/nilness
//   - https://golang.org/x/tools/go/analysis/passes/pkgfact
//   - https://golang.org/x/tools/go/analysis/passes/printf
//   - https://golang.org/x/tools/go/analysis/passes/reflectvaluecompare
//   - https://golang.org/x/tools/go/analysis/passes/shadow
//   - https://golang.org/x/tools/go/analysis/passes/shift
//   - https://golang.org/x/tools/go/analysis/passes/sigchanyzer
//   - https://golang.org/x/tools/go/analysis/passes/slog
//   - https://golang.org/x/tools/go/analysis/passes/sortslice
//   - https://golang.org/x/tools/go/analysis/passes/stdmethods
//   - https://golang.org/x/tools/go/analysis/passes/stdversion
//   - https://golang.org/x/tools/go/analysis/passes/stringintconv
//   - https://golang.org/x/tools/go/analysis/passes/structtag
//   - https://golang.org/x/tools/go/analysis/passes/testinggoroutine
//   - https://golang.org/x/tools/go/analysis/passes/tests
//   - https://golang.org/x/tools/go/analysis/passes/timeformat
//   - https://golang.org/x/tools/go/analysis/passes/unmarshal
//   - https://golang.org/x/tools/go/analysis/passes/unreachable
//   - https://golang.org/x/tools/go/analysis/passes/unsafeptr
//   - https://golang.org/x/tools/go/analysis/passes/unusedresult
//   - https://golang.org/x/tools/go/analysis/passes/unusedwrite
//   - https://golang.org/x/tools/go/analysis/passes/usesgenerics
//   - https://honnef.co/go/tools/staticcheck - SA and S1 analyzers
//   - https://github.com/kisielk/errcheck/errcheck
//   - https://github.com/securego/gosec/v2/analyzers
//   - https://github.com/DarkOmap/metricsService/internal/osexitcheck - check for call os.Exit
//
// # Usage
//
//		$ go install github.com/DarkOmap/metricsService/cmd/staticlint
//	 	$ staticlint help
//		$ staticlint ./...
package main
