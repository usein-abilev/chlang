// AST Builder implementation
package ast

import (
	"fmt"
	"log"
	"strings"

	compilerError "github.com/usein-abilev/chlang/frontend/errors"
	scanner "github.com/usein-abilev/chlang/frontend/scanner"
	chToken "github.com/usein-abilev/chlang/frontend/token"
)

// Parser is a struct that builds an AST from a token stream
type Parser struct {
	lexer  *scanner.Scanner
	tokens []chToken.Token
	errors []error
	index  int

	// current token
	current            *chToken.Token
	functionScopeLevel int
}

// Init creates a new AST builder/parser
func Init(lexer *scanner.Scanner) *Parser {
	parser := &Parser{lexer: lexer}
	parser.tokens = make([]chToken.Token, 0)
	parser.tokens = append(parser.tokens, lexer.Scan())
	parser.current = &parser.tokens[0]
	parser.index = 0
	return parser
}

// TODO: The best approach will be to use a some sort of a state machine to reduce the stack memory consumption
// But for now, we are using a recursive descent parser
func (p *Parser) Parse() (*Program, *[]error) {
	program := &Program{Statements: make([]Statement, 0)}
	for p.current.Type != chToken.EOF {
		if p.current.Type == chToken.ILLEGAL {
			log.Fatalf("Illegal token: %s", p.current)
			break
		}

		statement := p.parseStatement()
		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}
	}

	return program, &p.errors
}

// Parses a statement until a semicolon or a right brace or an end of line
func (p *Parser) parseStatement() Statement {
	switch p.current.Type {
	case chToken.SEMICOLON, chToken.NEW_LINE:
		p.consume(p.current.Type)
		return nil
	case chToken.VAR:
		return p.parseVarStatement()
	case chToken.CONST:
		return p.parseConstStatement()
	case chToken.IF:
		expr := p.parseIfExpression()
		if expr == nil {
			return &BadStatement{}
		}
		return &ExpressionStatement{Expression: expr, Span: expr.Span}
	case chToken.FOR:
		// parsing for-range statement: for i in 1..10 { ... }
		forToken := p.consume(chToken.FOR)

		p.expect(chToken.IDENTIFIER)
		identifier := p.parseIdentifier()

		p.consume(chToken.IN)

		start := p.parseExpression()
		p.consume(chToken.DOT_DOT)
		end := p.parseExpression()

		block := p.parseBlockStatement()

		return &ForRangeStatement{
			Span:       chToken.Span{Start: forToken.Position, End: p.current.Position},
			Identifier: identifier,
			Range:      [2]Expression{start, end},
			Body:       block,
		}
	case chToken.BREAK:
		breakToken := p.consume(chToken.BREAK)
		return &BreakStatement{Span: chToken.Span{
			Start: breakToken.Position,
			End:   p.current.Position,
		}}
	case chToken.CONTINUE:
		continueToken := p.consume(chToken.CONTINUE)
		return &ContinueStatement{Span: chToken.Span{
			Start: continueToken.Position,
			End:   p.current.Position,
		}}
	case chToken.RETURN:
		p.consume(chToken.RETURN)
		if p.functionScopeLevel == 0 {
			p.reportError(&compilerError.SyntaxError{
				Position:  p.current.Position,
				ErrorLine: p.lexer.GetLineByPosition(p.current.Position),
				Message:   "return statement outside of function",
				Help:      "return statement can only be used inside a block statement: functions, if statements, loops etc.",
			})
			p.nextStatement()
			return &BadStatement{}
		}
		spanStart := p.current.Position
		if p.current.Type == chToken.SEMICOLON || p.current.Type == chToken.NEW_LINE {
			return &ReturnStatement{Span: chToken.Span{Start: spanStart, End: p.current.Position}}
		}
		expr := p.parseExpression()
		if p.current.Type == chToken.SEMICOLON {
			p.consume(p.current.Type)
		}
		return &ReturnStatement{Expression: expr, Span: chToken.Span{Start: spanStart, End: p.current.Position}}
	case chToken.FUNCTION:
		return p.parseFunStatement()
	case chToken.LEFT_BRACE:
		return p.parseBlockStatement()
	default:
		expr := p.parseExpression()
		ok := p.expectOneOf(
			chToken.SEMICOLON,
			chToken.NEW_LINE,
			chToken.RIGHT_BRACE,
			chToken.EOF,
		)
		if !ok {
			p.nextStatement()
			return &BadStatement{}
		}
		return &ExpressionStatement{Expression: expr}
	}
	return nil
}

