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
		Used: true,
		Name: "println",
		Type: &symbols.ChlangFunctionType{ // (...string) -> void
			SpreadType: symbols.SymbolTypeString,
			Return:     symbols.SymbolTypeVoid,
		},
		EntityType: symbols.SymbolEntityFunction,
		Function: &symbols.FuncSymbolSignature{
			SpreadArgument: true,
			Args:           make([]*symbols.SymbolEntity, 0),
			ReturnType:     symbols.SymbolTypeVoid,
		},
	})
}

func (c *Checker) visitStatement(statement ast.Statement) {
	switch stmt := statement.(type) {
	case *ast.TypeDeclarationStatement:
		c.visitTypeDeclaration(stmt)
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
			Type:       symbols.SymbolTypeInt32,
			EntityType: symbols.SymbolEntityVariable,
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
		if !symbols.IsLeftCompatibleType(signature.ReturnType, exprReturnType) {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("function '%s' returns '%s', but expression type is '%s'", c.function.Name, signature.ReturnType, exprReturnType),
				Position: stmt.Span.Start,
			})
		}
	default:
		log.Printf("[checker.go]: Unknown statement type: %T", stmt)
	}
}

func (c *Checker) visitTypeDeclaration(stmt *ast.TypeDeclarationStatement) {
	if sym := c.SymbolTable.LookupInScope(stmt.Name.Value); sym != nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("type '%s' has already been declared at %s", stmt.Name.Value, sym.Position),
			Position: stmt.Span.Start,
		})
		return
	}

	// parse ast type specification
	typeSpec := c.convertASTType(stmt.Spec)
	symbols := &symbols.SymbolEntity{
		Name:       stmt.Name.Value,
		Type:       typeSpec,
		EntityType: symbols.SymbolEntityUserDefined,
		Position:   stmt.Span.Start,
	}
	c.SymbolTable.Insert(symbols)
}

func (c *Checker) convertASTType(spec ast.Expression) symbols.ChlangType {
	switch s := spec.(type) {
	case *ast.Identifier:
		if tag, ok := symbols.GetPrimitiveTypeByTag(s.Value); ok {
			return tag
		}
		symbol := c.SymbolTable.Lookup(s.Value)
		if symbol == nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("type '%s' not found", s.Value),
				Position: s.Span.Start,
			})
			return symbols.SymbolTypeInvalid
		}
		if symbol.EntityType != symbols.SymbolEntityUserDefined {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("'%s' is not a type", s.Value),
				Position: s.Span.Start,
			})
			return symbols.SymbolTypeInvalid
		}
		return symbol.Type
	case *ast.ArrayType:
		arrayType := symbols.ChlangArrayType{
			ElementType: c.convertASTType(s.Type),
		}
		if s.Size != nil {
			size, ok := s.Size.(*ast.IntLiteral)
			if !ok {
				fmt.Printf("[warn]: array size must be a constant integer: %T\n", s.Size)
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  "array size must be a constant integer",
					Position: s.Size.GetSpan().Start,
				})
				return symbols.SymbolTypeInvalid
			}
			sizeType, length := c.inferIntLiteral(size)
			if sizeType == symbols.SymbolTypeInvalid {
				return symbols.SymbolTypeInvalid
			}
			arrayType.Length = int(length)
		}
		return arrayType
	default:
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("unknown type specification: %T", spec),
			Position: spec.GetSpan().Start,
		})
		return symbols.SymbolTypeInvalid
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

	var constValueType symbols.ChlangPrimitiveType

	// double-check, because we already checking for an initial value during the parsing stage
	if stmt.Value == nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  "constants must have an initial value",
			Position: stmt.Span.Start,
		})
		return
	} else {
		exprType := c.inferExpression(stmt.Value)
		if _, ok := exprType.(symbols.ChlangPrimitiveType); !ok {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("invalid type of constant '%s'", stmt.Name.Value),
				HelpMsg:  "only primitive types support for a constant declaration",
				Position: stmt.Span.Start,
				Span:     stmt.Span,
			})
			return
		}
		constValueType = exprType.(symbols.ChlangPrimitiveType)
	}

	if stmt.Type != nil {
		constType := c.convertASTType(stmt.Type)
		if !symbols.IsLeftCompatibleType(constType, constValueType) {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message: fmt.Sprintf("constant '%s' has type '%s', but value type is '%s'", stmt.Name.Value, constType, constValueType),
			})
			return
		}
	}

	symbol := &symbols.SymbolEntity{
		Name:       stmt.Name.Value,
		Type:       constValueType,
		EntityType: symbols.SymbolEntityConstant,
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

	var varType symbols.ChlangType
	if stmt.Value != nil {
		varType = c.inferExpression(stmt.Value)
		if stmt.Type == nil {
			if ty, ok := varType.(symbols.ChlangPrimitiveType); ok {
				if ty.IsFloat() {
					varType = symbols.SymbolTypeFloat64
				}
				if ty.IsUnsigned() {
					varType = symbols.GetMaxType(symbols.SymbolTypeUint32, ty)
				}
				if ty.IsSigned() {
					varType = symbols.GetMaxType(symbols.SymbolTypeInt32, ty)
				}
			} else {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  "unsupported type for variable declaration",
					Position: stmt.Span.Start,
				})
				return
			}
		} else {
			typeTag := c.convertASTType(stmt.Type)
			if !symbols.IsLeftCompatibleType(typeTag, varType) {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  fmt.Sprintf("variable '%s' has type '%s', but value type is '%s'", stmt.Name.Value, typeTag, varType),
					Position: stmt.Span.Start,
				})
				return
			}
			varType = typeTag
		}
	} else if stmt.Type == nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("variable '%s' must have a type or an initial value", stmt.Name.Value),
			Position: stmt.Span.Start,
		})
		return
	} else {
		varType = c.convertASTType(stmt.Type)
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
		EntityType: symbols.SymbolEntityVariable,
		Position:   stmt.Span.Start,
	}
	c.SymbolTable.Insert(symbol)
	stmt.Symbol = symbol
}

