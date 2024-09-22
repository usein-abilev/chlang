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

func (at ArrayType) PrintTree(level int) {
	printIndent(level)
	fmt.Println("ArrayType")
	printIndent(level + 1)
	fmt.Println("Type:")
	at.Type.PrintTree(level + 2)
	if at.Size != nil {
		printIndent(level + 1)
		fmt.Println("Size:")
		at.Size.PrintTree(level + 2)
	}
}

func (ft FunctionType) PrintTree(level int) {
	printIndent(level)
	fmt.Println("FunctionType")

	if len(ft.Args) > 0 {
		printIndent(level + 1)
		fmt.Println("Args:")
		for _, arg := range ft.Args {
			arg.PrintTree(level + 2)
		}
	}

	if ft.ReturnType != nil {
		printIndent(level + 1)
		fmt.Println("ReturnType:")
		ft.ReturnType.PrintTree(level + 2)
	}
}

func (ie ArrayExpression) PrintTree(level int) {
	printIndent(level)
	fmt.Println("ArrayExpression")
	for _, element := range ie.Elements {
		element.PrintTree(level + 1)
	}
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

	if ie.ElseBlock != nil {
		printIndent(level + 1)
		fmt.Println("ElseBlock:")
		ie.ElseBlock.PrintTree(level + 2)
	}
}

func (p TypeDeclarationStatement) PrintTree(level int) {
	printIndent(level)
	fmt.Printf("TypeDeclarationStatement: %s\n", p.Name.Value)

	printIndent(level + 1)
	fmt.Println("Name:")
	p.Name.PrintTree(level + 2)

	if p.Spec != nil {
		printIndent(level + 1)
		fmt.Println("Spec:")
		p.Spec.PrintTree(level + 2)
	}
}

func (p ConstDeclarationStatement) PrintTree(level int) {
	printIndent(level)
	fmt.Printf("ConstDeclaration: %s\n", p.Name.Value)

	if p.Type != nil {
		printIndent(level + 1)
		fmt.Println("Type:")
		p.Type.PrintTree(level + 2)
	}

	if p.Value != nil {
		printIndent(level + 1)
		fmt.Println("Expr:")
		p.Value.PrintTree(level + 2)
	}
}

func (p VarDeclarationStatement) PrintTree(level int) {
	printIndent(level)
	fmt.Printf("VarDeclaration: %s\n", p.Name.Value)

	if p.Type != nil {
		printIndent(level + 1)
		fmt.Println("Type:")
		p.Type.PrintTree(level + 2)
	}

	if p.Value != nil {
		printIndent(level + 1)
		fmt.Println("Expr:")
		p.Value.PrintTree(level + 2)
	}
}

func (f ForRangeStatement) PrintTree(level int) {
	printIndent(level)
	fmt.Println("ForStatement")

	if f.Identifier != nil {
		printIndent(level + 1)
		fmt.Printf("Identifier: %s\n", f.Identifier.Value)
	}

	printIndent(level + 1)
	fmt.Print("Range ")
	if f.Range.Inclusive {
		fmt.Println("Inclusive")
	} else {
		fmt.Println("Exclusive")
	}
	f.Range.Start.PrintTree(level + 2)
	f.Range.End.PrintTree(level + 2)

	if f.Body != nil {
		printIndent(level + 1)
		fmt.Println("Body:")
		f.Body.PrintTree(level + 2)
	}
}

func (bs BreakStatement) PrintTree(level int) {
	printIndent(level)
	fmt.Println("BreakStatement")
}

func (cs ContinueStatement) PrintTree(level int) {
	printIndent(level)
	fmt.Println("ContinueStatement")
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
		fmt.Println("ReturnType:")
		fds.ReturnType.PrintTree(level + 2)
	}
}

func (fa FuncArgument) PrintTree(level int) {
	printIndent(level)
	fmt.Printf("FuncArgument: %s, Type:\n", fa.Name.Value)
	fa.Type.PrintTree(level + 1)
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