func (p *Parser) parseIfExpression() *IfExpression {
	startPos := p.current.Position
	p.consume(chToken.IF)
	condition := p.parseExpression()
	thenBlock := p.parseBlockStatement()
	var elseBlock Statement
	if p.current.Type == chToken.ELSE {
		p.consume(chToken.ELSE)
		if p.current.Type == chToken.IF {
			elseBlock = p.parseIfExpression()
		} else {
			elseBlock = p.parseBlockStatement()
		}
	}

	return &IfExpression{
		Condition: condition,
		ThenBlock: thenBlock,
		ElseBlock: elseBlock,
		Span: chToken.Span{
			Start: startPos,
			End:   p.current.Position,
		},
	}
}

func (p *Parser) parseFunStatement() *FuncDeclarationStatement {
	funToken := p.consume(chToken.FUNCTION)
	p.functionScopeLevel++
	defer func() {
		p.functionScopeLevel--
	}()
	identifier := p.parseIdentifier()
	p.consume(chToken.LEFT_PAREN)
	params := p.parseFnParameters()
	p.consume(chToken.RIGHT_PAREN)

	var returnType *Identifier
	if p.current.Type == chToken.ARROW {
		p.consume(chToken.ARROW)
		p.expect(chToken.IDENTIFIER)
		returnType = p.parseIdentifier()
	}
	body := p.parseBlockStatement()
	return &FuncDeclarationStatement{
		FunToken:   funToken,
		Name:       identifier,
		Params:     params,
		Body:       body,
		ReturnType: returnType,
		Span: chToken.Span{
			Start: funToken.Position,
			End:   p.current.Position,
		},
	}
}

func (p *Parser) parseFnParameters() []*FuncArgument {
	params := make([]*FuncArgument, 0)
	for p.current.Type != chToken.RIGHT_PAREN {
		identifier := p.parseIdentifier()
		p.consume(chToken.COLON) // type annotation
		p.expect(chToken.IDENTIFIER)
		idType := p.parseIdentifier()

		arg := &FuncArgument{
			Name: identifier,
			Type: idType,
		}
		params = append(params, arg)
		if p.current.Type == chToken.COMMA {
			p.consume(chToken.COMMA)
		}
	}
	return params
}

func (p *Parser) parseConstStatement() *ConstDeclarationStatement {
	constToken := p.consume(chToken.CONST)
	identifier := p.parseIdentifier()

	var varType *Identifier
	if p.current.Type == chToken.COLON {
		p.consume(chToken.COLON)
		varType = p.parseIdentifier()
	}

	p.consume(chToken.ASSIGN)
	expression := p.parseExpression()
	p.expectOneOf(chToken.SEMICOLON, chToken.NEW_LINE)

	return &ConstDeclarationStatement{
		ConstToken: constToken,
		Name:       identifier,
		Type:       varType,
		Value:      expression,
		Span: chToken.Span{
			Start: constToken.Position,
			End:   p.current.Position,
		},
	}
}

func (p *Parser) parseVarStatement() *VarDeclarationStatement {
	letToken := p.consume(chToken.VAR)
	identifier := p.parseIdentifier()

	var varType *Identifier
	if p.current.Type == chToken.COLON {
		p.consume(chToken.COLON)
		varType = p.parseIdentifier()
	}
	var expression Expression
	if p.current.Type == chToken.ASSIGN {
		p.consume(chToken.ASSIGN)
		expression = p.parseExpression()
	}
	p.expectOneOf(chToken.SEMICOLON, chToken.NEW_LINE, chToken.EOF)

	return &VarDeclarationStatement{
		LetToken: letToken,
		Name:     identifier,
		Type:     varType,
		Value:    expression,
		Span: chToken.Span{
			Start: letToken.Position,
			End:   p.current.Position,
		},
	}
}