// Checks function signature and adds it into the symbol table
func (c *Checker) visitFuncSignature(decl *ast.FuncDeclarationStatement) {
	if _, ok := symbols.GetPrimitiveTypeByTag(decl.Name.Value); ok {
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

	functionType := &symbols.ChlangFunctionType{}

	// infer return type
	if decl.ReturnType == nil {
		functionType.Return = symbols.SymbolTypeVoid
	} else {
		functionType.Return = c.convertASTType(decl.ReturnType)
	}

	if functionType.Return == symbols.SymbolTypeInvalid {
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

		if functionType.Return != symbols.SymbolTypeVoid {
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
		Type:       functionType,
		EntityType: symbols.SymbolEntityFunction,
		Position:   decl.Span.Start,
		Function: &symbols.FuncSymbolSignature{
			Args:       make([]*symbols.SymbolEntity, 0),
			ReturnType: functionType.Return,
		},
	}

	for _, arg := range decl.Params {
		argType := c.convertASTType(arg.Type)
		if argType == symbols.SymbolTypeInvalid {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("unknown type '%s' for argument '%s'", arg.Type, arg.Name.Value),
				Position: arg.Name.Span.Start,
			})
		} else if argType == symbols.SymbolTypeVoid {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  "cannot use 'void' as a type for function argument",
				Position: arg.Type.GetSpan().Start,
			})
		}

		argSymbol := &symbols.SymbolEntity{
			EntityType: symbols.SymbolEntityVariable,
			Type:       argType,
			Name:       arg.Name.Value,
			Position:   arg.Name.Span.Start,
			Used:       false,
		}

		functionType.Args = append(functionType.Args, argType)

		// Note that we don't insert the arguments into the symbol table because we don't check the function body here except for the arguments.
		// Arguments will populates into the symbol table in 'visitFuncDeclaration' function
		funcSymbol.Function.Args = append(funcSymbol.Function.Args, argSymbol)
	}

	c.SymbolTable.Insert(funcSymbol)

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
func (c *Checker) inferExpression(expr ast.Expression) symbols.ChlangType {
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
			if !symbols.IsLeftCompatibleType(leftType, rightType) {
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
		if sym.EntityType != symbols.SymbolEntityFunction {
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
				if !symbols.IsLeftCompatibleType(argSymbol, argExprType) {
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
		if e.Suffix != "" {
			return c.checkIntLiteralSuffix(e)
		}
		intType, _ := c.inferIntLiteral(e)
		e.Type = intType // save the type for the code generation
		return intType
	case *ast.FloatLiteral:
		bitSize := 64
		symbolType := symbols.SymbolTypeFloat64

		if e.Suffix != "" {
			switch e.Suffix {
			case "f32":
				bitSize = 32
				symbolType = symbols.SymbolTypeFloat32
			case "f64":
				bitSize = 64
			default:
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  fmt.Sprintf("unknown float type suffix '%s'", e.Suffix),
					Position: e.Span.Start,
				})
			}
		}

		if _, err := strconv.ParseFloat(e.Value, bitSize); err != nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("value '%s' is out of float %d-bit range", e.Value, bitSize),
				Position: e.Span.Start,
			})
			return symbols.SymbolTypeInvalid
		}

		e.Type = symbolType // save the type for the code generation
		return symbolType
	case *ast.BoolLiteral:
		return symbols.SymbolTypeBool
	case *ast.ArrayExpression:
		arrayType := &symbols.ChlangArrayType{
			ElementType: symbols.SymbolTypeInvalid,
			Length:      0,
		}
		for _, elem := range e.Elements {
			elemType := c.inferExpression(elem)
			if arrayType.ElementType == symbols.SymbolTypeInvalid {
				arrayType.ElementType = elemType
			} else if arrayType.ElementType != elemType {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  fmt.Sprintf("array element type mismatch: expected '%s', but got '%s'", arrayType.ElementType, elemType),
					Position: elem.GetSpan().Start,
				})
			}
		}
		return arrayType
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
			operandType, ok := rightType.(symbols.ChlangPrimitiveType)
			if !ok || !operandType.IsNumeric() {
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
		inferred, err := c.checkTypesCompatibility(leftType, rightType, e.Operator)

		if err != nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  err.Error(),
				Position: e.Span.Start,
			})
			return symbols.SymbolTypeInvalid
		}

		return inferred
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

		var elseType symbols.ChlangType
		switch elseBlock := e.ElseBlock.(type) {
		case *ast.BlockStatement:
			elseType = c.inferIfBlockStatement(elseBlock)
		case *ast.IfExpression:
			elseType = c.inferExpression(elseBlock)
		}

		if !symbols.IsCompatibleType(thenType, elseType) {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("cannot determine a single type of if expression (then: %s, else: %s)", thenType, elseType),
				HelpMsg:  "",
				Span:     e.Span,
				Position: e.Span.Start,
			})
			return symbols.SymbolTypeInvalid
		}

		// TODO: Refactor this, because it's not a good way to determine the type of the if expression
		// It's better to return a composite type and check it in the upper level
		return c.getMaxTypeOf(thenType, elseType)
	}

	c.Errors = append(c.Errors, &errors.SemanticError{
		Message: fmt.Sprintf("type infer: unknown expression type: %T", expr),
	})
	return symbols.SymbolTypeInvalid
}

