package checker

import (
	"fmt"
	"log"
	"strconv"

	"github.com/usein-abilev/chlang/frontend/ast"
	"github.com/usein-abilev/chlang/frontend/checker/env"
	"github.com/usein-abilev/chlang/frontend/errors"
	chToken "github.com/usein-abilev/chlang/frontend/token"
)

type Checker struct {
	// Current symbol table (scope)
	Env      *env.Env
	Errors   []error
	Warnings []error

	// Current function being checked
	function *env.EnvSymbolEntity
}

// Check performs semantic analysis on the AST
// It populates the symbol table and checks for type mismatches
// It also will transform the AST into a more optimized form
func Check(program *ast.Program, env *env.Env) *Checker {
	c := &Checker{Env: env}

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
		case *ast.TypeDeclarationStatement:
			c.visitTypeDeclaration(decl)
		case *ast.StructDeclarationStatement:
			c.visitStructDeclaration(decl)
		case *ast.TraitDeclarationStatement:
			c.visitTraitDeclaration(decl)
		case *ast.ImplStatement:
			c.visitImplDeclaration(decl)
		case *ast.FuncDeclarationStatement:
			c.visitFuncSignature(decl)
		}
	}
}

// Adds built-in functions to the symbol table
func (c *Checker) addBuiltinFunctions() {
	c.Env.InsertSymbol(&env.EnvSymbolEntity{
		Used: true,
		Name: "println",
		Type: &env.ChlangFunctionType{ // (...string) -> void
			SpreadType: env.SymbolTypeString,
			Return:     env.SymbolTypeVoid,
		},
		EntityType: env.SymbolEntityFunction,
	})
}

func (c *Checker) visitStatement(statement ast.Statement) {
	switch stmt := statement.(type) {
	case *ast.TypeDeclarationStatement,
		*ast.StructDeclarationStatement,
		*ast.TraitDeclarationStatement,
		*ast.ImplStatement:
		// Types declarations are already processed in the 'populateSymbolDeclarations' method
	case *ast.ConstDeclarationStatement:
		c.visitConstDeclaration(stmt)
	case *ast.VarDeclarationStatement:
		c.visitVarDeclaration(stmt)
	case *ast.FuncDeclarationStatement:
		c.visitFuncBody(stmt)
	case *ast.ExpressionStatement:
		c.inferExpression(stmt.Expression)
	case *ast.ForRangeStatement:
		c.Env.OpenScope()

		// insert range variable into the symbol table
		rangeVar := &env.EnvSymbolEntity{
			Used:       false,
			Name:       stmt.Identifier.Value,
			Type:       env.SymbolTypeInt32,
			EntityType: env.SymbolEntityVariable,
			Span:       stmt.Span,
		}
		c.Env.InsertSymbol(rangeVar)
		stmt.Identifier.Symbol = rangeVar

		// We don't need a new block scope, because the for-range statement is already a block statement
		for _, statement := range stmt.Body.Statements {
			c.visitStatement(statement)
		}

		c.Env.CloseScope()
	case *ast.BlockStatement:
		c.Env.OpenScope()
		c.populateSymbolDeclarations(stmt.Statements)
		for _, statement := range stmt.Statements {
			c.visitStatement(statement)
		}
		c.Env.CloseScope()
	case *ast.ReturnStatement:
		exprReturnType := c.inferExpression(stmt.Expression)
		if c.function == nil {
			c.reportError("unexpected 'return' statement outside the function", stmt.Span)
			return
		}
		signature := c.function.Type.(*env.ChlangFunctionType)
		if !env.IsLeftCompatibleType(signature.Return, exprReturnType) {
			c.reportError(
				fmt.Sprintf("function '%s' returns '%s', but expression type is '%s'", c.function.Name, signature.Return, exprReturnType),
				stmt.Span,
			)
		}
	default:
		log.Printf("[checker.go]: Unknown statement type: %T", stmt)
	}
}

func (c *Checker) visitTypeDeclaration(stmt *ast.TypeDeclarationStatement) {
	panic("Type declaration is not implemented yet")
	// if sym := c.Env.LookupType(stmt.Name.Value); sym != nil {
	// 	c.Errors = append(c.Errors, &errors.SemanticError{
	// 		Message:  fmt.Sprintf("type '%s' has already been declared at %s", stmt.Name.Value, sym.Span),
	// 		Span:     stmt.Span,
	// 		Position: stmt.Span.Start,
	// 	})
	// 	return
	// }

	// typeSpec := c.resolveASTType(stmt.Spec)
	// typeEntity := &env.EnvTypeEntity{
	// 	Name: stmt.Name.Value,
	// 	Used: false,
	// 	Spec: &env.ChlangUserType{
	// 		Name: stmt.Name.Value,
	// 		Spec: typeSpec,
	// 	},
	// }
	// c.Env.InsertType(typeEntity)
}

