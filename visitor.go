package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strconv"
)

type visitor struct {
	githubComPkgErrorsLocalName string
	githubComPkgErrorsNeeded    bool
	stdErrorsNeeded             bool
	fmtNeeded                   bool
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	if v.githubComPkgErrorsLocalName == "" {
		panic("v.githubComPkgErrorsLocalName is empty, should be like \"errors\"")
	}
	switch n := node.(type) {
	case *ast.CallExpr:
		switch f := n.Fun.(type) {
		case *ast.SelectorExpr:
			switch x := f.X.(type) {
			case *ast.Ident:
				switch x.Name {
				case v.githubComPkgErrorsLocalName:
					switch f.Sel.Name {
					case "Wrap", "Wrapf":
						processWrap(n, f.Sel.Name)
						v.fmtNeeded = true
					case "Errorf":
						processErrorf(n, f.Sel.Name)
						v.fmtNeeded = true
					case "New", "As", "Is":
						v.stdErrorsNeeded = true
					default:
						fmt.Fprintf(os.Stderr, "WARNING: unsupported function ``%s.%s`, you'll have to modify the source manually.\n",
							x.Name, f.Sel.Name)
						v.githubComPkgErrorsNeeded = true
					}
				}
			}
		}
	}
	if v.githubComPkgErrorsNeeded && v.stdErrorsNeeded {
		v.stdErrorsNeeded = false
	}
	return v
}

// processWrap processes errors.Wrap and errors.Wrapf
func processWrap(callExpr *ast.CallExpr, funSelName string) {
	switch funSelName {
	case "Wrap", "Wrapf":
	default:
		panic(fmt.Errorf("expected funSelName to be \"Wrap\" or \"Wrapf\", got%q", funSelName))
	}
	errExpr := callExpr.Args[0]
	sExpr := callExpr.Args[1]
	argExprs := callExpr.Args[2:]

	callExpr.Fun = &ast.SelectorExpr{
		X: &ast.Ident{
			Name: "fmt",
		},
		Sel: &ast.Ident{
			Name: "Errorf",
		},
	}
	switch sE := sExpr.(type) {
	case *ast.BasicLit:
		sE.Value = strconv.Quote(unquote(sE.Value) + ": %w")
		callExpr.Args = []ast.Expr{sE}
	default:
		callExpr.Args = []ast.Expr{
			&ast.BinaryExpr{
				X:  sE,
				Op: token.ADD,
				Y: &ast.BasicLit{
					Kind:  token.STRING,
					Value: strconv.Quote(": %w"),
				},
			},
		}
	}
	callExpr.Args = append(callExpr.Args, argExprs...)
	callExpr.Args = append(callExpr.Args, errExpr)
}

// processWrap processes errors.Errorf
func processErrorf(callExpr *ast.CallExpr, funSelName string) {
	switch funSelName {
	case "Errorf":
	default:
		panic(fmt.Errorf("expected funSelName to be \"Errorf\", got%q", funSelName))
	}

	callExpr.Fun = &ast.SelectorExpr{
		X: &ast.Ident{
			Name: "fmt",
		},
		Sel: &ast.Ident{
			Name: "Errorf",
		},
	}
}
