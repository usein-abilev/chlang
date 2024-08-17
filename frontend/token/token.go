// The token package represents the different types of tokens that the lexer can produce
package token

import "fmt"

const (
	ILLEGAL            = iota
	EOF                // end of file
	NEW_LINE           // \n
	INT_LITERAL        // 123
	COMMENT            // // or /* */
	FLOAT_LITERAL      // 123.45
	STRING_LITERAL     // "hello"
	IDENTIFIER         // variable_name
	PLUS               // +
	MINUS              // -
	ASTERISK           // *
	EXPONENT           // **
	SLASH              // /
	PERCENT            // %
	BANG               // !
	AMPERSAND          // &
	PIPE               // |
	AND                // &&
	OR                 // ||
	CARET              // ^
	LESS               // <
	LEFT_SHIFT         // <<
	GREATER            // >
	RIGHT_SHIFT        // >>
	EQUALS             // ==
	NOT_EQUALS         // !=
	LESS_EQUALS        // <=
	GREATER_EQUALS     // >=
	ASSIGN             // =
	LEFT_SHIFT_ASSIGN  // <<=
	RIGHT_SHIFT_ASSIGN // >>=
	PLUS_ASSIGN        // +=
	MINUS_ASSIGN       // -=
	ASTERISK_ASSIGN    // *=
	EXPONENT_ASSIGN    // **=
	SLASH_ASSIGN       // /=
	PERCENT_ASSIGN     // %=
	AMPERSAND_ASSIGN   // &=
	PIPE_ASSIGN        // |=
	CARET_ASSIGN       // ^=
	ARROW              // ->
	LEFT_PAREN         // (
	RIGHT_PAREN        // )
	LEFT_BRACE         // {
	RIGHT_BRACE        // }
	LEFT_BRACKET       // [
	RIGHT_BRACKET      // ]
	COMMA              // ,
	DOT                // .
	ELLIPSIS           // ... (spread)
	COLON              // :
	SEMICOLON          // ;

	// Keywords
	VAR
	STRUCT
	CONST
	FUNCTION
	RETURN
	IF
	ELSE
	FOR
	TRUE
	FALSE
)

type TokenType int

type TokenPosition struct {
	Row      int
	Column   int
	Filename string
}

func (p TokenPosition) String() string {
	return fmt.Sprintf("Ln %d, Col %d", p.Row, p.Column)
}

// Token represents a token in the source code
type Token struct {
	Position TokenPosition
	Type     TokenType
	Literal  string
	Metadata *TokenMetadata
}

type TokenMetadata struct {
	// The number base for parsing, like: 16, 10, 8, 2
	IntegerBase int
}

type Span struct {
	Start TokenPosition
	End   TokenPosition
}

// tokenSymbolNames maps token types to their string representation (for debugging purposes)
var tokenSymbolNames = map[TokenType]string{
	ILLEGAL:          "<illegal>",
	EOF:              "<eof>",
	NEW_LINE:         "<newline>",
	COMMENT:          "<comment>",
	INT_LITERAL:      "integer",
	FLOAT_LITERAL:    "float",
	STRING_LITERAL:   "string",
	IDENTIFIER:       "identifier",
	ASSIGN:           "=",
	PLUS:             "+",
	MINUS:            "-",
	ASTERISK:         "*",
	SLASH:            "/",
	PERCENT:          "%",
	BANG:             "!",
	AMPERSAND:        "&",
	PIPE:             "|",
	AND:              "&&",
	OR:               "||",
	CARET:            "^",
	LESS:             "<",
	LESS_EQUALS:      "<=",
	GREATER:          ">",
	GREATER_EQUALS:   ">=",
	EQUALS:           "==",
	NOT_EQUALS:       "!=",
	PLUS_ASSIGN:      "+=",
	MINUS_ASSIGN:     "-=",
	ASTERISK_ASSIGN:  "*=",
	SLASH_ASSIGN:     "/=",
	PERCENT_ASSIGN:   "%=",
	AMPERSAND_ASSIGN: "&=",
	EXPONENT_ASSIGN:  "**=",
	PIPE_ASSIGN:      "|=",
	CARET_ASSIGN:     "^=",
	ARROW:            "->", // function return type arrow
	LEFT_PAREN:       "(",
	RIGHT_PAREN:      ")",
	LEFT_BRACE:       "{",
	RIGHT_BRACE:      "}",
	LEFT_BRACKET:     "[",
	RIGHT_BRACKET:    "]",
	COMMA:            ",",
	DOT:              ".",
	COLON:            ":",
	SEMICOLON:        ";",

	VAR:      "let",
	STRUCT:   "struct",
	CONST:    "const",
	FUNCTION: "fn",
	RETURN:   "return",
	IF:       "if",
	ELSE:     "else",
	FOR:      "for",
	TRUE:     "true",
	FALSE:    "false",
}