func (c *Checker) visitStructDeclaration(stmt *ast.StructDeclarationStatement) {
	if sym := c.Env.LookupType(stmt.Name.Value); sym != nil {
		c.reportError(
			fmt.Sprintf("struct '%s' has already been declared at %s", stmt.Name.Value, sym.Span),
			stmt.Span,
		)
		return
	}

	structType := &env.ChlangStructType{
		Name:   stmt.Name.Value,
		Fields: make([]*env.ChlangStructField, 0),
	}
	for _, field := range stmt.Body.Fields {
		if exists := structType.LookupField(field.Name.Value); exists != nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("duplicate field '%s' in struct '%s'", field.Name.Value, stmt.Name.Value),
				Position: field.Name.Span.Start,
			})
			return
		}
		fieldType := c.resolveASTType(field.Value)
		if fieldType == env.SymbolTypeInvalid {
			return
		}
		structType.Fields = append(structType.Fields, &env.ChlangStructField{
			Name: field.Name.Value,
			Type: fieldType,
		})
	}

	structEntity := &env.EnvTypeEntity{
		Name: stmt.Name.Value,
		Used: false,
		Spec: structType,
	}
	c.Env.InsertType(structEntity)
}

func (c *Checker) visitTraitDeclaration(stmt *ast.TraitDeclarationStatement) {
	if sym := c.Env.LookupType(stmt.Name.Value); sym != nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("trait '%s' has already been declared at %s", stmt.Name.Value, sym.Span),
			Span:     stmt.Span,
			Position: stmt.Span.Start,
		})
		return
	}

	// TODO: Implement trait type
	traitEntity := &env.EnvTypeEntity{
		Name: stmt.Name.Value,
		Used: false,
		Spec: &env.ChlangTraitType{
			Name:       stmt.Name.Value,
			Signatures: []*env.ChlangFunctionType{},
			// Declarations: []*hir.Function{},
		},
	}

	if len(stmt.MethodDeclarations) > 0 {
		panic("Trait method declarations are not implemented yet")
	}

	c.Env.InsertType(traitEntity)
}

func (c *Checker) visitImplDeclaration(implStmt *ast.ImplStatement) {
	receiverType := c.Env.LookupType(implStmt.Receiver.Value)
	if receiverType == nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("type '%s' not found", implStmt.Receiver.Value),
			Position: implStmt.Span.Start,
		})
		return
	}

	switch structType := receiverType.Spec.(type) {
	case *env.ChlangStructType:
		for _, implMethod := range implStmt.Methods {
			fn := structType.LookupMethod(implMethod.Signature.Name.Value)
			// check if the method is already implemented in the struct
			if fn != nil {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  fmt.Sprintf("method '%s' is already implemented in struct '%s'", implMethod.Signature.Name.Value, structType.Name),
					Position: implMethod.Span.Start,
				})
			} else if (structType.LookupField(implMethod.Signature.Name.Value)) != nil {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  fmt.Sprintf("method '%s' has the same name as a field in struct '%s'", implMethod.Signature.Name.Value, structType.Name),
					Position: implMethod.Span.Start,
				})
			} else {
				// TODO: check if the method signature is compatible with the trait method signature
			}
		}
		for _, trait := range implStmt.Traits {
			traitType := c.Env.LookupType(trait.Value)
			if traitType == nil {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  fmt.Sprintf("trait '%s' not found", trait.Value),
					Position: trait.Span.Start,
					Span:     trait.Span,
				})
				return
			}

			// TODO: Check if the trait is implemented by the struct
		}
	}
	panic("ImplStatement is not implemented yet")
}

func (c *Checker) resolveASTType(spec ast.Expression) env.ChlangType {
	switch s := spec.(type) {
	case *ast.Identifier:
		ty := c.Env.LookupType(s.Value)
		if ty == nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("type '%s' not found", s.Value),
				Position: s.Span.Start,
			})
			return env.SymbolTypeInvalid
		}
		ty.Used = true
		return ty.Spec
	case *ast.ArrayType:
		arrayType := &env.ChlangArrayType{
			ElementType: c.resolveASTType(s.Type),
		}
		if s.Size != nil {
			size, ok := s.Size.(*ast.IntLiteral)
			if !ok {
				fmt.Printf("[warn]: array size must be a constant integer: %T\n", s.Size)
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  "array size must be a constant integer",
					Position: s.Size.GetSpan().Start,
				})
				return env.SymbolTypeInvalid
			}
			sizeType, length := c.inferIntLiteral(size)
			if sizeType == env.SymbolTypeInvalid {
				return env.SymbolTypeInvalid
			}
			arrayType.Length = int(length)
		}
		return arrayType
	case *ast.StructType:
		structType := &env.ChlangStructType{
			Fields: make([]*env.ChlangStructField, 0),
		}
		for _, field := range s.Fields {
			fieldType := c.resolveASTType(field.Value)
			if fieldType == env.SymbolTypeInvalid {
				return env.SymbolTypeInvalid
			}
			structType.Fields = append(structType.Fields, &env.ChlangStructField{
				Name: field.Name.Value,
				Type: fieldType,
			})
		}
		return structType
	default:
		c.reportError(fmt.Sprintf("unknown type specification: %T", spec), spec.GetSpan())
		return env.SymbolTypeInvalid
	}
}

