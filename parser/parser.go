// AST Builder implementation
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
	}
	Expression interface {
		Node()
	}
	UnaryExp struct {
		Span  span
		Op    *tokenpkg.Token
		Right Expression
	}
	BinaryExp struct {
		Span  span
		Left  Expression
		Op    *tokenpkg.Token
		Right Expression
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
	// Variable declaration statement such as `let a = 1`
	VarDeclarationStatement struct {
		Span     span
		LetToken *tokenpkg.Token
		Name     *Identifier
		Type     *Identifier
		Value    Expression
	}
)

func (Identifier) Node()              {}
func (IntLiteral) Node()              {}
func (FloatLiteral) Node()            {}
func (UnaryExp) Node()                {}
func (BinaryExp) Node()               {}
func (VarDeclarationStatement) Node() {}

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
		}
		statement := p.parseStatement()
		program.Statements = append(program.Statements, statement)
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
	case tokenpkg.VAR:
		return p.parseVarStatement()
	default:
		log.Fatalf("Unexpected token: %s at %s", p.current, p.current.Pos)
	}
	return nil
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
		if op.Type == tokenpkg.EXPONENT {
			precedence -= 1
		}
		right := p.parseBinaryExpression(precedence)
		left = &BinaryExp{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parsePrimary() Expression {
	switch p.current.Type {
	case tokenpkg.INT_LITERAL:
		token := p.consume(tokenpkg.INT_LITERAL)
		return &IntLiteral{Value: token.Literal}
	case tokenpkg.FLOAT_LITERAL:
		token := p.consume(tokenpkg.FLOAT_LITERAL)
		return &FloatLiteral{Value: token.Literal}
	case tokenpkg.IDENTIFIER:
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
	default:
		log.Fatalf("Unexpected token: %s at %s", p.current, p.current.Pos)
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
