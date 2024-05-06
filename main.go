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
	"golang.org/x/tools/go/ast/astutil"
)

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: refactor-access <filename> <field> <getter> <setter>")
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

// RefactorAccess, given an io.Reader, field name, getter name, and
// setter name, refactors the content to use the getter and setter
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

/*
// Visit is a method that is called for each node in the AST.
// It is used to refactor the file.
func (v *visitor) Visit(n ast.Node) ast.Visitor {
	// Switch on the type of the AST node.  The x variable is the node
	// cast to the appropriate type.
	switch x := n.(type) {
	case *ast.AssignStmt:
		// AST node is an assignment statement.
		// If the assignment is to a field, refactor it to use the setter
		if len(x.Lhs) == 1 {
			// Check if the assignment is to a field.
			ident, ok := x.Lhs[0].(*ast.SelectorExpr)
			if ok {
				// If the field is is in our field map, refactor the assignment
				if ident.Sel.Name == v.fieldMap[ident.Sel.Name] {
					// Refactor by turning the assignment into a call to the setter
					x.Lhs[0] = &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ident.X,
							Sel: &ast.Ident{Name: v.fieldMap[ident.Sel.Name]},
						},
						Args: []ast.Expr{x.Rhs[0]},
					}
				}
			}
		} else {
			// lhs is a tuple; just issue a warning for now
			fmt.Println("Warning: tuple assignment not supported")
		}
	case *ast.SelectorExpr:
		// AST node is a selector expression.  This means that a field is being accessed.
		// If the field is in our field map, refactor the access to use the getter.
		name, ok := v.fieldMap[x.Sel.Name]
		_ = name
		if ok {
			// Refactor the access by turning the selector expression into a call to the getter

			x.Sel.Name = v.fieldMap[x.Sel.Name]
		}
	}

	return v
}
*/

func transformFieldAccesses(node ast.Node, fieldMap FieldMap) ast.Node {
	return astutil.Apply(node, nil, func(cursor *astutil.Cursor) bool {
		n := cursor.Node()

		switch exp := n.(type) {
		case *ast.SelectorExpr:
			if _, ok := exp.X.(*ast.Ident); ok {
				// exp.X is an identifier rather than a complex expression
				// Check if this is part of an assignment (so we don't transform it to a getter)
				_, isAssign := cursor.Parent().(*ast.AssignStmt)
				if !isAssign {
					// is not part of an assignment
					// Check if the field is in our field map
					getterName, ok := fieldMap.Getter(exp.Sel.Name)
					if ok {
						// Construct the getter call
						getterCall := &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   exp.X,
								Sel: ast.NewIdent(getterName),
							},
						}
						// Manually adjust the comment placement
						getterCall.Lparen = exp.Sel.End()
						getterCall.Rparen = exp.Sel.End()
						// Replace node and transfer comments
						cursor.Replace(getterCall)
						transferComments(exp, getterCall)
					}
				}
			}
		case *ast.AssignStmt:
			// AST node is an assignment statement.
			if len(exp.Lhs) == 1 && len(exp.Rhs) == 1 {
				// we have exactly one LHS and RHS
				if selExpr, ok := exp.Lhs[0].(*ast.SelectorExpr); ok {
					// selExpr is a selector expression
					if _, ok := selExpr.X.(*ast.Ident); ok {
						// selExpr.X is an identifier rather than a
						// complex expression.
						// get the field name
						fieldName := selExpr.Sel.Name
						// Check if the field is in our field map
						setterName, ok := fieldMap.Setter(fieldName)
						if ok {
							// Field is in our field map -- construct the setter call
							setterCall := &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   selExpr.X,
									Sel: ast.NewIdent(setterName),
								},
								Args: []ast.Expr{exp.Rhs[0]},
							}
							exprStmt := &ast.ExprStmt{X: setterCall}
							// Replace node and transfer comments
							cursor.Replace(exprStmt)
							transferComments(exp, exprStmt)
						}
					}
				}
			}
		}
		return true
	})
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