func (c *Checker) visitConstDeclaration(stmt *ast.ConstDeclarationStatement) {
	if ty := c.Env.LookupType(stmt.Name.Value); ty != nil {
		c.reportError(fmt.Sprintf("cannot use type '%s' as a constant name", stmt.Name.Value), stmt.Span)
		return
	}

	if sym := c.Env.LookupSymbolLocal(stmt.Name.Value); sym != nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("constant '%s' has already been declared at %s", stmt.Name.Value, sym.Span),
			Position: stmt.Span.Start,
			Span:     stmt.Span,
		})
		return
	}

	var constValueType env.ChlangPrimitiveType

	// double-check, because we already checking for an initial value during the parsing stage
	if stmt.Value == nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  "constants must have an initial value",
			Position: stmt.Span.Start,
		})
		return
	} else {
		exprType := c.inferExpression(stmt.Value)
		if _, ok := exprType.(env.ChlangPrimitiveType); !ok {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("invalid type of constant '%s'", stmt.Name.Value),
				HelpMsg:  "only primitive types support for a constant declaration",
				Position: stmt.Span.Start,
				Span:     stmt.Span,
			})
			return
		}
		constValueType = exprType.(env.ChlangPrimitiveType)
	}

	if stmt.Type != nil {
		constType := c.resolveASTType(stmt.Type)
		if !env.IsLeftCompatibleType(constType, constValueType) {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message: fmt.Sprintf("constant '%s' has type '%s', but value type is '%s'", stmt.Name.Value, constType, constValueType),
			})
			return
		}
	}

	symbol := &env.EnvSymbolEntity{
		Name:       stmt.Name.Value,
		Type:       constValueType,
		EntityType: env.SymbolEntityConstant,
		Span:       stmt.Span,
	}
	c.Env.InsertSymbol(symbol)
	stmt.Symbol = symbol
}

func (c *Checker) visitVarDeclaration(stmt *ast.VarDeclarationStatement) {
	if sym := c.Env.LookupSymbolLocal(stmt.Name.Value); sym != nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("variable '%s' has already been declared at %s", stmt.Name.Value, sym.Span),
			Position: stmt.Span.Start,
			Span:     stmt.Span,
		})
		return
	}

	var varType env.ChlangType
	if stmt.Value != nil {
		varType = c.inferExpression(stmt.Value)
		if stmt.Type == nil {
			varType = c.getGeneralTypeOf(varType)
		} else {
			typeTag := c.resolveASTType(stmt.Type)
			if !env.IsLeftCompatibleType(typeTag, varType) {
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
		varType = c.resolveASTType(stmt.Type)
	}

	if varType == env.SymbolTypeInvalid {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("invalid type of variable '%s'", stmt.Name.Value),
			Position: stmt.Span.Start,
		})
	}

	symbol := &env.EnvSymbolEntity{
		Name:       stmt.Name.Value,
		Type:       varType,
		EntityType: env.SymbolEntityVariable,
		Span:       stmt.Span,
	}
	c.Env.InsertSymbol(symbol)
	stmt.Symbol = symbol
}