func (p *Parser) parseBlockStatement() *BlockStatement {
	block := &BlockStatement{Statements: make([]Statement, 0)}
	if p.current.Type == chToken.LEFT_BRACE {
		p.consume(chToken.LEFT_BRACE)
		for p.current.Type != chToken.RIGHT_BRACE {
			statement := p.parseStatement()
			if statement != nil {
				block.Statements = append(block.Statements, statement)
			}
		}
		p.consume(chToken.RIGHT_BRACE)
	} else {
		// single statement block
		statement := p.parseStatement()
		if statement != nil {
			block.Statements = append(block.Statements, statement)
		}
	}
	return block
}

func (p *Parser) parseIdentifier() *Identifier {
	start := p.current.Position
	token := p.consume(chToken.IDENTIFIER)
	return &Identifier{Token: token, Value: token.Literal, Span: chToken.Span{
		Start: start,
		End:   p.current.Position,
	}}
}

func (p *Parser) parseCallExpression() *CallExpression {
	identifier := p.parseIdentifier()
	p.consume(chToken.LEFT_PAREN)
	args := make([]Expression, 0)
	for p.current.Type != chToken.RIGHT_PAREN {
		arg := p.parseExpression()
		p.skipWhile(chToken.NEW_LINE)
		args = append(args, arg)
		if p.current.Type == chToken.COMMA {
			p.consume(chToken.COMMA)
		}
	}
	p.consume(chToken.RIGHT_PAREN)
	return &CallExpression{
		Function: identifier,
		Args:     args,
		Span: chToken.Span{
			Start: identifier.Span.Start,
			End:   p.current.Position,
		},
	}
}

func (p *Parser) parseExpression() Expression {
	return p.parseBinaryExpression(0)
}

// Pratt parser for binary expressions
func (p *Parser) parseBinaryExpression(min int) Expression {
	spanStart := p.current.Position
	left := p.parsePrimary()

	// parse assignment expression
	if p.current.Type == chToken.ASSIGN {
		op := p.consume(chToken.ASSIGN)
		right := p.parseBinaryExpression(min)
		return &AssignExpression{Left: left, Operator: op, Right: right, Span: chToken.Span{
			Start: spanStart,
			End:   p.current.Position,
		}}
	}

	for p.current.Type != chToken.EOF && min < chToken.GetOperatorPrecedence(p.current.Type) {
		op := p.consume(p.current.Type)
		p.skipWhile(chToken.NEW_LINE)
		var precedence int = chToken.GetOperatorPrecedence(op.Type)
		if chToken.IsRightAssociative(op.Type) {
			precedence -= 1
		}
		right := p.parseBinaryExpression(precedence)
		left = &BinaryExpression{Left: left, Operator: op, Right: right, Span: chToken.Span{
			Start: spanStart,
			End:   p.current.Position,
		}}
	}
	return left
}

func (p *Parser) parsePrimary() Expression {
	p.skipWhile(chToken.NEW_LINE)
	startExprPos := p.current.Position
	switch p.current.Type {
	case chToken.TRUE, chToken.FALSE:
		token := p.consume(p.current.Type)
		return &BoolLiteral{Value: token.Literal, Span: chToken.Span{
			Start: startExprPos,
			End:   p.current.Position,
		}}
	case chToken.IF:
		return p.parseIfExpression()
	case chToken.INT_LITERAL:
		token := p.consume(chToken.INT_LITERAL)
		intBase := 10
		if token.Metadata != nil {
			intBase = token.Metadata.IntegerBase
		}
		return &IntLiteral{
			Value: token.Literal,
			Base:  intBase,
			Span: chToken.Span{
				Start: startExprPos,
				End:   p.current.Position,
			}}
	case chToken.FLOAT_LITERAL:
		token := p.consume(chToken.FLOAT_LITERAL)
		return &FloatLiteral{Value: token.Literal, Span: chToken.Span{
			Start: startExprPos,
			End:   p.current.Position,
		}}
	case chToken.STRING_LITERAL:
		token := p.consume(chToken.STRING_LITERAL)
		return &StringLiteral{Value: token.Literal, Span: chToken.Span{
			Start: startExprPos,
			End:   p.current.Position,
		}}
	case chToken.IDENTIFIER:
		if p.peek().Type == chToken.LEFT_PAREN {
			return p.parseCallExpression()
		}
		return p.parseIdentifier()
	case chToken.LEFT_PAREN:
		p.consume(chToken.LEFT_PAREN)
		expression := p.parseExpression()
		p.consume(chToken.RIGHT_PAREN)
		return expression
	case chToken.PLUS, chToken.MINUS, chToken.BANG:
		op := p.consume(p.current.Type)
		expression := p.parsePrimary()
		return &UnaryExpression{Operator: op, Right: expression, Span: chToken.Span{
			Start: startExprPos,
			End:   p.current.Position,
		}}
	}

	previous := p.prev()
	p.reportError(&compilerError.SyntaxError{
		Position:  previous.Position,
		ErrorLine: p.lexer.GetLineByPosition(previous.Position),
		Message:   fmt.Sprintf("expected expression, but got '%s'", previous.Literal),
		Help:      "expected a primary expression",
	})
	return nil
}

