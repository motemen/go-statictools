// gocompletestruct checks whether struct literals in given
// sources are filled with all fields explicitly specified.
//
// Usage:
//   gocompletestruct <args...>
package main

import (
	"log"
	"os"

	"golang.org/x/tools/go/loader"

	"github.com/motemen/go-statictools/completestruct"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix(os.Args[0] + ": ")

	conf := loader.Config{}
	_, err := conf.FromArgs(os.Args[1:], false)
	if err != nil {
		log.Fatal(err)
	}

	prog, err := conf.Load()
	if err != nil {
		log.Fatal(err)
	}

	var hadError bool
	for _, pkg := range prog.InitialPackages() {
		err := completestruct.Check(prog.Fset, pkg)
		if err != nil {
			hadError = true

			if errs, ok := err.(completestruct.ErrMulti); ok {
				for _, err := range errs {
					log.Println(err)
				}
			} else {
				log.Println(err)
			}
		}
	}
	if hadError {
		os.Exit(2)
	}
}