// Checks function signature and adds it into the symbol table
func (c *Checker) visitFuncSignature(decl *ast.FuncDeclarationStatement) {
	if ty := c.Env.LookupType(decl.Signature.Name.Value); ty != nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("cannot use type '%s' as a function name", decl.Signature.Name.Value),
			Position: decl.Span.Start,
		})
		return
	}

	if sym := c.Env.LookupSymbolLocal(decl.Signature.Name.Value); sym != nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("function '%s' has already been declared at %s", decl.Signature.Name.Value, sym.Span),
			Position: decl.Span.Start,
			Span:     decl.Span,
		})
		return
	}

	functionType := &env.ChlangFunctionType{}

	// infer return type
	if decl.Signature.ReturnType == nil {
		functionType.Return = env.SymbolTypeVoid
	} else {
		functionType.Return = c.resolveASTType(decl.Signature.ReturnType)
	}

	if functionType.Return == env.SymbolTypeInvalid {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("invalid function '%s' return type", decl.Signature.Name.Value),
			Position: decl.Span.Start,
		})
		return
	}

	// if the function is entry point, is already used
	used := false
	if decl.Signature.Name.Value == "main" {
		used = true

		if functionType.Return != env.SymbolTypeVoid {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  "main function must return void",
				Position: decl.Span.Start,
			})
			return
		}
	}

	funcSymbol := &env.EnvSymbolEntity{
		Name:       decl.Signature.Name.Value,
		Used:       used,
		Type:       functionType,
		EntityType: env.SymbolEntityFunction,
		Span:       decl.Span,
	}

	for _, arg := range decl.Signature.Args {
		argType := c.resolveASTType(arg.Type)
		if argType == env.SymbolTypeInvalid {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("unknown type '%s' for argument '%s'", arg.Type, arg.Name.Value),
				Position: arg.Name.Span.Start,
			})
		} else if argType == env.SymbolTypeVoid {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  "cannot use 'void' as a type for function argument",
				Position: arg.Type.GetSpan().Start,
			})
		}

		argSymbol := &env.EnvSymbolEntity{
			EntityType: env.SymbolEntityVariable,
			Type:       argType,
			Name:       arg.Name.Value,
			Span:       arg.Name.GetSpan(),
			Used:       false,
		}

		functionType.Args = append(functionType.Args, argType)

		// Note that we don't insert the arguments into the symbol table because we don't check the function body here except for the arguments.
		// Arguments will populates into the symbol table in 'visitFuncDeclaration' function
		funcSymbol.FunctionArgs = append(funcSymbol.FunctionArgs, argSymbol)
	}

	if ok := c.Env.InsertSymbol(funcSymbol); !ok {
		panic("unexpected error: function symbol already exists")
	}
	decl.Symbol = funcSymbol
}

// Checks function body for type matching
func (c *Checker) visitFuncBody(stmt *ast.FuncDeclarationStatement) {
	funcSymbol := c.Env.LookupSymbolLocal(stmt.Signature.Name.Value)
	if funcSymbol == nil {
		// Since function declarations populates in the 'populateSymbolDeclaration' method
		// There is no way to have 'nil' result of lookup
		panic(fmt.Sprintf("Unexpected nil as result of lookupInScope function '%s'", stmt.Signature.Name.Value))
	}

	// check function arguments and visit function body
	c.Env.OpenScope()
	c.populateSymbolDeclarations(stmt.Body.Statements)
	prevFuncPtr := c.function
	c.function = funcSymbol

	// Pushing arguments into symbol table, we didn't check types
	// because it already done in the 'populateSymbolDeclarations' method
	for _, arg := range funcSymbol.FunctionArgs {
		c.Env.InsertSymbol(arg)
	}

	c.visitStatement(stmt.Body)
	c.Env.CloseScope()
	c.function = prevFuncPtr

	stmt.Symbol = funcSymbol
}

