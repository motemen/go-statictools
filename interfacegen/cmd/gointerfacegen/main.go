package main

import (
	"flag"
	"fmt"
	_ "go/ast"
	_ "go/parser"
	_ "go/token"
	"go/types"
	"golang.org/x/tools/go/loader"
	"log"
	"strings"
)

func main() {
	log.SetPrefix("gointerfacegen: ")
	log.SetFlags(0)

	flag.Parse()

	arg := flag.Arg(0)
	p := strings.LastIndex(arg, ".")
	pkgPath, name := arg[0:p], arg[p+1:len(arg)]

	log.Println(pkgPath, name)

	conf := loader.Config{}
	_, err := conf.FromArgs([]string{pkgPath}, false)
	if err != nil {
		log.Fatal(err)
	}

	prog, err := conf.Load()
	if err != nil {
		log.Fatal(err)
	}

	type methodSpec struct {
		Name    string
		Params  []string
		Results []string
	}

	for _, pkg := range prog.InitialPackages() {
		scope := pkg.Pkg.Scope()
		target := scope.Lookup(name)
		log.Println(target)
		for _, t := range []types.Type{target.Type(), types.NewPointer(target.Type())} {
			ms := types.NewMethodSet(t)
			specs := make([]methodSpec, ms.Len())
			for i := range specs {
				meth := ms.At(i)
				sig := meth.Obj().Type().(*types.Signature) // should not fail
				if !meth.Obj().Exported() {
					continue
				}

				var spec methodSpec
				spec.Name = meth.Obj().Name()
				spec.Params = make([]string, sig.Params().Len())
				for i := range spec.Params {
					v := sig.Params().At(i)
					spec.Params[i] = v.Name() + " " + types.TypeString(v.Type(), func(pkg *types.Package) string { return pkg.Name() })
				}
				spec.Results = make([]string, sig.Results().Len())
				for i := range spec.Results {
					v := sig.Results().At(i)
					spec.Results[i] = types.TypeString(v.Type(), func(pkg *types.Package) string { return pkg.Name() })
					if v.Name() != "" {
						spec.Results[i] = v.Name() + " " + spec.Results[i]
					}
				}
				specs[i] = spec

				var resultString string
				switch len(spec.Results) {
				case 0:
					resultString = ""
				case 1:
					resultString = " " + spec.Results[0]
				default:
					resultString = " (" + strings.Join(spec.Results, ", ") + ")"
				}
				fmt.Printf("%s(%s)%s\n", spec.Name, strings.Join(spec.Params, ", "), resultString)
			}
		}
	}
}
