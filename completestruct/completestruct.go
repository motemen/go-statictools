package completestruct

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"log"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/loader"
)

type ErrFieldsMissing struct {
	Fset   *token.FileSet
	Pos    token.Pos
	Name   string
	Fields []string
}

func (err ErrFieldsMissing) Position() token.Position {
	return err.Fset.Position(err.Pos)
}

func (err ErrFieldsMissing) Error() string {
	return fmt.Sprintf("%s: struct fields missing in %s{} literal: %s", err.Position(), err.Name, strings.Join(err.Fields, ", "))
}

type ErrMulti []error

func (err ErrMulti) Error() string {
	switch len(err) {
	case 0:
		return "no errors"
	case 1:
		return err[0].Error()
	default:
		return fmt.Sprintf("%s and %d error(s)", err[0].Error(), len(err)-1)
	}
}

func (err ErrMulti) String() string {
	lines := make([]string, len(err))
	for i, err := range err {
		lines[i] = "- " + err.Error()
	}
	return strings.Join(lines, "\n")
}

func Check(fset *token.FileSet, pkg *loader.PackageInfo) error {
	targetPkgs := map[*types.Package]bool{}
	for _, im := range pkg.Files[0].Imports {
		pkgName, ok := pkg.Implicits[im].(*types.PkgName)
		if !ok {
			continue
		}

		// TODO(motemen): provide options to filter target pkgs
		targetPkgs[pkgName.Imported()] = true
	}

	var errs []error
	for ident, obj := range pkg.Uses {
		tn, ok := obj.(*types.TypeName)
		// The identifier is a typename
		if !ok {
			continue
		}
		// The typename belongs to one of target packages
		if !targetPkgs[tn.Pkg()] && tn.Pkg() != pkg.Pkg {
			continue
		}
		// The typename is for a struct type
		st, ok := tn.Type().Underlying().(*types.Struct)
		if !ok {
			continue
		}

		missingField := map[string]bool{}
		for i := 0; i < st.NumFields(); i++ {
			field := st.Field(i)
			if field.Exported() {
				missingField[field.Name()] = true
			}
		}

		var compLit *ast.CompositeLit
		nodes, _ := pathEnclosingInterval(fset, pkg, ident.Pos(), ident.End())

	nextNode:
		for _, node := range nodes {
			// find struct literal with typename of ident
			var ok bool
			compLit, ok = node.(*ast.CompositeLit)
			if !ok {
				continue
			}
			switch expr := compLit.Type.(type) {
			case *ast.Ident:
				if expr != ident {
					continue nextNode
				}
			case *ast.SelectorExpr:
				if expr.Sel != ident {
					continue nextNode
				}
			default:
				continue nextNode
			}

			for _, elt := range compLit.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				ident, ok := kv.Key.(*ast.Ident)
				if !ok {
					continue
				}
				delete(missingField, ident.Name)
			}

			break
		}
		if compLit == nil {
			continue
		}

		if len(missingField) > 0 {
			fields := make([]string, 0, len(missingField))
			for i := 0; i < st.NumFields(); i++ {
				field := st.Field(i)
				if field.Exported() {
					if missingField[field.Name()] {
						fields = append(fields, field.Name())
					}
				}
			}

			var buf bytes.Buffer
			printer.Fprint(&buf, fset, compLit.Type)

			errs = append(errs, &ErrFieldsMissing{
				Fset:   fset,
				Pos:    ident.Pos(),
				Name:   buf.String(),
				Fields: fields,
			})
		}
	}

	if len(errs) > 0 {
		return ErrMulti(errs)
	}

	return nil
}

func pathEnclosingInterval(fset *token.FileSet, pkg *loader.PackageInfo, start, end token.Pos) (path []ast.Node, exact bool) {
	for _, f := range pkg.Files {
		file := fset.File(start)
		base := file.Base()
		if int(start) < base || base+file.Size() <= int(start) {
			continue
		}
		return astutil.PathEnclosingInterval(f, start, end)
	}

	return nil, false
}

func debug(format string, args ...interface{}) {
	log.Printf("debug: "+format, args...)
}
