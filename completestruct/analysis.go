package completestruct

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/astutil"
)

// Analyzer is the entry point for completestruct
var Analyzer = &analysis.Analyzer{
	Name: "completestruct",
	Doc:  "checks if all fields are filled in struct literals",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for ident, obj := range pass.TypesInfo.Uses {
		// The identifier is a typename
		tn, ok := obj.(*types.TypeName)
		if !ok {
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

		var nodes []ast.Node
		start, end := ident.Pos(), ident.End()
		for _, f := range pass.Files {
			file := pass.Fset.File(start)
			base := file.Base()
			if base <= int(start) && int(start) < base+file.Size() {
				nodes, _ = astutil.PathEnclosingInterval(f, start, end)
				break
			}
		}

		var compLit *ast.CompositeLit
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
			err := printer.Fprint(&buf, pass.Fset, compLit.Type)
			if err != nil {
				return nil, err
			}

			pass.ReportRangef(ident, "struct fields missing in %s{} literal: %s", buf.String(), strings.Join(fields, ", "))
		}
	}

	return nil, nil
}
