// AST Builder implementation
// TODO: Implement:
// * [ ] Constant declaration
// * [ ] Function declaration
// * [ ] Function calls
// * [ ] Function arguments spread operator
// * [ ] If-else statements
// * [ ] For loops
// * [ ] Range types
// * [ ] Match/When statements
// * [ ] Structs
// * [ ] Struct methods
// * [ ] Enums
// * [ ]  Arrays
package parser

import (
	"log"

	"github.com/usein-abilev/chlang/scanner"
	tokenpkg "github.com/usein-abilev/chlang/token"
)

type span struct {
	Start tokenpkg.TokenPosition
	End   tokenpkg.TokenPosition
}
type (
	Statement interface {
		Node()
		PrintTree(level int)
	}
	Expression interface {
		Node()
		PrintTree(level int)
	}

	Identifier struct {
		Span  span
		Token *tokenpkg.Token
		Value string
	}

	IntLiteral struct {
		Span  span
		Value string
	}
	FloatLiteral struct {
		Span  span
		Value string
	}

	UnaryExp struct {
		Span  span
		Op    *tokenpkg.Token
		Right Expression
	}
	BinaryExp struct {
		Span  span
		Op    *tokenpkg.Token
		Left  Expression
		Right Expression
	}
	AssignExp struct {
		Span  span
		Op    *tokenpkg.Token
		Left  Expression
		Right Expression
	}
	IfExpression struct {
		Span      span
		Condition Expression
		ThenBlock *BlockStatement
		ElseBlock []*BlockStatement
	}
	ExpressionStatement struct {
		Span span
		Expr Expression
	}
	// Function declaration statement
	FuncDeclarationStatement struct {
		Span       span
		FunToken   *tokenpkg.Token
		Name       *Identifier
		Params     []*FuncArgument
		Body       *BlockStatement
		ReturnType *Identifier
	}
	FuncArgument struct {
		Name *Identifier
		Type *Identifier
	}
	BlockStatement struct {
		Span       span
		Statements []Statement
	}
	ReturnStatement struct {
		Span       span
		Expression Expression
	}
	// Variable declaration statement such as `let a = 1`
	VarDeclarationStatement struct {
		Span     span
		LetToken *tokenpkg.Token
		Name     *Identifier
		Type     *Identifier
		Value    Expression
	}
)

func (Identifier) Node()               {}
func (IntLiteral) Node()               {}
func (FloatLiteral) Node()             {}
func (UnaryExp) Node()                 {}
func (BinaryExp) Node()                {}
func (AssignExp) Node()                {}
func (BlockStatement) Node()           {}
func (IfExpression) Node()             {}
func (ExpressionStatement) Node()      {}
func (ReturnStatement) Node()          {}
func (VarDeclarationStatement) Node()  {}
func (FuncDeclarationStatement) Node() {}

// Program is the root node of the AST
type Program struct {
	Statements []Statement
}

// Parser is a struct that builds an AST from a token stream
type Parser struct {
	lexer  *scanner.Scanner
	tokens []tokenpkg.Token
	index  int

	// current token
	current *tokenpkg.Token
}

// New creates a new AST builder/parser
func New(lexer *scanner.Scanner) *Parser {
	parser := &Parser{lexer: lexer}
	parser.tokens = make([]tokenpkg.Token, 0)
	parser.tokens = append(parser.tokens, lexer.Scan())
	parser.current = &parser.tokens[0]
	parser.index = 0
	return parser
}

// TODO: The best approach will be to use a some sort of a state machine to reduce the stack memory consumption
// But for now, we are using a recursive descent parser
func (p *Parser) Parse() *Program {
	program := &Program{Statements: make([]Statement, 0)}
	for p.current.Type != tokenpkg.EOF {
		if p.current.Type == tokenpkg.ILLEGAL {
			log.Fatalf("Illegal token: %s", p.current)
			break
		} else if p.current.Type == tokenpkg.EOL {
			p.next()
			continue
		} else if p.current.Type == tokenpkg.SEMICOLON {
			// skip semicolons because they are optional
			p.next()
			continue
		}

		statement := p.parseStatement()
		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}
	}

	for _, statement := range program.Statements {
		switch s := statement.(type) {
		case *VarDeclarationStatement:
			log.Printf("VarDeclarationStatement: %+v", map[string]interface{}{
				"Name":  s.Name.Value,
				"Value": s.Value,
				"Type":  s.Type,
			})
		}
	}

	return program
}

