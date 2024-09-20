package checker

import (
	"fmt"
	"log"
	"strconv"

	"github.com/usein-abilev/chlang/frontend/ast"
	"github.com/usein-abilev/chlang/frontend/ast/symbols"
	"github.com/usein-abilev/chlang/frontend/errors"
	chToken "github.com/usein-abilev/chlang/frontend/token"
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
// It also will transform the AST into a more optimized form
func Check(program *ast.Program) *Checker {
	c := &Checker{
		SymbolTable: symbols.NewSymbolTable(),
	}

	// Add built-in functions to the symbol table
	c.addBuiltinFunctions()
	c.populateSymbolDeclarations(program.Statements)

	for _, statement := range program.Statements {
		c.visitStatement(statement)
	}

	return c
}

// Populates symbols declarations information. This method doesn't check any types conflict
// It stores only declaration information with no types information at all, to solve the calling each other problem
func (c *Checker) populateSymbolDeclarations(declarations []ast.Statement) {
	for _, statement := range declarations {
		switch decl := statement.(type) {
		case *ast.FuncDeclarationStatement:
			c.visitFuncSignature(decl)
		}
	}
}

// Adds built-in functions to the symbol table
func (c *Checker) addBuiltinFunctions() {
	c.SymbolTable.Insert(&symbols.SymbolEntity{
		Used:       true,
		Name:       "println",
		Type:       symbols.SymbolTypeVoid,
		EntityType: symbols.SymbolTypeFunction,
		Function: &symbols.FuncSymbolSignature{
			SpreadArgument: true,
			Args:           make([]*symbols.SymbolEntity, 0),
			ReturnType:     symbols.SymbolTypeVoid,
		},
	})
}

func (c *Checker) visitStatement(statement ast.Statement) {
	switch stmt := statement.(type) {
	case *ast.ConstDeclarationStatement:
		c.visitConstDeclaration(stmt)
	case *ast.VarDeclarationStatement:
		c.visitVarDeclaration(stmt)
	case *ast.FuncDeclarationStatement:
		c.visitFuncBody(stmt)
	case *ast.ExpressionStatement:
		c.inferExpression(stmt.Expression)
	case *ast.ForRangeStatement:
		c.SymbolTable.OpenScope()

		// insert range variable into the symbol table
		rangeVar := &symbols.SymbolEntity{
			Function:   nil,
			Used:       false,
			Name:       stmt.Identifier.Value,
			Type:       symbols.SymbolTypeInt64,
			EntityType: symbols.SymbolTypeVariable,
			Position:   stmt.Span.Start,
		}
		c.SymbolTable.Insert(rangeVar)
		stmt.Identifier.Symbol = rangeVar

		// We don't need a new block scope, because the for-range statement is already a block statement
		for _, statement := range stmt.Body.Statements {
			c.visitStatement(statement)
		}

		c.SymbolTable.CloseScope()
	case *ast.BlockStatement:
		c.SymbolTable.OpenScope()
		c.populateSymbolDeclarations(stmt.Statements)
		for _, statement := range stmt.Statements {
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
	stmt.Symbol = symbol
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

	if varType == symbols.SymbolTypeInvalid {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("invalid type of variable '%s'", stmt.Name.Value),
			Position: stmt.Span.Start,
		})
	}

	symbol := &symbols.SymbolEntity{
		Name:       stmt.Name.Value,
		Type:       varType,
		EntityType: symbols.SymbolTypeVariable,
		Position:   stmt.Span.Start,
	}
	c.SymbolTable.Insert(symbol)
	stmt.Symbol = symbol
}

// Checks function signature and adds it into the symbol table
func (c *Checker) visitFuncSignature(decl *ast.FuncDeclarationStatement) {
	if symbols.GetTypeByTag(decl.Name.Value) != symbols.SymbolTypeInvalid {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("cannot use type '%s' as a function name", decl.Name.Value),
			Position: decl.Span.Start,
		})
		return
	}

	if sym := c.SymbolTable.LookupInScope(decl.Name.Value); sym != nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("function '%s' has already been declared at %s", decl.Name.Value, sym.Position),
			Position: decl.Span.Start,
		})
		return
	}

	var returnType symbols.SymbolValueType
	if decl.ReturnType == nil {
		returnType = symbols.SymbolTypeVoid
	} else {
		returnType = symbols.GetTypeByTag(decl.ReturnType.Value)
	}

	if returnType == symbols.SymbolTypeInvalid {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("invalid function '%s' return type", decl.Name.Value),
			Position: decl.Span.Start,
		})
		return
	}

	// if the function is entry point, is already used
	used := false
	if decl.Name.Value == "main" {
		used = true
		if returnType != symbols.SymbolTypeVoid {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  "main function must return void",
				Position: decl.Span.Start,
			})
			return
		}
	}

	funcSymbol := &symbols.SymbolEntity{
		Name:       decl.Name.Value,
		Used:       used,
		Type:       returnType,
		EntityType: symbols.SymbolTypeFunction,
		Position:   decl.Span.Start,
		Function: &symbols.FuncSymbolSignature{
			Args:       make([]*symbols.SymbolEntity, 0),
			ReturnType: returnType,
		},
	}
	c.SymbolTable.Insert(funcSymbol)

	for _, arg := range decl.Params {
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

		// Note that we don't insert the arguments into the symbol table because we don't check the function body here except for the arguments.
		// Arguments will populates into the symbol table in 'visitFuncDeclaration' function
		funcSymbol.Function.Args = append(funcSymbol.Function.Args, argSymbol)
	}

	decl.Symbol = funcSymbol
}

