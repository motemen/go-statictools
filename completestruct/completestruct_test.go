package completestruct

import "testing"

import (
	"go/ast"
	"go/parser"
	"golang.org/x/tools/go/loader"
	"strings"
)

func TestCheck(t *testing.T) {
	conf := loader.Config{
		ParserMode: parser.ParseComments,
	}
	_, err := conf.FromArgs([]string{"./testdata/testdata.go"}, false)
	if err != nil {
		t.Fatal(err)
	}

	prog, err := conf.Load()
	if err != nil {
		t.Fatal(err)
	}

	for _, pkg := range prog.InitialPackages() {
		commentForLine := map[int]*ast.CommentGroup{}
		for _, f := range pkg.Files {
			for _, cmt := range f.Comments {
				line := prog.Fset.Position(cmt.End()).Line
				commentForLine[line+1] = cmt
			}
		}
		err := Check(prog.Fset, pkg)
		if errs, ok := err.(ErrMulti); ok {
			for _, err := range errs {
				if errMissing, ok := err.(*ErrFieldsMissing); ok {
					cmt := commentForLine[errMissing.Position().Line]
					if cmt == nil || !strings.HasPrefix(cmt.Text(), "+test") {
						t.Errorf("unexpected errMissing: %s", errMissing)
						continue
					}

					spec := map[string]string{}
					lines := strings.Split(cmt.Text(), "\n")
					for _, line := range lines[1:] { // drop first line of "+test"
						p := strings.Index(line, "=")
						if p == -1 {
							continue
						}
						spec[line[0:p]] = line[p+1:]
					}

					if got, expected := errMissing.Name, spec["name"]; got != expected {
						t.Errorf("%s: name: got %v != %v", got, expected, errMissing.Position())
					}
					if got, expected := strings.Join(errMissing.Fields, ","), spec["fields"]; got != expected {
						t.Errorf("%s: fields: got %v != %v", got, expected, errMissing.Position())
					}
				} else {
					t.Fatalf("unexpected error: %s", err)
				}
			}
		} else {
			t.Fatalf("unexpected error: %s", err)
		}
	}
}

func TestErrFieldsMissing_Position(t *testing.T) {
}

func TestErrFieldsMissing_Error(t *testing.T) {
}
