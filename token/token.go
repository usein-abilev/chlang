// The token package represents the different types of tokens that the lexer can produce
package token

import "fmt"

const (
	ILLEGAL            = iota
	EOF                // end of file
	EOL                // end of line
	INT_LITERAL        // 123
	COMMENT            // // or /* */
	FLOAT_LITERAL      // 123.45
	STRING_LITERAL     // "hello"
	IDENTIFIER         // variable_name
	ASSIGN             // =
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
	LEFT_SHIFT_EQUALS  // <<=
	GREATER            // >
	RIGHT_SHIFT        // >>
	RIGHT_SHIFT_EQUALS // >>=
	EQUALS             // ==
	NOT_EQUALS         // !=
	LESS_EQUALS        // <=
	GREATER_EQUALS     // >=
	PLUS_EQUALS        // +=
	MINUS_EQUALS       // -=
	ASTERISK_EQUALS    // *=
	SLASH_EQUALS       // /=
	PERCENT_EQUALS     // %=
	AMPERSAND_EQUALS   // &=
	PIPE_EQUALS        // |=
	CARET_EQUALS       // ^=
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
	Row    int
	Column int
}

func (p TokenPosition) String() string {
	return fmt.Sprintf("Ln %d, Col %d", p.Row, p.Column)
}

// Token represents a token in the source code
type Token struct {
	Pos     TokenPosition
	Type    TokenType
	Literal string
}

// tokenSymbolNames maps token types to their string representation (for debugging purposes)
var tokenSymbolNames = map[TokenType]string{
	ILLEGAL:          "<illegal>",
	EOF:              "<eof>",
	EOL:              "<eol>",
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
	PLUS_EQUALS:      "+=",
	MINUS_EQUALS:     "-=",
	ASTERISK_EQUALS:  "*=",
	SLASH_EQUALS:     "/=",
	PERCENT_EQUALS:   "%=",
	AMPERSAND_EQUALS: "&=",
	PIPE_EQUALS:      "|=",
	CARET_EQUALS:     "^=",
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
	_ = iota
	precLowest
	precEquals      // ==
	precLessGreater // > or <
	precSum         // +
	precProduct     // *
	precPrefix      // -X or !X
)

// operatorPrecedence maps operators to their precedence (used in the parser to determine the order of operations)
var operatorPrecedence = map[TokenType]int{
	PLUS:           precSum,
	MINUS:          precSum,
	ASTERISK:       precProduct,
	SLASH:          precProduct,
	PERCENT:        precProduct,
	AMPERSAND:      precProduct,
	PIPE:           precProduct,
	CARET:          precProduct,
	LEFT_SHIFT:     precProduct,
	RIGHT_SHIFT:    precProduct,
	EXPONENT:       precProduct,
	EQUALS:         precEquals,
	NOT_EQUALS:     precEquals,
	LESS:           precLessGreater,
	LESS_EQUALS:    precLessGreater,
	GREATER:        precLessGreater,
	GREATER_EQUALS: precLessGreater,
	AND:            precLessGreater,
	OR:             precLessGreater,
}

func GetOperatorPrecedence(op TokenType) int {
	if precedence, ok := operatorPrecedence[op]; ok {
		return precedence
	}
	return 0
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
	return fmt.Sprintf("Token{Type: '%s', Literal: %q, Ln %d, Col %d}", tokenSymbolNames[t.Type], t.Literal, t.Pos.Row, t.Pos.Column)
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