func (p *Parser) expect(t chToken.TokenType) bool {
	if p.current.Type != t {
		message := fmt.Sprintf("Unexpected token, expected '%s', but got '%s'", chToken.TokenSymbolName(t), chToken.TokenSymbolName(p.current.Type))
		p.reportError(&compilerError.SyntaxError{
			Position:  p.current.Position,
			ErrorLine: p.lexer.GetLineByPosition(p.current.Position),
			Message:   message,
			Help:      "",
		})
		log.Fatalf(message)
		return false
	}

	return true
}

func (p *Parser) expectOneOf(types ...chToken.TokenType) bool {
	for _, t := range types {
		if p.current.Type == t {
			return true
		}
	}

	typesString := make([]string, len(types))
	for i, t := range types {
		typesString[i] = "'" + chToken.TokenSymbolName(t) + "'"
	}

	p.reportError(&compilerError.SyntaxError{
		Position:  p.current.Position,
		ErrorLine: p.lexer.GetLineByPosition(p.current.Position),
		Message:   "unexpected token",
		Help:      fmt.Sprintf("expected one of %s", strings.Join(typesString, ", ")),
	})

	fmt.Printf("expected one of %s, but got '%s' at %s", strings.Join(typesString, ", "), p.current.Literal, p.current.Position)

	return false
}

// Reports an error in the parser and continues parsing after recovery
func (p *Parser) reportError(err error) {
	p.errors = append(p.errors, err)
	p.errorRecover()
}

func (p *Parser) errorRecover() {
	// for p.current.Type != chToken.SEMICOLON && p.current.Type != chToken.EOF {
	// 	p.next()
	// }
}

var statementStartTokens = map[chToken.TokenType]bool{
	chToken.VAR:        true,
	chToken.IF:         true,
	chToken.FOR:        true,
	chToken.RETURN:     true,
	chToken.FUNCTION:   true,
	chToken.LEFT_BRACE: true,
}

var statementEndTokens = map[chToken.TokenType]bool{
	chToken.SEMICOLON:   true,
	chToken.NEW_LINE:    true,
	chToken.EOF:         true,
	chToken.RIGHT_BRACE: true,
}

// Skip to the next statement (error recovery)
func (p *Parser) nextStatement() {
	for p.current.Type != chToken.EOF {
		if _, ok := statementEndTokens[p.current.Type]; ok {
			p.next()
			return
		}
		if _, ok := statementStartTokens[p.current.Type]; ok {
			return
		}
		p.next()
	}
}

func (p *Parser) skipWhile(t chToken.TokenType) {
	for p.current.Type == t {
		p.next()
	}
}

func (p *Parser) consume(t chToken.TokenType) *chToken.Token {
	p.expect(t)
	prev := p.current
	p.next()
	if prev.Type == chToken.NEW_LINE {
		p.skipWhile(chToken.NEW_LINE)
	}
	return prev
}

func (p *Parser) prev() *chToken.Token {
	if p.index-1 < 0 {
		return nil
	}
	return &p.tokens[p.index-1]
}

func (p *Parser) peek() *chToken.Token {
	// preload the next token
	if p.index+1 >= len(p.tokens) {
		p.tokens = append(p.tokens, p.lexer.Scan())
	}
	return &p.tokens[p.index+1]
}

func (p *Parser) next() *chToken.Token {
	p.index++
	if p.index >= len(p.tokens) {
		p.tokens = append(p.tokens, p.lexer.Scan())
	}
	p.current = &p.tokens[p.index]
	return p.current
}