// Check expression type and return its internal type
// If the expression is nil, return env.SymbolTypeVoid
func (c *Checker) inferExpression(expr ast.Expression) env.ChlangType {
	if expr == nil {
		return env.SymbolTypeVoid
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		sym := c.Env.LookupSymbol(e.Value)
		if sym == nil {
			c.reportError(fmt.Sprintf("identifier '%s' not found", e.Value), e.Span)
			return env.SymbolTypeInvalid
		}
		sym.Used = true
		e.Symbol = sym
		return sym.Type
	case *ast.InitStructExpression:
		sym := c.Env.LookupType(e.Name.Value)
		if sym == nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("struct '%s' not found", e.Name.Value),
				Position: e.Span.Start,
				Span:     e.Span,
			})
			return env.SymbolTypeInvalid
		}
		structType, isStructType := sym.Spec.(*env.ChlangStructType)
		if !isStructType {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("'%s' is not a struct", e.Name.Value),
				Position: e.Span.Start,
				Span:     e.Span,
			})
			return env.SymbolTypeInvalid
		}
		if len(e.Fields) != len(structType.Fields) {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("struct '%s' expects %d fields, but got %d", e.Name.Value, len(structType.Fields), len(e.Fields)),
				Position: e.Span.Start,
				Span:     e.Span,
			})
			return env.SymbolTypeInvalid
		}

		// type checking of struct fields
		for _, field := range e.Fields {
			structField := structType.LookupField(field.Name.Value)
			if structField == nil {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  fmt.Sprintf("field '%s' not found in struct '%s'", field.Name.Value, e.Name.Value),
					Position: field.Name.Span.Start,
				})
				return env.SymbolTypeInvalid
			}
			fieldType := c.inferExpression(field.Value)
			if !env.IsLeftCompatibleType(structField.Type, fieldType) {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  fmt.Sprintf("field '%s' expects type '%s', but got '%s'", field.Name.Value, structField.Type, fieldType),
					Position: field.Value.GetSpan().Start,
				})
				return env.SymbolTypeInvalid
			}
		}
		sym.Used = true
		return sym.Spec
	case *ast.AssignExpression:
		leftType := c.inferExpression(e.Left)
		rightType := c.inferExpression(e.Right)

		if leftType == env.SymbolTypeInvalid || rightType == env.SymbolTypeInvalid {
			return env.SymbolTypeInvalid
		}

		if chToken.IsAssignment(e.Operator.Type) {
			switch e.Left.(type) {
			case *ast.Identifier:
			case *ast.IndexExpression:
			default:
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  "left side of an assignment must be an identifier",
					Position: e.Span.Start,
				})
				return env.SymbolTypeInvalid
			}
			if !env.IsLeftCompatibleType(leftType, rightType) {
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
				return env.SymbolTypeInvalid
			}
			return leftType
		}

		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("type infer: unknown assign operator type: %s", e.Operator.Literal),
			Span:     e.Span,
			Position: e.Operator.Position,
		})
		return env.SymbolTypeInvalid
	case *ast.MemberExpression:
		panic("MemberExpression is not implemented yet")
	case *ast.CallExpression:
		var fnSymbol *env.EnvSymbolEntity

		switch callee := e.Function.(type) {
		case *ast.Identifier:
			sym := c.Env.LookupSymbol(callee.Value)
			if sym == nil {
				c.reportError(fmt.Sprintf("function '%s' not found", callee.Value), e.Span)
				return env.SymbolTypeInvalid
			}
			if sym.EntityType != env.SymbolEntityFunction {
				c.reportError(fmt.Sprintf("'%s' is not a function", callee.Value), e.Span)
				return env.SymbolTypeInvalid
			}
			fnSymbol = sym
		case *ast.MemberExpression:
			panic("MemberExpression is not implemented yet")
			// ty, monoFunc := c.inferMemberExpression(expr.(*ast.MemberExpression))
			// if ty == env.SymbolTypeInvalid || monoFunc == nil {
			// 	c.reportError(fmt.Sprintf("unknown function '%s'", callee.Member.Value), e.Span)
			// 	return env.SymbolTypeInvalid
			// }
			// if _, ok := ty.(*env.ChlangFunctionType); !ok {
			// 	c.reportError(fmt.Sprintf("'%s' is not a function", callee.Member.Value), e.Span)
			// 	return env.SymbolTypeInvalid
			// }
			// functionType = ty.(*env.ChlangFunctionType)
			// functionName = monoFunc.RawName
		}

		functionType := fnSymbol.Type.(*env.ChlangFunctionType)
		if functionType.SpreadType == nil {
			if len(functionType.Args) != len(e.Args) {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  fmt.Sprintf("function '%s' expects %d arguments, but got %d", fnSymbol.Name, len(functionType.Args), len(e.Args)),
					Position: e.Span.Start,
				})
				return env.SymbolTypeInvalid
			}
			for idx, argExpr := range e.Args {
				argExprType := c.inferExpression(argExpr)
				argSymbol := functionType.Args[idx]
				if !env.IsLeftCompatibleType(argSymbol, argExprType) {
					c.Errors = append(c.Errors, &errors.SemanticError{
						Message:  fmt.Sprintf("function '%s' expects argument '%s' to be '%s', but got '%s'", fnSymbol.Name, functionType.Args[idx], argSymbol, argExprType),
						Position: e.Span.Start,
					})
				}
			}
		} else {
			for _, argExpr := range e.Args {
				t := c.inferExpression(argExpr)
				if t == env.SymbolTypeInvalid {
					return env.SymbolTypeInvalid
				}
			}
			fmt.Printf("[warn]: Type checking of the spread arguments not implemented! Ignoring type check for function '%s'.\n", fnSymbol.Name)
		}
		fnSymbol.Used = true
		return functionType.Return
	case *ast.StringLiteral:
		return env.SymbolTypeString
	case *ast.IntLiteral:
		var intType env.ChlangPrimitiveType
		if e.Suffix != "" {
			intType = c.checkIntLiteralSuffix(e)
		} else {
			intType, _ = c.inferIntLiteral(e)
		}
		e.Type = intType
		return intType
	case *ast.FloatLiteral:
		bitSize := 64
		symbolType := env.SymbolTypeFloat64

		if e.Suffix != "" {
			switch e.Suffix {
			case "f32":
				bitSize = 32
				symbolType = env.SymbolTypeFloat32
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
			return env.SymbolTypeInvalid
		}

		return symbolType
	case *ast.BoolLiteral:
		return env.SymbolTypeBool
	case *ast.ArrayExpression:
		arrayType := &env.ChlangArrayType{
			ElementType: env.SymbolTypeInvalid,
			Length:      len(e.Elements),
		}
		for _, elem := range e.Elements {
			elemType := c.inferExpression(elem)
			if arrayType.ElementType == env.SymbolTypeInvalid {
				arrayType.ElementType = elemType
			} else if !env.IsCompatibleType(arrayType.ElementType, elemType) {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  fmt.Sprintf("array element type mismatch: expected '%s', but got '%s'", arrayType.ElementType, elemType),
					Position: elem.GetSpan().Start,
				})
			}
			arrayType.ElementType = c.getMaxTypeOf(arrayType.ElementType, elemType)
		}
		return arrayType
	case *ast.IndexExpression:
		arrayType := c.inferExpression(e.Left)
		if arrayType == env.SymbolTypeInvalid {
			return env.SymbolTypeInvalid
		}
		indexType := c.inferExpression(e.Index)
		if indexType == env.SymbolTypeInvalid {
			return env.SymbolTypeInvalid
		}

		if arrayType, ok := arrayType.(*env.ChlangArrayType); ok {
			if !indexType.(*env.ChlangPrimitiveType).IsInteger() {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  fmt.Sprintf("index operator requires integer type, but got '%s'", indexType),
					Position: e.Span.Start,
				})
				return env.SymbolTypeInvalid
			}
			return arrayType.ElementType
		}

		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("index operator applied to non-array type '%s'", arrayType),
			Position: e.Span.Start,
		})
		return env.SymbolTypeInvalid
	case *ast.UnaryExpression:
		rightType := c.inferExpression(e.Right)
		if rightType == env.SymbolTypeInvalid {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  "invalid unary operand",
				Position: e.Span.Start,
			})
			return env.SymbolTypeInvalid
		}
		switch e.Operator.Type {
		case chToken.MINUS, chToken.PLUS:
			operandType, ok := rightType.(env.ChlangPrimitiveType)
			if !ok || !operandType.IsNumeric() {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  fmt.Sprintf("unary operator '%s' requires only numeric operand", e.Operator.Literal),
					Position: e.Span.Start,
				})
				return env.SymbolTypeInvalid
			}
			if operandType.IsUnsigned() && e.Operator.Type == chToken.MINUS {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  "cannot apply unary minus to unsigned integer",
					Position: e.Span.Start,
				})
				return env.SymbolTypeInvalid
			}
			return rightType
		case chToken.BANG:
			if rightType != env.SymbolTypeBool {
				c.Errors = append(c.Errors, &errors.SemanticError{
					Message:  fmt.Sprintf("unary operator '%s' requires only boolean operand", e.Operator.Literal),
					Position: e.Span.Start,
				})
				return env.SymbolTypeInvalid
			}
			return rightType
		}
		return env.SymbolTypeInvalid
	case *ast.BinaryExpression:
		leftType := c.inferExpression(e.Left)
		rightType := c.inferExpression(e.Right)
		inferred, err := c.checkTypesCompatibility(leftType, rightType, e.Operator)

		if err != nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  err.Error(),
				Position: e.Span.Start,
			})
			return env.SymbolTypeInvalid
		}

		return inferred
	case *ast.IfExpression:
		condType := c.inferExpression(e.Condition)
		if condType != env.SymbolTypeBool {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("invalid condition type expected 'bool', but got '%s'", condType),
				HelpMsg:  "",
				Span:     e.Span,
				Position: e.Span.Start,
			})
			return env.SymbolTypeInvalid
		}

		thenType := c.inferIfBlockStatement(e.ThenBlock)

		if e.ElseBlock == nil {
			return thenType
		}

		var elseType env.ChlangType
		switch elseBlock := e.ElseBlock.(type) {
		case *ast.BlockStatement:
			elseType = c.inferIfBlockStatement(elseBlock)
		case *ast.IfExpression:
			elseType = c.inferExpression(elseBlock)
		}

		if !env.IsCompatibleType(thenType, elseType) {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("cannot determine a single type of if expression (then: %s, else: %s)", thenType, elseType),
				HelpMsg:  "",
				Span:     e.Span,
				Position: e.Span.Start,
			})
			return env.SymbolTypeInvalid
		}

		// TODO: Refactor this, because it's not a good way to determine the type of the if expression
		// It's better to return a composite type and check it in the upper level
		return c.getMaxTypeOf(thenType, elseType)
	}

	c.Errors = append(c.Errors, &errors.SemanticError{
		Message:  fmt.Sprintf("type infer: unknown expression type: %T", expr),
		Span:     expr.GetSpan(),
		Position: expr.GetSpan().Start,
	})
	return env.SymbolTypeInvalid
}

