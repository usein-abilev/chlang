package ast

import (
	"fmt"
)

func PrintAST(program *Program) {
	program.PrintTree(0)
}

func (p Program) PrintTree(level int) {
	printIndent(level)
	fmt.Println("Program")
	for _, statement := range p.Statements {
		statement.PrintTree(level + 1)
	}
}

func (p BadStatement) PrintTree(level int) {
	printIndent(level)
	fmt.Println("BadStatement: (!)")
}

func (p BadExpression) PrintTree(level int) {
	printIndent(level)
	fmt.Println("BadExpression: (!)")
}

func (p Identifier) PrintTree(level int) {
	printIndent(level)
	fmt.Printf("Identifier: %s\n", p.Value)
}

func (p AssignExpression) PrintTree(level int) {
	printIndent(level)
	fmt.Printf("AssignExp:\n")
	p.Left.PrintTree(level + 1)
	if p.Right != nil {
		printIndent(level + 1)
		fmt.Println("Expr:")
		p.Right.PrintTree(level + 2)
	}
}

func (r ReturnStatement) PrintTree(level int) {
	printIndent(level)
	fmt.Println("ReturnStatement")
	if r.Expression != nil {
		printIndent(level + 1)
		fmt.Println("Expr:")
		r.Expression.PrintTree(level + 2)
	}
}

func (b BinaryExpression) PrintTree(level int) {
	printIndent(level)
	fmt.Printf("BinaryExp: %s\n", b.Operator.Literal)
	if b.Left != nil {
		printIndent(level + 1)
		fmt.Println("Left:")
		b.Left.PrintTree(level + 2)
	}
	if b.Right != nil {
		printIndent(level + 1)
		fmt.Println("Right:")
		b.Right.PrintTree(level + 2)
	}
}

func (ie IntLiteral) PrintTree(level int) {
	printIndent(level)
	fmt.Printf("IntLiteral: %s\n", ie.Value)
}
func (ie BoolLiteral) PrintTree(level int) {
	printIndent(level)
	fmt.Printf("BoolLiteral: %s\n", ie.Value)
}
func (ie StringLiteral) PrintTree(level int) {
	printIndent(level)
	fmt.Printf("StringLiteral: %s\n", ie.Value)
}
func (ie FloatLiteral) PrintTree(level int) {
	printIndent(level)
	fmt.Printf("IntLiteral: %s\n", ie.Value)
}

func (ie UnaryExpression) PrintTree(level int) {
	printIndent(level)
	fmt.Printf("UnaryExp: %s\n", ie.Operator.Literal)
	if ie.Right != nil {
		printIndent(level + 1)
		fmt.Println("Expr:")
		ie.Right.PrintTree(level + 2)
	}
}

func (ie CallExpression) PrintTree(level int) {
	printIndent(level)
	fmt.Printf("CallExpression: %s\n", ie.Function.Value)
	if len(ie.Args) > 0 {
		printIndent(level + 1)
		fmt.Println("Arguments:")
		for _, arg := range ie.Args {
			arg.PrintTree(level + 2)
		}
	}
}

func (ie IfExpression) PrintTree(level int) {
	printIndent(level)
	fmt.Println("IfExpression")

	if ie.Condition != nil {
		printIndent(level + 1)
		fmt.Println("Condition:")
		ie.Condition.PrintTree(level + 2)
	}

	if ie.ThenBlock != nil {
		printIndent(level + 1)
		fmt.Println("ThenBlock:")
		ie.ThenBlock.PrintTree(level + 2)
	}

	if len(ie.ElseBlock) > 0 {
		printIndent(level + 1)
		fmt.Println("ElseBlock:")
		for _, block := range ie.ElseBlock {
			block.PrintTree(level + 2)
		}
	}
}

func (p VarDeclarationStatement) PrintTree(level int) {
	printIndent(level)
	fmt.Printf("VarDeclaration: %s\n", p.Name.Value)

	if p.Type != nil {
		printIndent(level + 1)
		fmt.Printf("Type: %s\n", p.Type.Value)
	}

	if p.Value != nil {
		printIndent(level + 1)
		fmt.Println("Expr:")
		p.Value.PrintTree(level + 2)
	}
}

func (es ExpressionStatement) PrintTree(level int) {
	printIndent(level)
	fmt.Println("ExpressionStatement")
	if es.Expression != nil {
		es.Expression.PrintTree(level + 1)
	}
}

func (fds FuncDeclarationStatement) PrintTree(level int) {
	printIndent(level)
	fmt.Printf("FuncDeclaration: %s\n", fds.Name.Value)

	if len(fds.Params) > 0 {
		printIndent(level + 1)
		fmt.Println("Params:")
		for _, param := range fds.Params {
			param.PrintTree(level + 2)
		}
	}

	if fds.Body != nil {
		printIndent(level + 1)
		fmt.Println("Body:")
		fds.Body.PrintTree(level + 2)
	}

	if fds.ReturnType != nil {
		printIndent(level + 1)
		fmt.Printf("ReturnType: %s\n", fds.ReturnType.Value)
	}
}

func (fa FuncArgument) PrintTree(level int) {
	printIndent(level)
	fmt.Printf("FuncArgument: %s, Type: %s\n", fa.Name.Value, fa.Type.Value)
}

func (bl BlockStatement) PrintTree(level int) {
	printIndent(level)
	fmt.Println("BlockStatement")
	for _, statement := range bl.Statements {
		statement.PrintTree(level + 1)
	}
}

func printIndent(level int) {
	for i := 0; i < level; i++ {
		fmt.Print("|  ")
	}
}
