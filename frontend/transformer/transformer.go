// The AST Transformer package is responsible for transforming and optimizing the AST.
package transformer

import (
	"fmt"
	"math"
	"strconv"

	"github.com/usein-abilev/chlang/frontend/ast"
	"github.com/usein-abilev/chlang/frontend/token"
)

func Transform(program *ast.Program) *ast.Program {
	fmt.Printf("Transforming AST\n")
	for idx, statement := range program.Statements {
		program.Statements[idx] = evaluateConstant(statement)
	}
	return program
}

func evaluateConstant(rawNode ast.Node) ast.Node {
	switch node := rawNode.(type) {
	case *ast.ReturnStatement:
		node.Expression = evaluateConstant(node.Expression)
		return node
	case *ast.BlockStatement:
		for i, statement := range node.Statements {
			node.Statements[i] = evaluateConstant(statement)
		}
		return node
	case *ast.VarDeclarationStatement:
		node.Value = evaluateConstant(node.Value)
		return node
	case *ast.BinaryExpression:
		left := evaluateConstant(node.Left)
		right := evaluateConstant(node.Right)

		if ast.IsConstantASTNode(left) && ast.IsConstantASTNode(right) {
			leftNode, leftIsInt := left.(*ast.IntLiteral)
			rightNode, rightIsInt := right.(*ast.IntLiteral)

			if leftIsInt && rightIsInt {
				leftInt, _ := strconv.ParseInt(leftNode.Value, leftNode.Base, 64)
				rightInt, _ := strconv.ParseInt(rightNode.Value, rightNode.Base, 64)

				switch node.Operator.Type {
				case token.PLUS:
					return &ast.IntLiteral{
						Span:  leftNode.Span,
						Value: strconv.FormatInt(leftInt+rightInt, 10),
						Base:  10,
					}
				case token.MINUS:
					return &ast.IntLiteral{
						Span:  node.Span,
						Value: strconv.FormatInt(leftInt-rightInt, 10),
						Base:  10,
					}
				case token.ASTERISK:
					return &ast.IntLiteral{
						Span:  node.Span,
						Value: strconv.FormatInt(leftInt*rightInt, 10),
						Base:  10,
					}
				case token.SLASH:
					return &ast.IntLiteral{
						Span:  node.Span,
						Value: strconv.FormatInt(leftInt/rightInt, 10),
						Base:  10,
					}
				case token.PERCENT:
					return &ast.IntLiteral{
						Span:  node.Span,
						Value: strconv.FormatInt(leftInt%rightInt, 10),
						Base:  10,
					}
				case token.EXPONENT:
					return &ast.IntLiteral{
						Span:  node.Span,
						Value: strconv.FormatInt(int64(math.Pow(float64(leftInt), float64(rightInt))), 10),
						Base:  10,
					}
				}
			}
		}

		return node
	}

	return rawNode
}
