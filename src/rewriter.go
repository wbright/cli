package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/format"
	"bytes"
	"fmt"
	"io/ioutil"
)

func main() {
	fset := token.NewFileSet()

	newFile, err := parser.ParseFile(fset, "src/cf/commands/foo.go", nil, 0)
	if err != nil {
		panic(err)
	}

	count := 0

	ast.Inspect(newFile, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.CompositeLit:
				switch s := n.Type.(type){
				case *ast.SelectorExpr:
					if (s.Sel.Name == "Application") {
						if len(n.Elts) == 0 {
							return true
						}

						count++
						name := fmt.Sprintf("appAuto_%d", count)

						rewriteStructLiteralAsIdentifierAtTopOfBlock(newFile, n, name)
						replaceStructLiteralWithIdentifier(newFile, n, name)

						return false
					}
				}
			}

			return true
		})

	src, err := gofmtFile(newFile, fset)
	if err != nil {
		println(err.Error())
	}
	println(string(src))
	ioutil.WriteFile("src/cf/commands/foo1.go", src, 0666)
}

func rewriteStructLiteralAsIdentifierAtTopOfBlock(newFile *ast.File, n *ast.CompositeLit, name string) {
	var foundStmtNode ast.Stmt
	var blocksSeen []*ast.BlockStmt
	ast.Inspect(newFile, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.AssignStmt:
				expr := node.Rhs[0]
				if expr == n {
					foundStmtNode = node
				}
				ast.Inspect(expr, func(parentNode ast.Node) bool {
					switch parentNode := parentNode.(type) {
					case *ast.KeyValueExpr:
						if parentNode.Value == n {
							foundStmtNode = node
						}
					}
					return true
				})
			case *ast.ExprStmt:
				if node.X == n {
					foundStmtNode = node
				}
			case *ast.ReturnStmt:
				for _, expr := range node.Results {
					if expr == n {
						foundStmtNode = node
					}
				}
			case *ast.BlockStmt:
				if foundStmtNode == nil {
					blocksSeen = append(blocksSeen, node)
				}
			}
			return true
		})

	var block *ast.BlockStmt
	var insertionIndex int

	for _, b := range blocksSeen {
		for i, stmt := range b.List {
			if stmt == foundStmtNode {
				block = b
				insertionIndex = i
			}
		}
	}

	lhsExpr := []ast.Expr{ast.NewIdent(name)}
	rhsExpr := []ast.Expr{&ast.CompositeLit{Type: n.Type}}

	block.List = insert(block.List, insertionIndex, ast.AssignStmt{
			Lhs: lhsExpr,
			Rhs: rhsExpr,
			Tok: token.DEFINE,
		})


	for i, elt := range n.Elts {
		keyVal := elt.(*ast.KeyValueExpr)
		fieldName := keyVal.Key.(*ast.Ident)

		selector := &ast.SelectorExpr{
			X: ast.NewIdent(name),
			Sel: ast.NewIdent(fieldName.Name),
		}
		innerLhs := []ast.Expr{selector}
		innerRhs := []ast.Expr{keyVal.Value}

		block.List = insert(block.List, i + insertionIndex + 1, ast.AssignStmt{
				Lhs: innerLhs,
				Rhs: innerRhs,
				Tok: token.ASSIGN,
			})
	}
}

func replaceStructLiteralWithIdentifier(file *ast.File, n *ast.CompositeLit, name string) {
	ast.Inspect(file, func(parentNode ast.Node) bool {
			switch parentNode := parentNode.(type){
			case *ast.KeyValueExpr:
				if parentNode.Value == n {
					parentNode.Value = ast.NewIdent(name)
				}
			case *ast.AssignStmt:
				if parentNode.Rhs[0] == n {
					parentNode.Rhs[0] = ast.NewIdent(name)
				}
			case *ast.ReturnStmt:
				for i, expr := range parentNode.Results {
					if expr == n {
						parentNode.Results[i] = ast.NewIdent(name)
					}
				}
			}
			return true
		})
}

func insert(list []ast.Stmt, i int, stmt ast.AssignStmt) []ast.Stmt {
	list = append(list, &stmt)
	copy(list[i + 1:], list[i:])
	list[i] = &stmt
	return list
}

func gofmtFile(f *ast.File, fset *token.FileSet) ([]byte, error) {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
