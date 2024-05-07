package main

import (
	"go/ast"

	// . "github.com/stevegt/goadapt"
	"golang.org/x/tools/go/ast/astutil"
)

// transformFieldAccesses transforms field accesses in the AST node to
// getter and setter calls.  It uses astutil.Apply to traverse the AST
// and replace field accesses with getter and setter calls.  It properly
// handles comments by transferring them to the new nodes.  It uses the
// FieldMap to determine the getter and setter names for each field.
// It properly handles nested field accesses such as `a.b.c`,
// converting them to `GetA().GetB().GetC()`.
func transformFieldAccesses(node ast.Node, fieldMap FieldMap) ast.Node {
	return astutil.Apply(node, nil, func(cursor *astutil.Cursor) bool {
		n := cursor.Node()

		transformNode(cursor, n, fieldMap)
		return true
	})
}

func transformNode(cursor *astutil.Cursor, n ast.Node, fieldMap FieldMap) {
	switch exp := n.(type) {
	case *ast.SelectorExpr:
		if _, ok := exp.X.(*ast.Ident); ok {
			// exp.X is an identifier rather than a complex expression
			// Check if this is part of an assignment
			_, isAssign := cursor.Parent().(*ast.AssignStmt)
			if !isAssign {
				// is not part of an assignment

				// XXX replace with buildGetterChain
				var getterCall *ast.CallExpr
				// Check if the field is in our field map
				getterName, ok := fieldMap.Getter(exp.Sel.Name)
				if ok {
					// Construct the getter call
					getterCall = &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   exp.X,
							Sel: ast.NewIdent(getterName),
						},
					}
					// adjust the comment placement
					getterCall.Lparen = exp.Sel.End()
					getterCall.Rparen = exp.Sel.End()
				}

				// getterCall := buildGetterChain(exp, fieldMap)

				if getterCall != nil {
					// adjust the comment placement
					getterCall.Lparen = exp.Sel.End()
					getterCall.Rparen = exp.Sel.End()
					// Replace node
					cursor.Replace(getterCall)
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
						// Replace node
						cursor.Replace(exprStmt)
					}
				}
			}
		}
	}
}

// buildGetterChain recursively builds a chain of getter calls from a SelectorExpr
func buildGetterChain(exp *ast.SelectorExpr, fieldMap FieldMap) ast.Expr {
	var current ast.Expr = exp.X
	for {

		// Check if the field is in our field map
		getterName, ok := fieldMap.Getter(exp.Sel.Name)
		if ok {
			// Construct the getter call
			current = &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   current,
					Sel: ast.NewIdent(getterName),
				},
			}
		} else {
			// no change
			current = exp.X
		}

		if next, ok := exp.X.(*ast.SelectorExpr); ok {
			exp = next
		} else {
			break
		}
	}
	return current
}
