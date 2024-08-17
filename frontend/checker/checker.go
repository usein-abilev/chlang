package checker

import (
	"fmt"
	"log"
	"strconv"

	"github.com/usein-abilev/chlang/frontend/ast"
	"github.com/usein-abilev/chlang/frontend/ast/symbols"
	"github.com/usein-abilev/chlang/frontend/errors"
	"github.com/usein-abilev/chlang/frontend/token"
)

type Checker struct {
	// Current symbol table (scope)
	SymbolTable *symbols.SymbolTable
	Errors      []error
	Warnings    []error

	// Current function information
	isFunction         bool
	functionName       string
	functionReturnType symbols.SymbolValueType
}

// Check performs semantic analysis on the AST
// It populates the symbol table and checks for type mismatches
func Check(program *ast.Program) *Checker {
	c := &Checker{
		SymbolTable: symbols.NewSymbolTable(),
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
	c.SymbolTable.Insert(&symbols.SymbolEntity{
		Used:       true,
		Name:       "println",
		Type:       "void", // return type
		EntityType: symbols.SymbolTypeFunction,
	})
}

func (c *Checker) visitStatement(statement ast.Statement) {
	switch stmt := statement.(type) {
	case *ast.VarDeclarationStatement:
		c.visitVarDeclaration(stmt)
	case *ast.FuncDeclarationStatement:
		c.visitFuncDeclaration(stmt)
	case *ast.ExpressionStatement:
		c.inferExpression(stmt.Expression)
	case *ast.BlockStatement:
		c.SymbolTable.OpenScope()
		sorted := c.sortDeclarations(stmt.Statements)
		for _, statement := range sorted {
			c.visitStatement(statement)
		}
		c.SymbolTable.CloseScope()
	case *ast.ReturnStatement:
		exprReturnType := c.inferExpression(stmt.Expression)
		if !c.isCompatibleType(c.functionReturnType, exprReturnType) {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("function '%s' returns '%s', but expression type is '%s'", c.functionName, c.functionReturnType, exprReturnType),
				Position: stmt.Span.Start,
			})
		}
	default:
		log.Printf("[checker.go]: Unknown statement type: %T", stmt)
	}
}

func (c *Checker) visitVarDeclaration(stmt *ast.VarDeclarationStatement) {
	if sym := c.SymbolTable.LookupInScope(stmt.Name.Value); sym != nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("variable '%s' has already been declared at %s", stmt.Name.Value, sym.Position),
			Position: stmt.Span.Start,
		})
		return
	}

	var varType symbols.SymbolValueType
	if stmt.Type == nil {
		varType = c.inferExpression(stmt.Value)
	} else {
		varType = symbols.GetTypeByTag(stmt.Type.Value)
		valueType := c.inferExpression(stmt.Value)
		if !c.isCompatibleType(varType, valueType) {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("variable '%s' has type '%s', but value type is '%s'", stmt.Name.Value, varType, valueType),
				Position: stmt.Span.Start,
			})
			return
		}
	}

	c.SymbolTable.Insert(&symbols.SymbolEntity{
		Name:         stmt.Name.Value,
		Type:         varType.String(),
		InternalType: varType,
		EntityType:   symbols.SymbolTypeVariable,
		Position:     stmt.Span.Start,
	})
}

func (c *Checker) visitFuncDeclaration(stmt *ast.FuncDeclarationStatement) {
	if symbols.GetTypeByTag(stmt.Name.Value) != symbols.SymbolTypeInvalid {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("cannot use type '%s' as a function name", stmt.Name.Value),
			Position: stmt.Span.Start,
		})
		return
	}

	sym := c.SymbolTable.LookupInScope(stmt.Name.Value)
	if sym != nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("function '%s' has already been declared at %s", stmt.Name.Value, sym.Position),
			Position: stmt.Span.Start,
		})
		return
	}

	var returnType symbols.SymbolValueType
	if stmt.ReturnType == nil {
		returnType = symbols.SymbolTypeVoid
	} else {
		returnType = symbols.GetTypeByTag(stmt.ReturnType.Value)
	}

	// if the function is entry point, is already used
	used := false
	if stmt.Name.Value == "main" {
		used = true
		if returnType != symbols.SymbolTypeVoid {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  "main function must return void",
				Position: stmt.Span.Start,
			})
		}
	}

	// check function arguments and visit function body
	c.SymbolTable.OpenScope()
	for _, arg := range stmt.Params {
		internalType := symbols.GetTypeByTag(arg.Type.Value)
		if internalType == symbols.SymbolTypeInvalid {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("unknown type '%s' for argument '%s'", arg.Type.Value, arg.Name.Value),
				Position: arg.Name.Span.Start,
			})
		} else if internalType == symbols.SymbolTypeVoid {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  "cannot use 'void' as a type for function argument",
				Position: arg.Type.Span.Start,
			})
		}
		c.SymbolTable.Insert(&symbols.SymbolEntity{
			EntityType:   symbols.SymbolTypeVariable,
			InternalType: internalType,
			Name:         arg.Name.Value,
			Type:         arg.Type.Value,
			Position:     arg.Name.Span.Start,
			Used:         false,
		})
	}

	c.functionName = stmt.Name.Value
	c.functionReturnType = returnType
	c.isFunction = true
	c.visitStatement(stmt.Body)
	c.isFunction = false
	c.SymbolTable.CloseScope()

	c.SymbolTable.Insert(&symbols.SymbolEntity{
		Name:         stmt.Name.Value,
		Type:         returnType.String(),
		InternalType: returnType,
		EntityType:   symbols.SymbolTypeFunction,
		Position:     stmt.Span.Start,
		Used:         used,
	})
}