// Checks function body for type matching
func (c *Checker) visitFuncBody(stmt *ast.FuncDeclarationStatement) {
	funcSymbol := c.SymbolTable.LookupInScope(stmt.Name.Value)
	if funcSymbol == nil {
		// Since function declarations populates in the 'populateSymbolDeclaration' method
		// There is no way to have 'nil' result of lookup
		panic(fmt.Sprintf("Unexpected nil as result of lookupInScope function '%s'", stmt.Name.Value))
	}

	// check function arguments and visit function body
	c.SymbolTable.OpenScope()
	c.populateSymbolDeclarations(stmt.Body.Statements)
	prevFuncPtr := c.function
	c.function = funcSymbol

	// Pushing arguments into symbol table, we didn't check types
	// because it already done in the 'populateSymbolDeclarations' method
	for _, arg := range funcSymbol.Function.Args {
		c.SymbolTable.Insert(arg)
	}

	c.visitStatement(stmt.Body)
	c.SymbolTable.CloseScope()
	c.function = prevFuncPtr

	stmt.Symbol = funcSymbol
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
		e.Symbol = sym
		return sym.Type
	case *ast.AssignExpression:
		leftType := c.inferExpression(e.Left)
		rightType := c.inferExpression(e.Right)

		if leftType == symbols.SymbolTypeInvalid || rightType == symbols.SymbolTypeInvalid {
			return symbols.SymbolTypeInvalid
		}

		if chToken.IsAssignment(e.Operator.Type) {
			if _, ok := e.Left.(*ast.Identifier); !ok {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  "left side of an assignment must be an identifier",
					Position: e.Span.Start,
				})
				return symbols.SymbolTypeInvalid
			}
			if leftType != rightType {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message: fmt.Sprintf(
						"incompatible type of an assign expression '%s' (left: %s, right: %s)",
						e.Operator.Literal,
						leftType,
						rightType,
					),
					Span:     e.Span,
					Position: e.Span.Start,
				})
				return symbols.SymbolTypeInvalid
			}
			return leftType
		}

		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("type infer: unknown assign operator type: %s", e.Operator.Literal),
			Span:     e.Span,
			Position: e.Operator.Position,
		})
		return symbols.SymbolTypeInvalid
	case *ast.CallExpression:
		sym := c.SymbolTable.Lookup(e.Function.Value)
		e.Function.Symbol = sym

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
		if !sym.Function.SpreadArgument {
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
		} else {
			for _, argExpr := range e.Args {
				t := c.inferExpression(argExpr)
				if t == symbols.SymbolTypeInvalid {
					return symbols.SymbolTypeInvalid
				}
			}
			fmt.Printf("[warn]: Type checking of the spread arguments not implemented! Ignoring type check for function '%s'.\n", sym.Name)
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
		case chToken.MINUS, chToken.PLUS:
			if !rightType.IsNumeric() {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  fmt.Sprintf("unary operator '%s' requires only numeric operand", e.Operator.Literal),
					Position: e.Span.Start,
				})
				return symbols.SymbolTypeInvalid
			}
			return rightType
		case chToken.BANG:
			if rightType != symbols.SymbolTypeBool {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  fmt.Sprintf("unary operator '%s' requires only boolean operand", e.Operator.Literal),
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
			fmt.Printf("Error checking binary expression: %s\n", e.Operator.Literal)
			e.PrintTree(2)
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

		if e.ElseBlock == nil {
			return thenType
		}

		var elseType symbols.SymbolValueType
		switch elseBlock := e.ElseBlock.(type) {
		case *ast.BlockStatement:
			elseType = c.inferIfBlockStatement(elseBlock)
		case *ast.IfExpression:
			elseType = c.inferExpression(elseBlock)
		}

		if thenType != elseType {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("cannot determine a single type of if expression (then: %s, else: %s)", thenType, elseType),
				HelpMsg:  "",
				Span:     e.Span,
				Position: e.Span.Start,
			})
			return symbols.SymbolTypeInvalid
		}

		return thenType
	}

	c.Errors = append(c.Errors, &errors.SemanticError{
		Message: fmt.Sprintf("type infer: unknown expression type: %T", expr),
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
			return symbols.SymbolTypeFloat64
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
	c.populateSymbolDeclarations(block.Statements)
	returnType := symbols.SymbolTypeVoid
	for _, statement := range block.Statements {
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

func (c *Checker) checkTypeMismatch(a, b symbols.SymbolValueType, operator *chToken.Token) (symbols.SymbolValueType, error) {
	switch operator.Type {
	case chToken.PLUS,
		chToken.MINUS,
		chToken.ASTERISK,
		chToken.EXPONENT,
		chToken.PERCENT,
		chToken.SLASH,
		chToken.AMPERSAND,
		chToken.PIPE,
		chToken.CARET,
		chToken.LEFT_SHIFT,
		chToken.RIGHT_SHIFT:
		if a == b {
			return a, nil
		}
		if !a.IsNumeric() || !b.IsNumeric() {
			return symbols.SymbolTypeInvalid, &errors.SemanticError{
				Message:  fmt.Sprintf("typecheck: operator '%s' requires numeric operands", operator.Literal),
				Position: operator.Position,
				HelpMsg:  fmt.Sprintf("got '%s' and '%s'", a, b),
			}
		}
		return a, nil
	case chToken.EQUALS, chToken.NOT_EQUALS, chToken.LESS,
		chToken.LESS_EQUALS, chToken.GREATER, chToken.GREATER_EQUALS, chToken.AND, chToken.OR:
		if a == b {
			return symbols.SymbolTypeBool, nil
		}
		return symbols.SymbolTypeInvalid, fmt.Errorf("typecheck: operator '%s' requires operands of the same type (left: %s, right: %s)", operator.Literal, a, b)
	}
	return symbols.SymbolTypeInvalid, fmt.Errorf("typecheck: unknown operator: %s", operator.Literal)
}
