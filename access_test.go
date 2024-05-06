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

	// refactor the source code
	err = RefactorAccess(r, w, "Field", "GetField", "SetField")
	Tassert(t, err == nil, "RefactorAccess failed: %v", err)

	expectBuf, err := ioutil.ReadFile("testdata/expect.go-nofmt")
	Ck(err)
	expect := string(expectBuf)
	got := w.String()

	// diff to compare expected and actual
	diff := cmp.Diff(expect, got)
	if diff != "" {
		Pl(got)
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
