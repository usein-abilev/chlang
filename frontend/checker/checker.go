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

	// Current function symbol pointer. To provide type information like return type, argument types, etc.
	function *symbols.SymbolEntity
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
		Type:       symbols.SymbolTypeVoid,
		EntityType: symbols.SymbolTypeFunction,
	})
}

func (c *Checker) visitStatement(statement ast.Statement) {
	switch stmt := statement.(type) {
	case *ast.ConstDeclarationStatement:
		c.visitConstDeclaration(stmt)
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
		if c.function == nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  "unexpected 'return' statement outside the function",
				Position: stmt.Span.Start,
			})
			return
		}
		signature := c.function.Function
		if !c.isCompatibleType(signature.ReturnType, exprReturnType) {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("function '%s' returns '%s', but expression type is '%s'", c.function.Name, signature.ReturnType, exprReturnType),
				Position: stmt.Span.Start,
			})
		}
	default:
		log.Printf("[checker.go]: Unknown statement type: %T", stmt)
	}
}

func (c *Checker) visitConstDeclaration(stmt *ast.ConstDeclarationStatement) {
	if sym := c.SymbolTable.LookupInScope(stmt.Name.Value); sym != nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("constant '%s' has already been declared at %s", stmt.Name.Value, sym.Position),
			Position: stmt.Span.Start,
		})
		return
	}

	var constValueType symbols.SymbolValueType

	// double-check, because we already checking for an initial value during the parsing stage
	if stmt.Value == nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  "constants must have an initial value",
			Position: stmt.Span.Start,
		})
		return
	} else {
		constValueType = c.inferExpression(stmt.Value)
	}

	if stmt.Type != nil {
		constType := symbols.GetTypeByTag(stmt.Type.Value)
		if !c.isCompatibleType(constType, constValueType) {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message: fmt.Sprintf("constant '%s' has type '%s', but value type is '%s'", stmt.Name.Value, constType, constValueType),
			})
			return
		}
	}

	symbol := &symbols.SymbolEntity{
		Name:       stmt.Name.Value,
		Type:       constValueType,
		EntityType: symbols.SymbolTypeConstant,
		Position:   stmt.Span.Start,
	}
	c.SymbolTable.Insert(symbol)
	stmt.SymbolMetadata = symbol
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
	if stmt.Value != nil {
		varType = c.inferExpression(stmt.Value)
		if stmt.Type != nil && !c.isCompatibleType(varType, symbols.GetTypeByTag(stmt.Type.Value)) {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("variable '%s' has type '%s', but value type is '%s'", stmt.Name.Value, stmt.Type.Value, varType),
				Position: stmt.Span.Start,
			})
			return
		}
	} else if stmt.Type == nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("variable '%s' must have a type or an initial value", stmt.Name.Value),
			Position: stmt.Span.Start,
		})
		return
	} else {
		varType = symbols.GetTypeByTag(stmt.Type.Value)
	}

	symbol := &symbols.SymbolEntity{
		Name:       stmt.Name.Value,
		Type:       varType,
		EntityType: symbols.SymbolTypeVariable,
		Position:   stmt.Span.Start,
	}
	c.SymbolTable.Insert(symbol)
	stmt.SymbolMetadata = symbol
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

	funcSymbol := &symbols.SymbolEntity{
		Name:       stmt.Name.Value,
		Used:       used,
		Type:       returnType,
		EntityType: symbols.SymbolTypeFunction,
		Position:   stmt.Span.Start,
		Function: &symbols.FuncSymbolSignature{
			Args:       make([]*symbols.SymbolEntity, 0),
			ReturnType: returnType,
		},
	}
	c.SymbolTable.Insert(funcSymbol)

	// check function arguments and visit function body
	c.SymbolTable.OpenScope()
	c.function = funcSymbol

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
		argSymbol := &symbols.SymbolEntity{
			EntityType: symbols.SymbolTypeVariable,
			Type:       internalType,
			Name:       arg.Name.Value,
			Position:   arg.Name.Span.Start,
			Used:       false,
		}
		funcSymbol.Function.Args = append(funcSymbol.Function.Args, argSymbol)
		c.SymbolTable.Insert(argSymbol)
	}

	c.visitStatement(stmt.Body)
	c.SymbolTable.CloseScope()
	c.function = nil

	stmt.SymbolMetadata = funcSymbol
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
		return sym.Type
	case *ast.CallExpression:
		sym := c.SymbolTable.Lookup(e.Function.Value)
		if sym == nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("function '%s' not found", e.Function.Value),
				Position: e.Span.Start,
				Span:     e.Span,
			})
			return symbols.SymbolTypeInvalid
		}
		if sym.EntityType != symbols.SymbolTypeFunction {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("'%s' is not a function", e.Function.Value),
				Position: e.Span.Start,
			})
			return symbols.SymbolTypeInvalid
		}
		if len(sym.Function.Args) != len(e.Args) {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("function '%s' expects %d arguments, but got %d", e.Function.Value, len(sym.Function.Args), len(e.Args)),
				Position: e.Span.Start,
			})
			return symbols.SymbolTypeInvalid
		}
		for idx, argExpr := range e.Args {
			argExprType := c.inferExpression(argExpr)
			argSymbol := sym.Function.Args[idx].Type
			if !c.isCompatibleType(argSymbol, argExprType) {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  fmt.Sprintf("function '%s' expects argument '%s' to be '%s', but got '%s'", e.Function.Value, sym.Function.Args[idx].Name, argSymbol, argExprType),
					Position: e.Span.Start,
				})
			}
		}
		sym.Used = true
		return sym.Type
	case *ast.StringLiteral:
		return symbols.SymbolTypeString
	case *ast.IntLiteral:
		return c.inferNumberLiteralType(e)
	case *ast.FloatLiteral:
		return c.inferNumberLiteralType(e)
	case *ast.BoolLiteral:
		return symbols.SymbolTypeBool
	case *ast.UnaryExpression:
		rightType := c.inferExpression(e.Right)
		if rightType == symbols.SymbolTypeInvalid {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  "invalid unary operand",
				Position: e.Span.Start,
			})
			return symbols.SymbolTypeInvalid
		}
		switch e.Operator.Type {
		case token.MINUS, token.PLUS:
			if !rightType.IsNumeric() {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  fmt.Sprintf("Unary operator '%s' requires only numeric operand", e.Operator.Literal),
					Position: e.Span.Start,
				})
				return symbols.SymbolTypeInvalid
			}
			return rightType
		}
		return symbols.SymbolTypeInvalid
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
				Message:  fmt.Sprintf("value '%s' is out of integer 32-bit range", node.Value),
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
			Message:  fmt.Sprintf("value '%s' is out of integer 32-bit range", node.Value),
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
