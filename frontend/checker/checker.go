package checker

import (
	"fmt"
	"log"
	"strconv"

	"github.com/usein-abilev/chlang/frontend/ast"
	"github.com/usein-abilev/chlang/frontend/errors"
	"github.com/usein-abilev/chlang/frontend/token"
)

type Checker struct {
	SymbolTable *SymbolTable
	Errors      []error
	Warnings    []error
}

// Check performs semantic analysis on the AST
// It populates the symbol table and checks for type mismatches
func Check(program *ast.Program) *Checker {
	c := &Checker{
		SymbolTable: NewSymbolTable(),
	}

	// Add built-in functions to the symbol table
	c.addBuiltinFunctions()

	statements := c.sortDeclarations(program.Statements)
	program.Statements = statements

	for _, statement := range statements {
		c.visitStatement(statement)
	}

	return c
}

// Sorts structs, functions declarations to be at the top of the file
// This is done to ensure that the symbol table is populated correctly
// Functions can be used before they are declared
func (c *Checker) sortDeclarations(declarations []ast.Statement) []ast.Statement {
	first := make([]ast.Statement, 0)
	other := make([]ast.Statement, 0)
	for _, declaration := range declarations {
		if _, ok := declaration.(*ast.FuncDeclarationStatement); ok {
			first = append(first, declaration)
		} else {
			other = append(other, declaration)
		}
	}
	return append(first, other...)
}

// Adds built-in functions to the symbol table
func (c *Checker) addBuiltinFunctions() {
	c.SymbolTable.Insert(&SymbolEntity{
		Used:       true,
		Name:       "println",
		Type:       "void", // return type
		EntityType: SymbolTypeFunction,
	})
}

func (c *Checker) visitStatement(statement ast.Statement) {
	switch stmt := statement.(type) {
	case *ast.VarDeclarationStatement:
		c.visitVarDeclaration(stmt)
	case *ast.FuncDeclarationStatement:
		c.visitFuncDeclaration(stmt)
	case *ast.ExpressionStatement:
		c.visitExpression(stmt.Expression)
	default:
		log.Printf("[checker.go]: Unknown statement type: %T", stmt)
	}
}

func (c *Checker) visitVarDeclaration(stmt *ast.VarDeclarationStatement) {
	if sym := c.SymbolTable.Lookup(stmt.Name.Value); sym != nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("variable '%s' has already been declared at %s", stmt.Name.Value, stmt.Span.Start),
			Position: sym.Position,
		})
		return
	}

	var varType LangSymbolType
	if stmt.Type == nil {
		varType = c.visitExpression(stmt.Value)
	} else {
		varType = LangSymbolTypeFromStr(stmt.Type.Value)
		valueType := c.visitExpression(stmt.Value)

		if varType != valueType {
			log.Fatalf("Variable '%s' has type %s, but value has type %s", stmt.Name.Value, varType, valueType)
		}
	}

	c.SymbolTable.Insert(&SymbolEntity{
		Name:         stmt.Name.Value,
		Type:         varType.String(),
		InternalType: varType,
		EntityType:   SymbolTypeVariable,
		Position:     stmt.Span.Start,
	})
}

func (c *Checker) visitFuncDeclaration(stmt *ast.FuncDeclarationStatement) {
	sym := c.SymbolTable.Lookup(stmt.Name.Value)
	if sym != nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("function '%s' has already been declared", stmt.Name.Value),
			Position: sym.Position,
		})
		return
	}

	var returnType LangSymbolType
	if stmt.ReturnType == nil {
		returnType = LangTypeVoid
	} else {
		returnType = LangSymbolTypeFromStr(stmt.ReturnType.Value)
	}

	// if the function is entry point, is already used
	used := false
	if stmt.Name.Value == "main" {
		used = true
		if returnType != LangTypeVoid {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  "main function must return void",
				Position: stmt.Span.Start,
			})
		}
	}

	c.SymbolTable.Insert(&SymbolEntity{
		Name:         stmt.Name.Value,
		Type:         returnType.String(),
		InternalType: returnType,
		EntityType:   SymbolTypeFunction,
		Position:     stmt.Span.Start,
		Used:         used,
	})
}

func (c *Checker) visitExpression(expr ast.Expression) LangSymbolType {
	if expr == nil {
		return LangTypeInvalid
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		sym := c.SymbolTable.Lookup(e.Value)
		if sym == nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("identifier '%s' not found", e.Value),
				Position: e.Span.Start,
			})
			return LangTypeInvalid
		}
		sym.Used = true
		return sym.InternalType
	case *ast.CallExpression:
		sym := c.SymbolTable.Lookup(e.Function.Value)
		if sym == nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("function '%s' not found", e.Function.Value),
				Position: e.Span.Start,
			})
			return LangTypeInvalid
		}
		sym.Used = true
		return sym.InternalType
	case *ast.StringLiteral:
		return LangTypeString
	case *ast.IntLiteral:
		_, err := strconv.ParseInt(e.Value, 10, 32)
		if err != nil {
			_, err := strconv.ParseFloat(e.Value, 32)
			if err != nil {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  fmt.Sprintf("invalid integer literal: %s", e.Value),
					Position: e.Span.Start,
				})
				return LangTypeInvalid
			}
			return LangTypeFloat32
		}
		return LangTypeInt32
	case *ast.FloatLiteral:
		_, err := strconv.ParseFloat(e.Value, 32)
		if err != nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("invalid float literal: %s", e.Value),
				Position: e.Span.Start,
			})
			return LangTypeInvalid
		}
		return LangTypeFloat32
	case *ast.BoolLiteral:
		return LangTypeBool
	case *ast.BinaryExpression:
		leftType := c.visitExpression(e.Left)
		rightType := c.visitExpression(e.Right)
		inftype, err := c.checkTypeMismatch(leftType, rightType, e.Operator)
		if err != nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  err.Error(),
				Position: e.Span.Start,
			})
			return LangTypeInvalid
		}
		return inftype
	}

	c.Errors = append(c.Errors, &errors.SemanticError{
		Message: fmt.Sprintf("unknown expression type: %T", expr),
	})
	return LangTypeInvalid
}

func (c *Checker) checkTypeMismatch(a, b LangSymbolType, operator *token.Token) (LangSymbolType, error) {
	switch operator.Type {
	case token.PLUS, token.MINUS, token.ASTERISK, token.SLASH:
		if a == b {
			return a, nil
		}
		if !a.IsNumeric() || !b.IsNumeric() {
			return LangTypeInvalid, &errors.SemanticError{
				Message:  fmt.Sprintf("operator '%s' requires numeric operands", operator.Literal),
				Position: operator.Position,
				HelpMsg:  fmt.Sprintf("got '%s' and '%s'", a, b),
			}
		}
		return a, nil
	case token.EQUALS, token.NOT_EQUALS, token.LESS, token.LESS_EQUALS, token.GREATER, token.GREATER_EQUALS:
		if a == b {
			return LangTypeBool, nil
		}
		return LangTypeInvalid, fmt.Errorf("operator '%s' requires operands of the same type", operator.Literal)
	}
	return LangTypeInvalid, fmt.Errorf("unknown operator: %s", operator.Literal)
}
