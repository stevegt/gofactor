package main

import (
	"go/ast"

	// . "github.com/stevegt/goadapt"
	"golang.org/x/tools/go/ast/astutil"
)

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