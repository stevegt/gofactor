package main

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

func transformFieldAccesses(node ast.Node) ast.Node {
	return astutil.Apply(node, nil, func(cursor *astutil.Cursor) bool {
		n := cursor.Node()
		// Handle SelectorExpr specifically
		if selExpr, ok := n.(*ast.SelectorExpr); ok {
			// Create the chain of getters for the selector expression
			transformed := buildGetterChain(selExpr)
			cursor.Replace(transformed)
		}
		return true
	})
}

// buildGetterChain recursively builds a chain of getter calls from a SelectorExpr
func buildGetterChain(selExpr *ast.SelectorExpr) ast.Expr {
	var current ast.Expr = selExpr.X
	// Build getters upwards
	for {
		getterName := "Get" + strings.Title(selExpr.Sel.Name)
		current = &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   current,
				Sel: ast.NewIdent(getterName),
			},
		}

		// Check if we need to ascend further
		if next, ok := selExpr.X.(*ast.SelectorExpr); ok {
			selExpr = next
		} else {
			break
		}
	}
	return current
}

func main() {
	src := `package main
    type MyStruct struct {
        A struct {
            B struct {
                C struct {
                    D int
                }
            }
        }
    }

    func main() {
        x := MyStruct{}
		// This should become x.GetA().GetB().GetC().GetD()
        _ = x.A.B.C.D 
    }`

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	transformed := transformFieldAccesses(node)

	printer.Fprint(os.Stdout, fset, transformed)
}
