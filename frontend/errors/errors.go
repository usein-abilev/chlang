// This package contains the error reporting functionality for the compiler.
package errors

import (
	"fmt"
	"io"

	"github.com/usein-abilev/chlang/frontend/token"
)

type CompilerError interface {
	Write(w io.Writer)
	Error() string
}

type SyntaxWarning struct {
	Message  string
	HelpMsg  string
	Span     token.Span
	Position token.TokenPosition
}

func (e SyntaxWarning) Error() string {
	return e.Message
}

// SyntaxError represents a syntax error in the source code.
type SyntaxError struct {
	Position  token.TokenPosition
	ErrorLine string
	Message   string
	Help      string
}

// Error returns the error message.
func (e SyntaxError) Error() string {
	return e.Message
}

// SemanticError represents a semantic error in the source code.
// For example, type mismatches, undeclared variables, etc.
type SemanticError struct {
	Message  string
	HelpMsg  string
	Span     token.Span
	Position token.TokenPosition
}

func (e SemanticError) Error() string {
	return e.Message
}

// Write writes the error message to the given writer.
// It includes the error message, the position in the source code,
// the line where the error occurred, and a help message.
func (e SyntaxError) Write(w io.Writer) {
	fmt.Fprintf(w, "\033[31merror:\033[0m \033[34m%s\n", e.Message)
	filename := e.Position.Filename
	if filename == "" {
		filename = "source"
	}
	fmt.Fprintf(w, "--> <%s>%d:%d\033[0m\n", filename, e.Position.Row, e.Position.Column)
	// int to string, because we need to now indent size
	row := fmt.Sprintf("%d", e.Position.Row)
	rowSize := len(row)
	indent := " "
	for i := 0; i < rowSize; i++ {
		indent += " "
	}

	fmt.Fprintf(w, "%s |\n", indent)
	fmt.Fprintf(w, " %s | %s\n", row, e.ErrorLine)
	fmt.Fprintf(w, "%s |", indent)

	// add help message
	if e.Help != "" {
		for i := 0; i < e.Position.Column-1; i++ {
			fmt.Fprintf(w, " ")
		}
		fmt.Fprintf(w, "\033[31m^ %s\033[0m\n", e.Help)
	}
	fmt.Fprintf(w, "%s |\n", indent)
	fmt.Fprintf(w, "\n")
}

func (e SemanticError) Write(w io.Writer) {
	fmt.Fprintf(w, "\033[31merror:\033[0m \033[34m%s\n", e.Message)
	filename := e.Position.Filename
	if filename == "" {
		filename = "source"
	}
	fmt.Fprintf(w, "--> <%s>%d:%d\033[0m\n", filename, e.Position.Row, e.Position.Column)
	// fmt.Fprintf(w, "%s\n", e.Span.String())
	if e.HelpMsg != "" {
		fmt.Fprintf(w, "\033[31m%s\033[0m\n", e.HelpMsg)
	}
	fmt.Fprintf(w, "\n")
}
