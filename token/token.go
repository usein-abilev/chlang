// The token package represents the different types of tokens that the lexer can produce
package token

const (
	ILLEGAL            = iota
	EOF                // end of file
	EOL                // end of line
	COMMENT            // // or /* */
	INT_LITERAL        // 123
	FLOAT_LITERAL      // 123.45
	STRING_LITERAL     // "hello"
	IDENTIFIER         // variable_name
	ASSIGN             // =
	PLUS               // +
	MINUS              // -
	ASTERISK           // *
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

// Token represents a token in the source code
type Token struct {
	Type    int
	Literal string
}

var TokenSymbolNames = map[int]string{
	ILLEGAL:          "<illegal>",
	EOF:              "<eof>",
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

var identTokens = map[string]int{
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

func LookupKeyword(literal string) int {
	if ident, ok := identTokens[literal]; ok {
		return ident
	}
	return IDENTIFIER
}