func (p *Parser) parseStatement() Statement {
	switch p.current.Type {
	case tokenpkg.SEMICOLON:
		p.consume(tokenpkg.SEMICOLON)
		return nil
	case tokenpkg.VAR:
		return p.parseVarStatement()
	case tokenpkg.IF:
		return &ExpressionStatement{Expr: p.parseIfExpression()}
	case tokenpkg.FOR:
		log.Fatalf("For loop is not implemented yet")
	case tokenpkg.RETURN:
		p.consume(tokenpkg.RETURN)
		if p.current.Type == tokenpkg.SEMICOLON || p.current.Type == tokenpkg.RIGHT_BRACE {
			p.next()
			log.Printf("Return return nothing %s", p.current.Literal)
			return &ReturnStatement{}
		}
		expr := p.parseExpression()
		if p.current.Type == tokenpkg.SEMICOLON {
			p.consume(p.current.Type)
		}
		return &ReturnStatement{Expression: expr}
	case tokenpkg.FUNCTION:
		return p.parseFunStatement()
	case tokenpkg.IDENTIFIER:
		// assignment statement or function call or expression statement
		identifier := p.parseIdentifier()
		var expr Expression
		if tokenpkg.IsAssignment(p.current.Type) {
			p.next()
			expression := p.parseExpression()
			if p.current.Type == tokenpkg.SEMICOLON {
				p.consume(tokenpkg.SEMICOLON)
			}
			expr = AssignExp{Left: identifier, Right: expression}
		} else if p.current.Type == tokenpkg.LEFT_PAREN {
			// function call
			log.Fatalf("Function call is not implemented yet")
		} else {
			// expression statement
			expression := p.parseExpression()
			if p.current.Type == tokenpkg.SEMICOLON {
				p.consume(tokenpkg.SEMICOLON)
			}
			expr = expression
		}

		return &ExpressionStatement{Expr: expr}
	case tokenpkg.LEFT_BRACE:
		return p.parseBlockStatement()
	case tokenpkg.INT_LITERAL, tokenpkg.FLOAT_LITERAL:
		expression := p.parseExpression()
		if p.current.Type == tokenpkg.SEMICOLON {
			p.consume(tokenpkg.SEMICOLON)
		}
		return &ExpressionStatement{Expr: expression}
	default:
		log.Fatalf("Unexpected statement token: %s at %s", p.current, p.current.Pos)
	}
	return nil
}

func (p *Parser) parseIfExpression() *IfExpression {
	p.consume(tokenpkg.IF)
	condition := p.parseExpression()
	thenBlock := p.parseBlockStatement()
	elseBlock := make([]*BlockStatement, 0)

	for p.current.Type == tokenpkg.ELSE {
		p.consume(tokenpkg.ELSE)
		if p.current.Type == tokenpkg.IF {
			elseif := p.parseIfExpression()
			elseBlock = append(elseBlock, elseif.ThenBlock)
		} else {
			elseBlock = append(elseBlock, p.parseBlockStatement())
		}
	}

	return &IfExpression{
		Condition: condition,
		ThenBlock: thenBlock,
		ElseBlock: elseBlock,
	}
}

func (p *Parser) parseFunStatement() *FuncDeclarationStatement {
	funToken := p.consume(tokenpkg.FUNCTION)
	identifier := p.parseIdentifier()
	p.consume(tokenpkg.LEFT_PAREN)
	params := p.parseFnParameters()
	p.consume(tokenpkg.RIGHT_PAREN)
	// TODO: return type annotation (optional)
	var returnType *Identifier
	if p.current.Type == tokenpkg.ARROW {
		p.consume(tokenpkg.ARROW)
		p.expect(tokenpkg.IDENTIFIER)
		returnType = p.parseIdentifier()
	}
	body := p.parseBlockStatement()
	return &FuncDeclarationStatement{
		FunToken:   funToken,
		Name:       identifier,
		Params:     params,
		Body:       body,
		ReturnType: returnType,
	}
}

func (p *Parser) parseFnParameters() []*FuncArgument {
	params := make([]*FuncArgument, 0)
	for p.current.Type != tokenpkg.RIGHT_PAREN {
		identifier := p.parseIdentifier()
		p.consume(tokenpkg.COLON) // type annotation
		p.expect(tokenpkg.IDENTIFIER)
		idType := p.parseIdentifier()

		arg := &FuncArgument{
			Name: identifier,
			Type: idType,
		}
		params = append(params, arg)
		if p.current.Type == tokenpkg.COMMA {
			p.consume(tokenpkg.COMMA)
		}
	}
	return params
}