// Check expression type and return its internal type
// If the expression is nil, return symbols.SymbolTypeVoid
func (c *Checker) inferExpression(expr ast.Expression) symbols.SymbolValueType {
	if expr == nil {
		return symbols.SymbolTypeVoid
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		sym := c.SymbolTable.Lookup(e.Value)
		if sym == nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("identifier '%s' not found", e.Value),
				Position: e.Span.Start,
			})
			return symbols.SymbolTypeInvalid
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
			return symbols.SymbolTypeInvalid
		}
		sym.Used = true
		return sym.InternalType
	case *ast.StringLiteral:
		return symbols.SymbolTypeString
	case *ast.IntLiteral:
		return c.inferNumberLiteralType(e)
	case *ast.FloatLiteral:
		return c.inferNumberLiteralType(e)
	case *ast.BoolLiteral:
		return symbols.SymbolTypeBool
	case *ast.BinaryExpression:
		leftType := c.inferExpression(e.Left)
		rightType := c.inferExpression(e.Right)
		inftype, err := c.checkTypeMismatch(leftType, rightType, e.Operator)
		if err != nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  err.Error(),
				Position: e.Span.Start,
			})
			return symbols.SymbolTypeInvalid
		}
		return inftype
	case *ast.IfExpression:
		condType := c.inferExpression(e.Condition)
		if condType != symbols.SymbolTypeBool {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("invalid condition type expected 'bool', but got '%s'", condType),
				HelpMsg:  "",
				Span:     e.Span,
				Position: e.Span.Start,
			})
			return symbols.SymbolTypeInvalid
		}

		thenType := c.inferIfBlockStatement(e.ThenBlock)

		for _, elseBlock := range e.ElseBlock {
			elseType := c.inferIfBlockStatement(elseBlock)
			if thenType != elseType {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  "cannot determine a single type of if expression",
					HelpMsg:  "",
					Span:     e.Span,
					Position: e.Span.Start,
				})
				return symbols.SymbolTypeInvalid
			}
		}

		c.SymbolTable.CloseScope()
		return thenType
	}

	c.Errors = append(c.Errors, &errors.SemanticError{
		Message: fmt.Sprintf("unknown expression type: %T", expr),
	})
	return symbols.SymbolTypeInvalid
}

func (c *Checker) inferNumberLiteralType(expression ast.Expression) symbols.SymbolValueType {
	switch node := expression.(type) {
	case *ast.IntLiteral:
		_, err := strconv.ParseInt(node.Value, 0, 32)
		if err != nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("value '%s' is out of integer range", node.Value),
				Position: node.Span.Start,
			})
			return symbols.SymbolTypeInvalid
		}
		return symbols.SymbolTypeInt32
	case *ast.FloatLiteral:
		_, err := strconv.ParseFloat(node.Value, 64)
		if err == nil {
			return symbols.SymbolTypeFloat32
		}

		_, err = strconv.ParseInt(node.Value, 10, 32)
		if err == nil {
			return symbols.SymbolTypeInt32
		}

		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("value '%s' is out of integer range", node.Value),
			Position: node.Span.Start,
		})
		return symbols.SymbolTypeInvalid
	}

	return symbols.SymbolTypeInvalid
}

func (c *Checker) inferIfBlockStatement(block *ast.BlockStatement) symbols.SymbolValueType {
	c.SymbolTable.OpenScope()
	returnType := symbols.SymbolTypeVoid
	sorted := c.sortDeclarations(block.Statements)
	for _, statement := range sorted {
		switch stmt := statement.(type) {
		case *ast.ExpressionStatement:
			if stmt.Expression == nil {
				continue
			}
			exprType := c.inferExpression(stmt.Expression)
			returnType = exprType
		default:
			c.visitStatement(statement)
		}
	}
	c.SymbolTable.CloseScope()
	return returnType
}

func (c *Checker) isCompatibleType(a, b symbols.SymbolValueType) bool {
	return a == b // TODO: Right now, we just strictly comparing two internal types
}

func (c *Checker) checkTypeMismatch(a, b symbols.SymbolValueType, operator *token.Token) (symbols.SymbolValueType, error) {
	switch operator.Type {
	case token.PLUS, token.MINUS, token.ASTERISK, token.SLASH:
		if a == b {
			return a, nil
		}
		if !a.IsNumeric() || !b.IsNumeric() {
			return symbols.SymbolTypeInvalid, &errors.SemanticError{
				Message:  fmt.Sprintf("operator '%s' requires numeric operands", operator.Literal),
				Position: operator.Position,
				HelpMsg:  fmt.Sprintf("got '%s' and '%s'", a, b),
			}
		}
		return a, nil
	case token.EQUALS, token.NOT_EQUALS, token.LESS, token.LESS_EQUALS, token.GREATER, token.GREATER_EQUALS:
		if a == b {
			return symbols.SymbolTypeBool, nil
		}
		return symbols.SymbolTypeInvalid, fmt.Errorf("operator '%s' requires operands of the same type", operator.Literal)
	}
	return symbols.SymbolTypeInvalid, fmt.Errorf("unknown operator: %s", operator.Literal)
}