func (c *Checker) inferIfBlockStatement(block *ast.BlockStatement) symbols.ChlangType {
	c.SymbolTable.OpenScope()
	c.populateSymbolDeclarations(block.Statements)
	var returnType symbols.ChlangType = symbols.SymbolTypeVoid
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

func (c *Checker) inferIntLiteral(node *ast.IntLiteral) (symbols.ChlangPrimitiveType, int64) {
	bitSize := 64
	intValue, err := strconv.ParseInt(node.Value, 0, bitSize)
	if err != nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("value '%s' is out of integer %d-bit range", node.Value, bitSize),
			Position: node.Span.Start,
		})
		return symbols.SymbolTypeInvalid, 0
	}

	if intValue >= -128 && intValue <= 127 {
		return symbols.SymbolTypeInt8, intValue
	}

	if intValue >= -32768 && intValue <= 32767 {
		return symbols.SymbolTypeInt16, intValue
	}

	if intValue >= -2147483648 && intValue <= 2147483647 {
		return symbols.SymbolTypeInt32, intValue
	}

	if intValue >= -9223372036854775808 && intValue <= 9223372036854775807 {
		return symbols.SymbolTypeInt64, intValue
	}

	return symbols.SymbolTypeInvalid, intValue
}

func (c *Checker) checkIntLiteralSuffix(node *ast.IntLiteral) symbols.ChlangPrimitiveType {
	mode := node.Suffix[0]
	bitSize, err := strconv.Atoi(node.Suffix[1:])

	if mode == 'f' {
		if _, err := strconv.ParseFloat(node.Value, bitSize); err != nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("value '%s' is out of float %d-bit range", node.Value, bitSize),
				Position: node.Span.Start,
			})
			return symbols.SymbolTypeInvalid
		}

		if bitSize == 32 {
			return symbols.SymbolTypeFloat32
		}

		return symbols.SymbolTypeFloat64
	} else if mode == 'u' { // unsigned int
		if err != nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("invalid integer type suffix '%s'", node.Suffix),
				Position: node.Span.Start,
			})
			return symbols.SymbolTypeInvalid
		}

		uintValue, err := strconv.ParseUint(node.Value, 0, bitSize)
		if err != nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("value '%s' is out of unsigned %d-bit range", node.Value, bitSize),
				Position: node.Span.Start,
			})
			return symbols.SymbolTypeInvalid
		}

		if bitSize == 8 && uintValue <= 255 {
			return symbols.SymbolTypeUint8
		}

		if bitSize == 16 && uintValue <= 65535 {
			return symbols.SymbolTypeUint16
		}

		if bitSize == 32 && uintValue <= 4294967295 {
			return symbols.SymbolTypeUint32
		}

		if bitSize == 64 && uintValue <= 18446744073709551615 {
			return symbols.SymbolTypeUint64
		}
	} else if mode == 'i' {
		intValue, err := strconv.ParseInt(node.Value, 0, bitSize)
		if err != nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("value '%s' is out of integer %d-bit range", node.Value, bitSize),
				Position: node.Span.Start,
			})
			return symbols.SymbolTypeInvalid
		}

		if bitSize == 8 && intValue >= -128 && intValue <= 127 {
			return symbols.SymbolTypeInt8
		}

		if bitSize == 16 && intValue >= -32768 && intValue <= 32767 {
			return symbols.SymbolTypeInt16
		}

		if bitSize == 32 && intValue >= -2147483648 && intValue <= 2147483647 {
			return symbols.SymbolTypeInt32
		}

		if bitSize == 64 && intValue >= -9223372036854775808 && intValue <= 9223372036854775807 {
			return symbols.SymbolTypeInt64
		}
	}

	c.Errors = append(c.Errors, &errors.SemanticError{
		Message:  fmt.Sprintf("unknown integer type suffix '%s'", node.Suffix),
		Position: node.Span.Start,
	})

	return symbols.SymbolTypeInvalid
}

