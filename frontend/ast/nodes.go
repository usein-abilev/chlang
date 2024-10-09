package ast

import (
	"github.com/usein-abilev/chlang/frontend/token"
)

// NodeSymbolRef is a pointer to a symbol in the symbol table (this information will be filled in the checker phase)
type NodeSymbolRef interface{}

// NodeLiteralType is a pointer to a type in the env (this information will be filled in the checker phase)
type NodeLiteralType interface{}

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
		Span *token.Span
	}
	BadStatement struct {
		Span *token.Span
	}
)

type (
	Identifier struct {
		Span   *token.Span
		Value  string
		Token  *token.Token
		Symbol NodeSymbolRef
	}
	BoolLiteral struct {
		Span  *token.Span
		Value string
	}
	IntLiteral struct {
		Span   *token.Span
		Value  string
		Suffix string // type suffix: i8, i16, i32, i64, u8, u16, u32, u64
		Base   int    // integer base: 16, 10, 8, 2
		Type   NodeLiteralType
	}
	FloatLiteral struct {
		Span   *token.Span
		Suffix string // type suffix: f32, f64
		Value  string
		Type   NodeLiteralType
	}
	StringLiteral struct {
		Span  *token.Span
		Value string
		Type  NodeLiteralType
	}

	// Types nodes
	ArrayType struct {
		Span *token.Span
		Type Expression
		Size Expression
	}
	StructField struct {
		Span  *token.Span
		Name  *Identifier
		Value Expression // value or type
	}
	StructType struct {
		Span   *token.Span
		Fields []*StructField
	}
	FunctionType struct {
		Span       *token.Span
		Args       []Expression
		ReturnType Expression
	}
	FuncArgument struct {
		Name *Identifier
		Type Expression
		Ref  bool // is reference type (&self or self)
	}

	// Statements
	BlockStatement struct {
		Span       *token.Span
		Statements []Statement
	}
	ReturnStatement struct {
		Span       *token.Span
		Expression Expression
	}
	ForRangeStatement struct {
		Span       *token.Span
		Identifier *Identifier
		Body       *BlockStatement
		Range      *RangeExpr
	}
	BreakStatement struct {
		Span *token.Span
	}
	ContinueStatement struct {
		Span *token.Span
	}
)

// Expressions
type (
	RangeExpr struct {
		Span      *token.Span
		Start     Expression
		End       Expression
		Inclusive bool
	}
	InitStructExpression struct {
		Name   *Identifier
		Span   *token.Span
		Fields []*StructField
	}
	ArrayExpression struct {
		Span     *token.Span
		Elements []Expression
	}
	IndexExpression struct {
		Span  *token.Span
		Left  Expression
		Index Expression
	}
	MemberExpression struct {
		Span   *token.Span
		Left   Expression
		Member *Identifier
	}
	UnaryExpression struct {
		Span     *token.Span
		Operator *token.Token
		Right    Expression
	}
	BinaryExpression struct {
		Span     *token.Span
		Operator *token.Token
		Left     Expression
		Right    Expression
	}
	AssignExpression struct {
		Span     *token.Span
		Operator *token.Token
		Left     Expression
		Right    Expression
	}
	IfExpression struct {
		Span      *token.Span
		Condition Expression
		ThenBlock *BlockStatement
		ElseBlock Statement // block statement or if expression
	}
	CallExpression struct {
		Span     *token.Span
		Function Expression // identifier or member expression
		Args     []Expression
	}
	ExpressionStatement struct {
		Span       *token.Span
		Expression Expression
	}
)

// Declarations
type (
	TypeDeclarationStatement struct {
		Span *token.Span
		Name *Identifier
		Spec Expression // type specification: StructType, FunctionType, ArrayType, Identifier
	}
	StructDeclarationStatement struct {
		Span *token.Span
		Name *Identifier
		Body *StructType
	}
	TraitDeclarationStatement struct {
		Span               *token.Span
		Name               *Identifier
		MethodSignatures   []*FunctionSignature
		MethodDeclarations []*FuncDeclarationStatement
	}
	ImplStatement struct {
		Span     *token.Span                 // span of the impl block
		Receiver *Identifier                 // type that implements the trait
		Traits   []*Identifier               // traits that are implemented
		Methods  []*FuncDeclarationStatement // methods that are implemented
	}
	FunctionSignature struct {
		Span       *token.Span
		Name       *Identifier
		SelfArg    *FuncArgument
		Args       []*FuncArgument
		ReturnType Expression
	}
	FuncDeclarationStatement struct {
		Span      *token.Span
		Signature *FunctionSignature
		Body      *BlockStatement
		Symbol    NodeSymbolRef
	}
	ConstDeclarationStatement struct {
		Span       *token.Span
		ConstToken *token.Token
		Name       *Identifier
		Type       Expression
		Value      Expression
		Symbol     NodeSymbolRef
	}
	VarDeclarationStatement struct {
		Span     *token.Span
		LetToken *token.Token
		Name     *Identifier
		Type     Expression
		Value    Expression
		Symbol   NodeSymbolRef
	}
)

