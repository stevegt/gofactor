package main

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	// "github.com/sergi/go-diff/diffmatchpatch"

	. "github.com/stevegt/goadapt"
)

// test RefactorAccess
func TestRefactorAccess(t *testing.T) {

	// create io.Reader
	r, err := os.Open("testdata/src.go")
	Ck(err)
	// create io.Writer
	w := new(strings.Builder)

	// transformFieldAccesses transforms field accesses in the AST node to
	// getter and setter calls.  It uses astutil.Apply to traverse the AST
	// and replace field accesses with getter and setter calls.  It properly
	// handles comments by transferring them to the new nodes.  It uses the
	// FieldMap to determine the getter and setter names for each field.
	// It properly handles nested field accesses such as `a.b.c`,
	// converting them to `GetA().GetB().GetC()`.

	// RefactorAccess calls transformFieldAccesses
	err = RefactorAccess(r, w, "Field", "GetField", "SetField")
	Tassert(t, err == nil, "RefactorAccess failed: %v", err)

	expectBuf, err := ioutil.ReadFile("testdata/expect.go-nofmt")
	Ck(err)
	expect := string(expectBuf)
	got := w.String()

	// diff to compare expected and actual
	diff := cmp.Diff(expect, got)
	if diff != "" {
		Pf("\n=== expect:\n%v", expect)
		Pf("\n=== got:\n%v", got)
		Pl("\n=== diff:\n")
		t.Errorf("RefactorAccess() mismatch (-want +got):\n%s", diff)
	}

	/*
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(expect, got, false)
		if len(diffs) > 1 {
			t.Errorf("RefactorAccess() mismatch:\n%v", dmp.DiffPrettyText(diffs))
		}
	*/

}