// getGeneralTypeOf returns the minimal general type of the type
// For example, if the types are i8 the general type will be i32 (because all integer variables by default is the i32)
// Same behavior for the complex types, the array of i8 will be the array of i32
func (c *Checker) getGeneralTypeOf(exprType env.ChlangType) env.ChlangType {
	switch ty := exprType.(type) {
	case env.ChlangPrimitiveType:
		if ty.IsFloat() {
			exprType = env.SymbolTypeFloat64
		}
		if ty.IsUnsigned() {
			exprType = env.GetMaxType(env.SymbolTypeUint32, ty)
		}
		if ty.IsSigned() {
			exprType = env.GetMaxType(env.SymbolTypeInt32, ty)
		}
	case *env.ChlangArrayType:
		ty.ElementType = c.getGeneralTypeOf(ty.ElementType)
	}
	return exprType
}

// func (c *Checker) inferMemberExpression(expr *ast.MemberExpression) (env.ChlangType, *env.MonomorphicFunction) {
// 	left := c.inferExpression(expr.Left)
// 	if left == env.SymbolTypeInvalid {
// 		return env.SymbolTypeInvalid
// 	}
// 	member := expr.Member.Value

// 	switch leftType := left.(type) {
// 	case *env.ChlangStructType:
// 		if field := leftType.LookupField(member); field != nil {
// 			return field.Type, nil
// 		}
// 		if method := leftType.LookupMethod(member); method != nil {
// 			return method.Type.(*env.EnvTypeEntity), method
// 		}
// 		c.Errors = append(c.Errors, &errors.SemanticError{
// 			Message:  fmt.Sprintf("field '%s' not found in struct", member),
// 			Position: expr.Span.Start,
// 		})
// 		return env.SymbolTypeInvalid
// 		// case *env.ChlangArrayType:
// 		// 	if member != "len" {
// 		// 		c.Errors = append(c.Errors, &errors.SemanticError{
// 		// 			Message:  fmt.Sprintf("unknown array member '%s'", member),
// 		// 			Position: e.Span.Start,
// 		// 		})
// 		// 		return env.SymbolTypeInvalid
// 		// 	}
// 		// 	return &env.ChlangFunctionType{
// 		// 		SpreadType: nil,
// 		// 		Return:     env.SymbolTypeInt32,
// 		// 		Args:       []env.ChlangType{leftType},
// 		// 	}
// 	}
// 	fmt.Printf("[warn]: Unsupported member expression type %T, accessing to member '%s'.\n", left, member)
// 	return env.SymbolTypeInvalid
// }

