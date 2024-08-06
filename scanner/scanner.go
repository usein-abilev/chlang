// The scanner package is responsible for scanning the source code and producing tokens
package scanner

import (
	"fmt"
	"log"
	"math"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/usein-abilev/chlang/token"
)

type Scanner struct {
	input  string // input source code
	row    int    // row number
	column int    // column number

	char     rune // current character
	charSize int  // size of the current character
	offset   int  // offset of the current character (including current character)
}

const (
	endOfFile = -1
)

func New(input string) (*Scanner, error) {
	if !utf8.ValidString(input) {
		return nil, fmt.Errorf("invalid UTF-8 string")
	}

	scanner := &Scanner{
		input:  input,
		offset: 0,
		row:    1,
		column: 1,
	}

	scanner.next()

	return scanner, nil
}

// Scan scans the source code and returns the next token
func (s *Scanner) Scan() token.Token {
	if s.offset >= len(s.input) {
		return s.createEOF()
	}

	// skip whitespace: spaces, tabs, etc
	eol := false
	for isWhitespace(s.char) {
		if s.char == '\n' {
			eol = true
		}
		if s.next() == endOfFile {
			return s.createEOF()
		}
	}
	if eol {
		return token.Token{Type: token.EOL, Literal: "\n"}
	}

	// skip comments
	if s.char == '/' && (s.peek() == '/' || s.peek() == '*') {
		// skip comments
		nextChar := s.next()
		if nextChar == '/' {
			// single line comment
			for s.char != '\n' {
				if s.next() == endOfFile {
					break
				}
			}
		} else if nextChar == '*' {
			// multi-line comment
			for {
				if s.char == '*' && s.peek() == '/' {
					s.next()
					s.next()
					break
				}
				if s.next() == endOfFile {
					break
				}
			}
		}
		return s.Scan()
	}

	if isIdentStart(s.char) {
		return s.scanIdentifier()
	}

	if unicode.IsDigit(s.char) || (s.char == '.' && unicode.IsDigit(s.peek())) {
		return s.scanNumber()
	}

	if s.char == '"' || s.char == '\'' {
		return s.scanString()
	}

	switch s.char {
	case '.': // ., ...
		s.next()
		if s.char == '.' || s.peek() == '.' {
			s.next()
			s.next()
			return s.produceToken(token.ELLIPSIS, "...")
		}
		return s.produceToken(token.DOT, ".")
	case '=':
		s.next()
		switch s.char {
		case '=':
			s.next()
			return s.produceToken(token.EQUALS, "==")
		default:
			return s.produceToken(token.ASSIGN, "=")
		}
	case ':':
		s.next()
		return s.produceToken(token.COLON, ":")
	case ';':
		s.next()
		return s.produceToken(token.SEMICOLON, ";")
	case ',':
		s.next()
		return s.produceToken(token.COMMA, ",")
	case '(':
		s.next()
		return s.produceToken(token.LEFT_PAREN, "(")
	case ')':
		s.next()
		return s.produceToken(token.RIGHT_PAREN, ")")
	case '{':
		s.next()
		return s.produceToken(token.LEFT_BRACE, "{")
	case '}':
		s.next()
		return s.produceToken(token.RIGHT_BRACE, "}")
	case '[':
		s.next()
		return s.produceToken(token.LEFT_BRACKET, "[")
	case ']':
		s.next()
		return s.produceToken(token.RIGHT_BRACKET, "]")

	// Binary and logical operators
	case '&': // &, &&, &=
		s.next()
		switch s.char {
		case '&':
			s.next()
			return s.produceToken(token.AND, "&&")
		case '=':
			s.next()
			return s.produceToken(token.AMPERSAND_EQUALS, "&=")
		default:
			return s.produceToken(token.AMPERSAND, "&")
		}
	case '|': // |, ||, |=
		s.next()
		switch s.char {
		case '|':
			s.next()
			return s.produceToken(token.OR, "||")
		case '=':
			s.next()
			return s.produceToken(token.PIPE_EQUALS, "|=")
		default:
			return s.produceToken(token.PIPE, "|")
		}
	case '^': // ^, ^=
		s.next()
		switch s.char {
		case '=':
			s.next()
			return s.produceToken(token.CARET_EQUALS, "^=")
		default:
			return s.produceToken(token.CARET, "^")
		}
	case '<': // <, <=, <<, <<=
		s.next()
		switch s.char {
		case '=':
			s.next()
			return s.produceToken(token.LESS_EQUALS, "<=")
		case '<':
			s.next()
			if s.char == '=' {
				s.next()
				return s.produceToken(token.LEFT_SHIFT_EQUALS, "<<=")
			}
			return s.produceToken(token.LEFT_SHIFT, "<<")
		default:
			return s.produceToken(token.LESS, "<")
		}
	case '>': // >, >=, >>, >>=
		s.next()
		switch s.char {
		case '=':
			s.next()
			return s.produceToken(token.GREATER_EQUALS, ">=")
		case '>':
			s.next()
			if s.char == '=' {
				s.next()
				return s.produceToken(token.RIGHT_SHIFT_EQUALS, ">>=")
			}
			return s.produceToken(token.RIGHT_SHIFT, ">>")
		default:
			return s.produceToken(token.GREATER, ">")
		}

	case '+':
		s.next()
		switch s.char {
		case '=':
			s.next()
			return s.produceToken(token.PLUS_EQUALS, "+=")
		default:
			return s.produceToken(token.PLUS, "+")
		}
	case '-':
		s.next()
		switch s.char {
		case '=':
			s.next()
			return s.produceToken(token.MINUS_EQUALS, "-=")
		default:
			return s.produceToken(token.MINUS, "-")
		}
	case '*':
		s.next()
		switch s.char {
		case '=':
			s.next()
			return s.produceToken(token.ASTERISK_EQUALS, "*=")
		case '*':
			s.next()
			return s.produceToken(token.EXPONENT, "**")
		default:
			return s.produceToken(token.ASTERISK, "*")
		}
	case '/':
		s.next()
		if s.char == '=' {
			s.next()
			return s.produceToken(token.SLASH_EQUALS, "/=")
		}
		return s.produceToken(token.SLASH, "/")
	case '%':
		s.next()
		if s.char == '=' {
			s.next()
			return s.produceToken(token.PERCENT_EQUALS, "%=")
		}
		return s.produceToken(token.PERCENT, "%")
	}

	s.fatal("Unexpected character: '%c' %x", s.char, s.char)
	return s.produceToken(token.ILLEGAL, string(s.char))
}