func (p *Parser) parseVarStatement() *VarDeclarationStatement {
	letToken := p.consume(tokenpkg.VAR)
	identifier := p.parseIdentifier()

	var varType *Identifier
	if p.current.Type == tokenpkg.COLON {
		p.consume(tokenpkg.COLON)
		varType = p.parseIdentifier()
	}
	p.consume(tokenpkg.ASSIGN)

	expression := p.parseExpression()

	// semicolon is optional
	if p.current.Type == tokenpkg.SEMICOLON {
		p.consume(tokenpkg.SEMICOLON)
	}

	return &VarDeclarationStatement{
		LetToken: letToken,
		Name:     identifier,
		Type:     varType,
		Value:    expression,
	}
}

func (p *Parser) parseBlockStatement() *BlockStatement {
	block := &BlockStatement{Statements: make([]Statement, 0)}
	if p.current.Type == tokenpkg.LEFT_BRACE {
		p.consume(tokenpkg.LEFT_BRACE)
		for p.current.Type != tokenpkg.RIGHT_BRACE {
			statement := p.parseStatement()
			block.Statements = append(block.Statements, statement)
		}
		p.consume(tokenpkg.RIGHT_BRACE)
	} else {
		// single statement block
		statement := p.parseStatement()
		block.Statements = append(block.Statements, statement)
	}
	return block
}

func (p *Parser) parseIdentifier() *Identifier {
	token := p.consume(tokenpkg.IDENTIFIER)
	return &Identifier{Token: token, Value: token.Literal}
}

func (p *Parser) parseExpression() Expression {
	return p.parseBinaryExpression(0)
}

// Pratt parser for binary expressions
func (p *Parser) parseBinaryExpression(min int) Expression {
	left := p.parsePrimary()
	for p.current.Type != tokenpkg.EOF && min < tokenpkg.GetOperatorPrecedence(p.current.Type) {
		op := p.consume(p.current.Type)
		var precedence int = tokenpkg.GetOperatorPrecedence(op.Type)
		if tokenpkg.IsRightAssociative(op.Type) {
			precedence -= 1
		}
		right := p.parseBinaryExpression(precedence)
		left = &BinaryExp{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parsePrimary() Expression {
	switch p.current.Type {
	case tokenpkg.IF:
		return p.parseIfExpression()
	case tokenpkg.INT_LITERAL:
		token := p.consume(tokenpkg.INT_LITERAL)
		return &IntLiteral{Value: token.Literal}
	case tokenpkg.FLOAT_LITERAL:
		token := p.consume(tokenpkg.FLOAT_LITERAL)
		return &FloatLiteral{Value: token.Literal}
	case tokenpkg.IDENTIFIER:
		// TODO: identifier or function call
		return p.parseIdentifier()
	case tokenpkg.LEFT_PAREN:
		p.consume(tokenpkg.LEFT_PAREN)
		expression := p.parseExpression()
		p.consume(tokenpkg.RIGHT_PAREN)
		return expression
	case tokenpkg.PLUS, tokenpkg.MINUS, tokenpkg.BANG:
		op := p.consume(p.current.Type)
		expression := p.parsePrimary()
		return &UnaryExp{Op: op, Right: expression}
	}
	return nil
}

func (p *Parser) expect(t tokenpkg.TokenType) {
	if p.current.Type != t {
		log.Fatalf(
			"Unexpected token, expected '%s', but got '%s' at %s",
			tokenpkg.TokenSymbolName(t),
			tokenpkg.TokenSymbolName(p.current.Type),
			p.current.Pos,
		)
	}
}

func (p *Parser) consume(t tokenpkg.TokenType) *tokenpkg.Token {
	p.expect(t)
	prev := p.current
	p.next()
	return prev
}

func (p *Parser) peek() *tokenpkg.Token {
	// preload the next token
	if p.index+1 >= len(p.tokens) {
		p.tokens = append(p.tokens, p.lexer.Scan())
	}
	return &p.tokens[p.index+1]
}

func (p *Parser) next() *tokenpkg.Token {
	p.index++
	if p.index >= len(p.tokens) {
		p.tokens = append(p.tokens, p.lexer.Scan())
	}
	p.current = &p.tokens[p.index]
	return p.current
}