func (c *Checker) inferIfBlockStatement(block *ast.BlockStatement) env.ChlangType {
	c.Env.OpenScope()
	c.populateSymbolDeclarations(block.Statements)
	var returnType env.ChlangType = env.SymbolTypeVoid
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
	c.Env.CloseScope()
	return returnType
}

func (c *Checker) inferIntLiteral(node *ast.IntLiteral) (env.ChlangPrimitiveType, int64) {
	bitSize := 64
	intValue, err := strconv.ParseInt(node.Value, 0, bitSize)
	if err != nil {
		c.Errors = append(c.Errors, &errors.SemanticError{
			Message:  fmt.Sprintf("value '%s' is out of integer %d-bit range", node.Value, bitSize),
			Position: node.Span.Start,
		})
		return env.SymbolTypeInvalid, 0
	}

	if intValue >= -128 && intValue <= 127 {
		return env.SymbolTypeInt8, intValue
	}

	if intValue >= -32768 && intValue <= 32767 {
		return env.SymbolTypeInt16, intValue
	}

	if intValue >= -2147483648 && intValue <= 2147483647 {
		return env.SymbolTypeInt32, intValue
	}

	if intValue >= -9223372036854775808 && intValue <= 9223372036854775807 {
		return env.SymbolTypeInt64, intValue
	}

	return env.SymbolTypeInvalid, intValue
}

