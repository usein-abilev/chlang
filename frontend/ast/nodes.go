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
		Span  token.Span
		Value string
		Base  int // integer base: 16, 10, 8, 2
	}
	FloatLiteral struct {
		Span  token.Span
		Value string
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
		ReturnType *Identifier
		Symbol     *symbols.SymbolEntity
	}
	FuncArgument struct {
		Name *Identifier
		Type *Identifier
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
	ConstDeclarationStatement struct {
		Span       token.Span
		ConstToken *token.Token
		Name       *Identifier
		Type       *Identifier
		Value      Expression
		Symbol     *symbols.SymbolEntity
	}
	VarDeclarationStatement struct {
		Span     token.Span
		LetToken *token.Token
		Name     *Identifier
		Type     *Identifier
		Value    Expression
		Symbol   *symbols.SymbolEntity
	}
)

func (Range) Node()                     {}
func (Identifier) Node()                {}
func (IntLiteral) Node()                {}
func (BoolLiteral) Node()               {}
func (FloatLiteral) Node()              {}
func (StringLiteral) Node()             {}
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
func (VarDeclarationStatement) Node()   {}
func (ConstDeclarationStatement) Node() {}
func (FuncDeclarationStatement) Node()  {}

// BadExpression are used to represent a syntax error w/o halting the parser
func (BadExpression) Node() {}

// BadStatement are used to represent a syntax error w/o halting the parser
func (BadStatement) Node() {}

func IsConstantASTNode(node Node) bool {
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