func (c *Checker) checkTypesCompatibility(a, b symbols.ChlangType, operator *chToken.Token) (symbols.ChlangType, error) {
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
		leftType, leftIsPrimitive := a.(symbols.ChlangPrimitiveType)
		rightType, rightIsPrimitive := b.(symbols.ChlangPrimitiveType)

		if leftIsPrimitive && rightIsPrimitive && leftType.IsNumeric() && rightType.IsNumeric() {
			if a == b {
				return a, nil
			}
			if leftType.IsFloat() || rightType.IsFloat() {
				return symbols.SymbolTypeFloat64, nil
			}

			if leftType.IsSigned() && rightType.IsSigned() {
				return symbols.GetMaxType(leftType, rightType), nil
			}

			if leftType.IsUnsigned() && rightType.IsUnsigned() {
				return symbols.GetMaxType(leftType, rightType), nil
			}
		}

		return symbols.SymbolTypeInvalid, &errors.SemanticError{
			Message:  fmt.Sprintf("type mismatch: incompatible types (left: %s, right: %s) for operator '%s'", a, b, operator.Literal),
			Position: operator.Position,
			HelpMsg:  fmt.Sprintf("got '%s' and '%s'", a, b),
		}
	case chToken.EQUALS, chToken.NOT_EQUALS, chToken.LESS,
		chToken.LESS_EQUALS, chToken.GREATER, chToken.GREATER_EQUALS, chToken.AND, chToken.OR:
		if symbols.IsCompatibleType(a, b) {
			return symbols.SymbolTypeBool, nil
		}
		return symbols.SymbolTypeInvalid, fmt.Errorf("type mismatch: operator '%s' requires operands of the same type (left: %s, right: %s)", operator.Literal, a, b)
	}
	return symbols.SymbolTypeInvalid, fmt.Errorf("type mismatch: unknown operator: %s", operator.Literal)
}

func (c *Checker) getMaxTypeOf(left, right symbols.ChlangType) symbols.ChlangType {
	leftType, leftIsPrimitive := left.(symbols.ChlangPrimitiveType)
	rightType, rightIsPrimitive := right.(symbols.ChlangPrimitiveType)

	if leftIsPrimitive && rightIsPrimitive {
		return symbols.GetMaxType(leftType, rightType)
	}

	fmt.Printf("[warn]: getMaxTypeOf: unsupported types: %s, %s\n", left, right)
	return left
}