// type nodes
func (ArrayType) Node()    {}
func (FunctionType) Node() {}
func (StructField) Node()  {}
func (StructType) Node()   {}

func (RangeExpr) Node()                  {}
func (Identifier) Node()                 {}
func (IntLiteral) Node()                 {}
func (BoolLiteral) Node()                {}
func (FloatLiteral) Node()               {}
func (StringLiteral) Node()              {}
func (ArrayExpression) Node()            {}
func (IndexExpression) Node()            {}
func (MemberExpression) Node()           {}
func (InitStructExpression) Node()       {}
func (UnaryExpression) Node()            {}
func (BinaryExpression) Node()           {}
func (AssignExpression) Node()           {}
func (CallExpression) Node()             {}
func (BlockStatement) Node()             {}
func (IfExpression) Node()               {}
func (ExpressionStatement) Node()        {}
func (ReturnStatement) Node()            {}
func (ForRangeStatement) Node()          {}
func (BreakStatement) Node()             {}
func (ContinueStatement) Node()          {}
func (ImplStatement) Node()              {}
func (TypeDeclarationStatement) Node()   {}
func (StructDeclarationStatement) Node() {}
func (TraitDeclarationStatement) Node()  {}
func (VarDeclarationStatement) Node()    {}
func (ConstDeclarationStatement) Node()  {}
func (FuncDeclarationStatement) Node()   {}

func (e *ArrayType) GetSpan() *token.Span {
	return e.Span
}
func (e *StructField) GetSpan() *token.Span {
	return e.Span
}
func (e *StructType) GetSpan() *token.Span {
	return e.Span
}
func (e *FunctionType) GetSpan() *token.Span {
	return e.Span
}
func (e *RangeExpr) GetSpan() *token.Span {
	return e.Span
}
func (e *Identifier) GetSpan() *token.Span {
	return e.Span
}
func (e *IntLiteral) GetSpan() *token.Span {
	return e.Span
}
func (e *BoolLiteral) GetSpan() *token.Span {
	return e.Span
}
func (e *FloatLiteral) GetSpan() *token.Span {
	return e.Span
}
func (e *StringLiteral) GetSpan() *token.Span {
	return e.Span
}
func (e *ArrayExpression) GetSpan() *token.Span {
	return e.Span
}
func (e *IndexExpression) GetSpan() *token.Span {
	return e.Span
}
func (e *MemberExpression) GetSpan() *token.Span {
	return e.Span
}
func (e *UnaryExpression) GetSpan() *token.Span {
	return e.Span
}
func (e *BinaryExpression) GetSpan() *token.Span {
	return e.Span
}
func (e *AssignExpression) GetSpan() *token.Span {
	return e.Span
}
func (e *CallExpression) GetSpan() *token.Span {
	return e.Span
}
func (e *BlockStatement) GetSpan() *token.Span {
	return e.Span
}
func (e *IfExpression) GetSpan() *token.Span {
	return e.Span
}
func (e *ExpressionStatement) GetSpan() *token.Span {
	return e.Span
}
func (e *ReturnStatement) GetSpan() *token.Span {
	return e.Span
}
func (e *ForRangeStatement) GetSpan() *token.Span {
	return e.Span
}
func (e *BreakStatement) GetSpan() *token.Span {
	return e.Span
}
func (e *ContinueStatement) GetSpan() *token.Span {
	return e.Span
}
func (e *InitStructExpression) GetSpan() *token.Span {
	return e.Span
}
func (e *TypeDeclarationStatement) GetSpan() *token.Span {
	return e.Span
}
func (e *StructDeclarationStatement) GetSpan() *token.Span {
	return e.Span
}
func (e *TraitDeclarationStatement) GetSpan() *token.Span {
	return e.Span
}
func (e *ImplStatement) GetSpan() *token.Span {
	return e.Span
}
func (e *VarDeclarationStatement) GetSpan() *token.Span {
	return e.Span
}
func (e *ConstDeclarationStatement) GetSpan() *token.Span {
	return e.Span
}
func (e *FuncDeclarationStatement) GetSpan() *token.Span {
	return e.Span
}

// BadExpression are used to represent a syntax error w/o halting the parser
func (BadExpression) Node() {}
func (be *BadExpression) GetSpan() *token.Span {
	return be.Span
}

// BadStatement are used to represent a syntax error w/o halting the parser
func (BadStatement) Node() {}
func (be *BadStatement) GetSpan() *token.Span {
	return be.Span
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