func (s *Scanner) scanIdentifier() token.Token {
	start := s.offset
	for isIdentPart(s.char) && s.next() != endOfFile {
	}

	literal := s.input[start:s.offset]
	return s.produceToken(token.LookupKeyword(literal), literal)
}

func (s *Scanner) scanString() token.Token {
	start := s.offset
	quote := s.char
	s.next()

	for s.char != quote {
		if s.char == endOfFile {
			s.fatal("Unterminated string")
		}
		if s.char == '\\' {
			s.next()
		}
		s.next()
	}

	s.next()
	literal := s.input[start:s.offset]
	return s.produceToken(token.STRING_LITERAL, literal)
}

func (s *Scanner) scanNumber() token.Token {
	start := s.offset

	if s.char == '0' {
		s.next() // skip zero
		letter := false
		switch s.char {
		case 'x', 'X':
			s.next()
			s.scanNumberByCond(isHexDigit)
			letter = true
		case 'b', 'B':
			s.next()
			s.scanNumberByCond(isBinaryDigit)
			letter = true
		case 'o', 'O':
			s.next()
			s.scanNumberByCond(isOctalDigit)
			letter = true
		}

		if letter {
			literal := strings.ReplaceAll(s.input[start:s.offset], "_", "")
			return s.produceToken(token.INT_LITERAL, literal)
		}
	}

	dot := false
	exp := false

	if s.char == '.' && unicode.IsDigit(s.peek()) {
		dot = true
	}

	s.scanNumberByCond(func(ch rune) bool {
		if exp {
			if ch == '+' || ch == '-' {
				return true
			}
			return unicode.IsDigit(ch)
		}

		if ch == '.' {
			if dot {
				s.fatal("Multiple dots in number")
				return false
			}
			dot = true
			return true
		}

		if ch == 'e' || ch == 'E' {
			if exp {
				s.fatal("Multiple exponents in number")
				return false
			}
			exp = true
			return true
		}

		return unicode.IsDigit(rune(ch))
	})

	literal := strings.ReplaceAll(s.input[start:s.offset], "_", "")
	if dot || exp {
		// If the number contains a dot or an exponent, it is a float number
		return s.produceToken(
			token.FLOAT_LITERAL,
			literal,
		)
	}
	return s.produceToken(
		token.INT_LITERAL,
		literal,
	)
}

func (s *Scanner) scanNumberByCond(cond func(rune) bool) string {
	for s.char == '_' || cond(s.char) {
		if s.next() == endOfFile {
			break
		}
	}

	return s.input[s.offset:s.offset]
}

func (s *Scanner) next() rune {
	s.offset += s.charSize
	if s.offset >= len(s.input) {
		return endOfFile
	}

	char, size := utf8.DecodeRuneInString(s.input[s.offset:])
	s.charSize = size
	s.column++
	s.char = char

	if char == utf8.RuneError {
		s.fatal("Invalid UTF-8 character")
		return endOfFile
	}

	if s.char == '\n' {
		s.column = 1
		s.row++
	}

	return s.char
}

func (s *Scanner) peek() rune {
	char, _ := utf8.DecodeRuneInString(s.input[s.offset+s.charSize:])
	return char
}

func (s *Scanner) createEOF() token.Token {
	return s.produceToken(token.EOF, "")
}

func (s *Scanner) produceToken(t token.TokenType, literal string) token.Token {
	return token.Token{
		Type:    t,
		Literal: literal,
		Pos: token.TokenPosition{
			Row:    s.row,
			Column: s.column,
		},
	}
}

func (s *Scanner) fatal(msg string, args ...interface{}) {
	minBound := int64(math.Max(float64(s.row-1), 0))
	maxBound := int64(math.Min(float64(s.row+1), float64(len(strings.Split(s.input, "\n")))))
	lines := strings.Join(strings.Split(s.input, "\n")[minBound:maxBound], "\n")

	log.Printf("\033[31mFailed line: (pos:%d)\n===============================\n%s\n===============================\033[0m", s.offset, lines)
	log.Fatalf(fmt.Sprintf("\033[31m[Scanner error]: %s at L%d:%d\n\033[0m", msg, s.row, s.column), args...)
}

func isIdentStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '$'
}

func isIdentPart(ch rune) bool {
	return isIdentStart(ch) ||
		unicode.IsDigit(ch) ||
		unicode.Is(unicode.Nl, ch) ||
		unicode.Is(unicode.Mn, ch) || unicode.Is(unicode.Mc, ch) ||
		unicode.Is(unicode.Me, ch) || unicode.Is(unicode.Nd, ch) // Mn, Mc, Me, Nd, Nl
}

func isHexDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isOctalDigit(ch rune) bool {
	return ch >= '0' && ch <= '7'
}

func isBinaryDigit(ch rune) bool {
	return ch == '0' || ch == '1'
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
