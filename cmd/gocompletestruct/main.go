// gocompletestruct checks whether struct literals in given
// sources are filled with all fields explicitly specified.
//
// Usage:
//   gocompletestruct <args...>
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/motemen/go-statictools/completestruct"
)

func main() {
	singlechecker.Main(completestruct.Analyzer)
}