func (c *Checker) checkIntLiteralSuffix(node *ast.IntLiteral) env.ChlangPrimitiveType {
	mode := node.Suffix[0]
	bitSize, err := strconv.Atoi(node.Suffix[1:])

	if mode == 'f' {
		if _, err := strconv.ParseFloat(node.Value, bitSize); err != nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("value '%s' is out of float %d-bit range", node.Value, bitSize),
				Position: node.Span.Start,
			})
			return env.SymbolTypeInvalid
		}

		if bitSize == 32 {
			return env.SymbolTypeFloat32
		}

		return env.SymbolTypeFloat64
	} else if mode == 'u' { // unsigned int
		if err != nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("invalid integer type suffix '%s'", node.Suffix),
				Position: node.Span.Start,
			})
			return env.SymbolTypeInvalid
		}

		uintValue, err := strconv.ParseUint(node.Value, 0, bitSize)
		if err != nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("value '%s' is out of unsigned %d-bit range", node.Value, bitSize),
				Position: node.Span.Start,
			})
			return env.SymbolTypeInvalid
		}

		if bitSize == 8 && uintValue <= 255 {
			return env.SymbolTypeUint8
		}

		if bitSize == 16 && uintValue <= 65535 {
			return env.SymbolTypeUint16
		}

		if bitSize == 32 && uintValue <= 4294967295 {
			return env.SymbolTypeUint32
		}

		if bitSize == 64 && uintValue <= 18446744073709551615 {
			return env.SymbolTypeUint64
		}
	} else if mode == 'i' {
		intValue, err := strconv.ParseInt(node.Value, 0, bitSize)
		if err != nil {
			c.Errors = append(c.Errors, &errors.SemanticError{
				Message:  fmt.Sprintf("value '%s' is out of integer %d-bit range", node.Value, bitSize),
				Position: node.Span.Start,
			})
			return env.SymbolTypeInvalid
		}

		if bitSize == 8 && intValue >= -128 && intValue <= 127 {
			return env.SymbolTypeInt8
		}

		if bitSize == 16 && intValue >= -32768 && intValue <= 32767 {
			return env.SymbolTypeInt16
		}

		if bitSize == 32 && intValue >= -2147483648 && intValue <= 2147483647 {
			return env.SymbolTypeInt32
		}

		if bitSize == 64 && intValue >= -9223372036854775808 && intValue <= 9223372036854775807 {
			return env.SymbolTypeInt64
		}
	}

	c.Errors = append(c.Errors, &errors.SemanticError{
		Message:  fmt.Sprintf("unknown integer type suffix '%s'", node.Suffix),
		Position: node.Span.Start,
	})

	return env.SymbolTypeInvalid
}

func (c *Checker) checkTypesCompatibility(a, b env.ChlangType, operator *chToken.Token) (env.ChlangType, error) {
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
		leftType, leftIsPrimitive := a.(env.ChlangPrimitiveType)
		rightType, rightIsPrimitive := b.(env.ChlangPrimitiveType)

		if leftIsPrimitive && rightIsPrimitive && leftType.IsNumeric() && rightType.IsNumeric() {
			if a == b {
				return a, nil
			}
			if leftType.IsFloat() || rightType.IsFloat() {
				return env.SymbolTypeFloat64, nil
			}

			if leftType.IsSigned() && rightType.IsSigned() {
				return env.GetMaxType(leftType, rightType), nil
			}

			if leftType.IsUnsigned() && rightType.IsUnsigned() {
				return env.GetMaxType(leftType, rightType), nil
			}
		}

		return env.SymbolTypeInvalid, &errors.SemanticError{
			Message:  fmt.Sprintf("type mismatch: incompatible types (left: %s, right: %s) for operator '%s'", a, b, operator.Literal),
			Position: operator.Position,
			HelpMsg:  fmt.Sprintf("got '%s' and '%s'", a, b),
		}
	case chToken.EQUALS, chToken.NOT_EQUALS, chToken.LESS,
		chToken.LESS_EQUALS, chToken.GREATER, chToken.GREATER_EQUALS, chToken.AND, chToken.OR:
		if env.IsCompatibleType(a, b) {
			return env.SymbolTypeBool, nil
		}
		return env.SymbolTypeInvalid, fmt.Errorf("type mismatch: operator '%s' requires operands of the same type (left: %s, right: %s)", operator.Literal, a, b)
	}
	return env.SymbolTypeInvalid, fmt.Errorf("type mismatch: unknown operator: %s", operator.Literal)
}

func (c *Checker) getMaxTypeOf(left, right env.ChlangType) env.ChlangType {
	leftType, leftIsPrimitive := left.(env.ChlangPrimitiveType)
	rightType, rightIsPrimitive := right.(env.ChlangPrimitiveType)

	if leftIsPrimitive && rightIsPrimitive {
		return env.GetMaxType(leftType, rightType)
	}

	fmt.Printf("[warn]: getMaxTypeOf: unsupported types: %s, %s\n", left, right)
	return left
}

func (c *Checker) reportError(message string, span *chToken.Span) {
	c.Errors = append(c.Errors, &errors.SemanticError{
		Message:  message,
		Span:     span,
		Position: span.Start,
	})
}
