package ast

import (
	"github.com/usein-abilev/chlang/frontend/ast/symbols"
	"github.com/usein-abilev/chlang/frontend/token"
)

// Base types for AST nodes
type (
	Node interface {
		Node()
		PrintTree(level int)
		GetSpan() *token.Span
	}
	Statement interface {
		Node
	}
	Expression interface {
		Node
	}
)

// Errored nodes, used to represent syntax errors
type (
	BadExpression struct {
		Span token.Span
	}
	BadStatement struct {
		Span token.Span
	}
)

type (
	Identifier struct {
		Span   token.Span
		Value  string
		Token  *token.Token
		Symbol *symbols.SymbolEntity
	}
	BoolLiteral struct {
		Span  token.Span
		Value string
	}
	IntLiteral struct {
		Span token.Span

		// Type is determined during type checking
		Type symbols.ChlangPrimitiveType

		Value  string
		Suffix string // type suffix: i8, i16, i32, i64, u8, u16, u32, u64
		Base   int    // integer base: 16, 10, 8, 2
	}
	FloatLiteral struct {
		Span token.Span

		// Type is determined during type checking
		Type symbols.ChlangPrimitiveType

		Suffix string // type suffix: f32, f64
		Value  string
	}
	StringLiteral struct {
		Span  token.Span
		Value string
	}
	Range struct {
		Span      token.Span
		Start     Expression
		End       Expression
		Inclusive bool
	}

	// Types nodes
	ArrayType struct {
		Span token.Span
		Type Expression
		Size Expression
	}
	FunctionType struct {
		Span       token.Span
		Args       []Expression
		ReturnType Expression
	}

	// Expressions
	ArrayExpression struct {
		Span     token.Span
		Elements []Expression
	}
	IndexExpression struct {
		Span  token.Span
		Left  Expression
		Index Expression
	}
	UnaryExpression struct {
		Span     token.Span
		Operator *token.Token
		Right    Expression
	}
	BinaryExpression struct {
		Span     token.Span
		Operator *token.Token
		Left     Expression
		Right    Expression
	}
	AssignExpression struct {
		Span     token.Span
		Operator *token.Token
		Left     Expression
		Right    Expression
	}
	IfExpression struct {
		Span      token.Span
		Condition Expression
		ThenBlock *BlockStatement
		ElseBlock Statement // block statement or if expression
	}
	CallExpression struct {
		Span     token.Span
		Function *Identifier
		Args     []Expression
	}
	ExpressionStatement struct {
		Span       token.Span
		Expression Expression
	}
	FuncDeclarationStatement struct {
		Span       token.Span
		Name       *Identifier
		Params     []*FuncArgument
		Body       *BlockStatement
		Symbol     *symbols.SymbolEntity
		ReturnType Expression
	}
	FuncArgument struct {
		Name *Identifier
		Type Expression
	}
	BlockStatement struct {
		Span       token.Span
		Statements []Statement
	}
	ReturnStatement struct {
		Span       token.Span
		Expression Expression
	}
	ForRangeStatement struct {
		Span       token.Span
		Identifier *Identifier
		Body       *BlockStatement
		Range      *Range
	}
	BreakStatement struct {
		Span token.Span
	}
	ContinueStatement struct {
		Span token.Span
	}
	TypeDeclarationStatement struct {
		Span token.Span
		Name *Identifier
		Spec Expression // type specification
	}
	ConstDeclarationStatement struct {
		Span       token.Span
		ConstToken *token.Token
		Name       *Identifier
		Type       Expression
		Value      Expression
		Symbol     *symbols.SymbolEntity
	}
	VarDeclarationStatement struct {
		Span     token.Span
		LetToken *token.Token
		Name     *Identifier
		Type     Expression
		Value    Expression
		Symbol   *symbols.SymbolEntity
	}
)

// type nodes
func (ArrayType) Node()    {}
func (FunctionType) Node() {}

func (Range) Node()                     {}
func (Identifier) Node()                {}
func (IntLiteral) Node()                {}
func (BoolLiteral) Node()               {}
func (FloatLiteral) Node()              {}
func (StringLiteral) Node()             {}
func (ArrayExpression) Node()           {}
func (IndexExpression) Node()           {}
func (UnaryExpression) Node()           {}
func (BinaryExpression) Node()          {}
func (AssignExpression) Node()          {}
func (CallExpression) Node()            {}
func (BlockStatement) Node()            {}
func (IfExpression) Node()              {}
func (ExpressionStatement) Node()       {}
func (ReturnStatement) Node()           {}
func (ForRangeStatement) Node()         {}
func (BreakStatement) Node()            {}
func (ContinueStatement) Node()         {}
func (TypeDeclarationStatement) Node()  {}
func (VarDeclarationStatement) Node()   {}
func (ConstDeclarationStatement) Node() {}
func (FuncDeclarationStatement) Node()  {}

func (e *ArrayType) GetSpan() *token.Span {
	return &e.Span
}
func (e *FunctionType) GetSpan() *token.Span {
	return &e.Span
}
func (e *Range) GetSpan() *token.Span {
	return &e.Span
}
func (e *Identifier) GetSpan() *token.Span {
	return &e.Span
}
func (e *IntLiteral) GetSpan() *token.Span {
	return &e.Span
}
func (e *BoolLiteral) GetSpan() *token.Span {
	return &e.Span
}
func (e *FloatLiteral) GetSpan() *token.Span {
	return &e.Span
}
func (e *StringLiteral) GetSpan() *token.Span {
	return &e.Span
}
func (e *ArrayExpression) GetSpan() *token.Span {
	return &e.Span
}
func (e *IndexExpression) GetSpan() *token.Span {
	return &e.Span
}
func (e *UnaryExpression) GetSpan() *token.Span {
	return &e.Span
}
func (e *BinaryExpression) GetSpan() *token.Span {
	return &e.Span
}
func (e *AssignExpression) GetSpan() *token.Span {
	return &e.Span
}
func (e *CallExpression) GetSpan() *token.Span {
	return &e.Span
}
func (e *BlockStatement) GetSpan() *token.Span {
	return &e.Span
}
func (e *IfExpression) GetSpan() *token.Span {
	return &e.Span
}
func (e *ExpressionStatement) GetSpan() *token.Span {
	return &e.Span
}
func (e *ReturnStatement) GetSpan() *token.Span {
	return &e.Span
}
func (e *ForRangeStatement) GetSpan() *token.Span {
	return &e.Span
}
func (e *BreakStatement) GetSpan() *token.Span {
	return &e.Span
}
func (e *ContinueStatement) GetSpan() *token.Span {
	return &e.Span
}
func (e *TypeDeclarationStatement) GetSpan() *token.Span {
	return &e.Span
}
func (e *VarDeclarationStatement) GetSpan() *token.Span {
	return &e.Span
}
func (e *ConstDeclarationStatement) GetSpan() *token.Span {
	return &e.Span
}
func (e *FuncDeclarationStatement) GetSpan() *token.Span {
	return &e.Span
}

// BadExpression are used to represent a syntax error w/o halting the parser
func (BadExpression) Node() {}
func (be *BadExpression) GetSpan() *token.Span {
	return &be.Span
}

// BadStatement are used to represent a syntax error w/o halting the parser
func (BadStatement) Node() {}
func (be *BadStatement) GetSpan() *token.Span {
	return &be.Span
}

func IsLiteralASTNode(node Node) bool {
	switch node.(type) {
	case *IntLiteral, *FloatLiteral, *BoolLiteral, *StringLiteral:
		return true
	}
	return false
}

// Program is the root node of the AST
type Program struct {
	Statements []Statement
}