// Precedence levels for operators
const (
	_               = iota
	precAssign      // =, +=, -=, *=, /=, %=, &=, |=, ^=, <<=, >>=, **=
	precBitwise     // &, |, ^
	precLessGreater // > or <
	precSum         // +, -
	precEquals      // ==
	precProduct     // *, /, %
	precPrefix      // -X or !X
	precHighest
)

// operatorPrecedence maps operators to their precedence (used in the parser to determine the order of operations)
var operatorPrecedence = map[TokenType]int{
	ASSIGN:             precAssign,
	LEFT_SHIFT_ASSIGN:  precAssign,
	RIGHT_SHIFT_ASSIGN: precAssign,
	PLUS_ASSIGN:        precAssign,
	MINUS_ASSIGN:       precAssign,
	ASTERISK_ASSIGN:    precAssign,
	EXPONENT_ASSIGN:    precAssign,
	SLASH_ASSIGN:       precAssign,
	PERCENT_ASSIGN:     precAssign,
	AMPERSAND_ASSIGN:   precAssign,
	PIPE_ASSIGN:        precAssign,
	CARET_ASSIGN:       precAssign,
	PLUS:               precSum,
	MINUS:              precSum,
	ASTERISK:           precProduct, // multiplication
	SLASH:              precProduct, // division
	PERCENT:            precProduct, // modulo
	AMPERSAND:          precBitwise, // bitwise AND
	PIPE:               precBitwise, // bitwise OR
	CARET:              precBitwise, // bitwise XOR
	LEFT_SHIFT:         precBitwise, // bitwise left shift
	RIGHT_SHIFT:        precBitwise, // bitwise right shift
	EQUALS:             precEquals,  // ==
	NOT_EQUALS:         precEquals,  // !=
	EXPONENT:           precHighest, // exponentiation
	LESS:               precLessGreater,
	LESS_EQUALS:        precLessGreater,
	GREATER:            precLessGreater,
	GREATER_EQUALS:     precLessGreater,
	AND:                precLessGreater,
	OR:                 precLessGreater,
}

func GetOperatorPrecedence(op TokenType) int {
	if precedence, ok := operatorPrecedence[op]; ok {
		return precedence
	}
	return -1
}

func IsRightAssociative(op TokenType) bool {
	switch op {
	case EXPONENT, EXPONENT_ASSIGN, ASSIGN, PLUS_ASSIGN, MINUS_ASSIGN, ASTERISK_ASSIGN,
		SLASH_ASSIGN, PERCENT_ASSIGN, AMPERSAND_ASSIGN, PIPE_ASSIGN, CARET_ASSIGN,
		LEFT_SHIFT_ASSIGN, RIGHT_SHIFT_ASSIGN:
		return true
	}
	return false
}

func IsAssignment(op TokenType) bool {
	switch op {
	case ASSIGN, EXPONENT_ASSIGN, PLUS_ASSIGN, MINUS_ASSIGN, ASTERISK_ASSIGN, SLASH_ASSIGN,
		PERCENT_ASSIGN, AMPERSAND_ASSIGN, PIPE_ASSIGN, CARET_ASSIGN, LEFT_SHIFT_ASSIGN, RIGHT_SHIFT_ASSIGN:
		return true
	}
	return false
}

// Maps keywords to their token type (used in the lexer to determinate whether an identifier is a keyword)
var identTokens = map[string]TokenType{
	"let":    VAR,
	"const":  CONST,
	"struct": STRUCT,
	"fn":     FUNCTION,
	"return": RETURN,
	"if":     IF,
	"else":   ELSE,
	"for":    FOR,
	"true":   TRUE,
	"false":  FALSE,
}

func (t Token) String() string {
	return fmt.Sprintf("Token{Type: '%s', Literal: %q, Ln %d, Col %d}", tokenSymbolNames[t.Type], t.Literal, t.Position.Row, t.Position.Column)
}

func TokenSymbolName(t TokenType) string {
	return tokenSymbolNames[t]
}

func LookupKeyword(literal string) TokenType {
	if ident, ok := identTokens[literal]; ok {
		return ident
	}
	return IDENTIFIER
}
