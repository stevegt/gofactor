package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"io/ioutil"
	"os"

	. "github.com/stevegt/goadapt"
)

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: gofactor <filename> <field> <getter> <setter>")
		os.Exit(1)
	}

	fn := os.Args[1]
	field := os.Args[2]
	getter := os.Args[3]
	setter := os.Args[4]

	// get an io.Reader for the file
	in, err := os.Open(fn)
	Ck(err)

	// open a temporary file for writing, getting an io.Writer
	out, err := ioutil.TempFile("", "refactor-access")
	Ck(err)

	err = RefactorAccess(in, out, field, getter, setter)
	Ck(err)

	// close the files
	err = in.Close()
	Ck(err)
	err = out.Close()
	Ck(err)

	// rename the temporary file to the original file
	err = os.Rename(out.Name(), fn)
	Ck(err)
}

// FieldMap is a map of field names to getter and setter names.
type FieldMap map[string]struct {
	getter string
	setter string
}

// Add adds a field to the FieldMap.
func (fm FieldMap) Add(field, getter, setter string) {
	fm[field] = struct{ getter, setter string }{getter, setter}
}

// Getter returns the getter name for a field.  If the field
// does not exist in the FieldMap, it returns false.
func (fm FieldMap) Getter(field string) (getter string, ok bool) {
	if v, ok := fm[field]; ok {
		return v.getter, true
	}
	return "", false
}

// Setter returns the setter name for a field.  If the field
// does not exist in the FieldMap, it returns false.
func (fm FieldMap) Setter(field string) (setter string, ok bool) {
	if v, ok := fm[field]; ok {
		return v.setter, true
	}
	return "", false
}

// RefactorAccess takes an io.Reader, field name, getter name, and
// setter name, and refactors the content to use the getter and setter
// rather than direct access to the field.  It writes the refactored
// content to the io.Writer.
//
// It functions similarly to:
// perl -pne 's/\.right = (.*)/.SetRight($1)/g'
// perl -pne 's/\.right/.Right()/g'
func RefactorAccess(rd io.Reader, wr io.Writer, field, getter, setter string) (err error) {
	// read the file
	src, err := io.ReadAll(rd)
	Ck(err)

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	Ck(err)

	// Create a map of the field name to the getter and setter names
	fieldMap := FieldMap{}
	fieldMap.Add(field, getter, setter)

	// Transform the AST
	transformed := transformFieldAccesses(node, fieldMap)

	printer.Fprint(wr, fset, transformed)

	return nil
}

// Function to transfer comments from old node to new node
func transferComments(oldNode, newNode ast.Node) {
	oldCmt := oldNode.End()
	newCmt := newNode.End()
	if oldCmt.IsValid() && newCmt.IsValid() {
		// Adjust the position of the comments
		if oldCmt != newCmt {
			// Logic to ensure comments are repositioned correctly
			// This part is abstract since positioning will depend on specifics not covered here
		}
	}
}
