package vm

import (
	"fmt"
	"strconv"

	"github.com/usein-abilev/chlang/frontend/ast"
	"github.com/usein-abilev/chlang/frontend/ast/symbols"
	"github.com/usein-abilev/chlang/frontend/token"
)

type ForLoopContext struct {
	conditionAddress  int
	endBranches       []int // jump instructions to the end of loop
	conditionBranches []int // jump instructions to the condition
	parent            *ForLoopContext
}

// The codegen package generates the assembly code for the RVM (register based VM).
// It takes the AST generated by the frontend package for input and process compiling stage.
// But AST representation should replaced with the intermediate representation (IR) of the program in the future.
type RVMGenerator struct {
	program *ast.Program

	// current allocator context
	function   *FunctionObject
	forContext *ForLoopContext

	lastBlockExpressionRegister RegisterAddress
}

var mappedBinaryOperatorsToOpcodes = map[token.TokenType]Opcode{
	token.PLUS:           OpcodeAdd,
	token.MINUS:          OpcodeSub,
	token.ASTERISK:       OpcodeMul,
	token.EXPONENT:       OpcodePow,
	token.PERCENT:        OpcodeMod,
	token.SLASH:          OpcodeDiv,
	token.EQUALS:         OpcodeEq,
	token.GREATER:        OpcodeGt,
	token.GREATER_EQUALS: OpcodeGte,
	token.LESS:           OpcodeLt,
	token.LESS_EQUALS:    OpcodeLte,
	token.NOT_EQUALS:     OpcodeNeq,
	token.AND:            OpcodeAnd,
	token.AMPERSAND:      OpcodeAnd,
	token.OR:             OpcodeOr,
	token.PIPE:           OpcodeOr,
	token.CARET:          OpcodeXor,
	token.LEFT_SHIFT:     OpcodeShl,
	token.RIGHT_SHIFT:    OpcodeShr,
}

var mappedAssignOperatorsToOpcodes = map[token.TokenType]Opcode{
	token.ASSIGN:             OpcodeMove,
	token.PLUS_ASSIGN:        OpcodeAdd,
	token.MINUS_ASSIGN:       OpcodeSub,
	token.SLASH_ASSIGN:       OpcodeDiv,
	token.EXPONENT_ASSIGN:    OpcodePow,
	token.ASTERISK_ASSIGN:    OpcodeMul,
	token.PERCENT_ASSIGN:     OpcodeMod,
	token.AMPERSAND_ASSIGN:   OpcodeAnd,
	token.PIPE_ASSIGN:        OpcodeOr,
	token.CARET_ASSIGN:       OpcodeXor,
	token.LEFT_SHIFT_ASSIGN:  OpcodeShl,
	token.RIGHT_SHIFT_ASSIGN: OpcodeShr,
}

func NewRVMGenerator(program *ast.Program) *RVMGenerator {
	moduleFunction := &FunctionObject{
		name:         "<module>",
		instructions: []VMInstruction{},
		locals:       []LocalRegister{},
		constants:    []ConstantValue{},
	}

	// add build-in functions
	for name := range BuildInFunctions {
		moduleFunction.addConstant(name, &OperandValue{
			Kind:  OperandTypeBuildInFunction,
			Value: name,
		})
	}

	return &RVMGenerator{
		program:  program,
		function: moduleFunction,
	}
}

func (g *RVMGenerator) Generate() *FunctionObject {
	for _, statement := range g.program.Statements {
		g.emitStatement(statement)
	}
	g.function.Print()
	return g.function
}

func (g *RVMGenerator) emitStatement(statement ast.Statement) {
	switch statement := statement.(type) {
	case *ast.ConstDeclarationStatement:
		value := getOperandValueFromConstant(statement.Value)
		g.function.addConstant(statement.Name.Value, value)
	case *ast.VarDeclarationStatement:
		g.visitVarDeclaration(statement)
	case *ast.FuncDeclarationStatement:
		g.visitFuncDeclaration(statement)
	case *ast.ForRangeStatement:
		g.function.enterScope()

		// prologue
		loopVar := g.function.addLocal(statement.Identifier.Value)
		startReg := g.emitExpressionAligned(statement.Range.Start)
		g.function.emit(OpcodeMove, loopVar, startReg)

		// condition
		endReg := g.emitExpression(statement.Range.End)
		g.function.bindLocal(endReg, "<for_loop_range_end>")

		condReg := g.function.addTemp()
		var conditionAddress int
		if statement.Range.Inclusive {
			conditionAddress = g.function.emit(OpcodeLte, condReg, loopVar, endReg)
		} else {
			conditionAddress = g.function.emit(OpcodeLt, condReg, loopVar, endReg)
		}
		falseBranch := g.function.emit(OpcodeJumpIf)
		g.function.popTempRegister() // free condition register

		g.forContext = &ForLoopContext{
			conditionAddress:  conditionAddress,
			endBranches:       []int{},
			conditionBranches: []int{},
			parent:            g.forContext,
		}

		for _, statement := range statement.Body.Statements {
			g.emitStatement(statement)
		}

		// incrementing for variable and jumping to condition
		oneReg := g.function.addTemp()
		g.function.popTempRegister() // TODO: Need to optimize these calls by merging 'addTemp' and 'popTempRegister' into one method
		incrementAddr := g.function.emit(OpcodeLoadConst, oneReg, g.function.emitConstantValue(
			&OperandValue{
				Kind:  OperandTypeInt64,
				Value: int64(1),
			}),
		)
		g.function.emit(OpcodeAdd, loopVar, loopVar, oneReg)
		g.function.emit(OpcodeJump, conditionAddress)

		for _, instruction := range g.forContext.conditionBranches {
			g.function.PatchInstruction(instruction, incrementAddr)
		}

		// patch all instructions that should jump to the end of loop
		endLoopAddress := len(g.function.instructions)
		g.function.PatchInstruction(falseBranch, condReg, false, endLoopAddress)
		for _, instruction := range g.forContext.endBranches {
			g.function.PatchInstruction(instruction, endLoopAddress)
		}
		g.forContext = g.forContext.parent

		// g.function.Print()
		g.function.leaveScope()
	case *ast.BreakStatement:
		if g.forContext == nil {
			panic("break statement outside of loop")
		}
		g.forContext.endBranches = append(g.forContext.endBranches, g.function.emit(OpcodeJump))
	case *ast.ContinueStatement:
		if g.forContext == nil {
			panic("continue statement outside of loop")
		}
		g.forContext.conditionBranches = append(g.forContext.conditionBranches, g.function.emit(OpcodeJump))
	case *ast.ReturnStatement:
		returnRegister := g.emitExpressionAligned(statement.Expression)
		g.function.emit(OpcodeReturn, returnRegister, 1)
	case *ast.ExpressionStatement:
		g.lastBlockExpressionRegister = g.emitExpressionAligned(statement.Expression)
	case *ast.BlockStatement:
		g.function.enterScope()
		for _, statement := range statement.Statements {
			g.emitStatement(statement)
		}
		g.function.leaveScope()
	default:
		panic(fmt.Sprintf("unknown statement type '%T'", statement))
	}
}

func (g *RVMGenerator) visitVarDeclaration(decl *ast.VarDeclarationStatement) {
	if decl.Value == nil {
		g.function.addLocal(decl.Name.Value)
		return
	}
	leftRegister := g.emitExpressionAligned(decl.Value)
	bound := g.function.bindLocal(leftRegister, decl.Name.Value)
	if !bound {
		// if the result of the expression is variable's register, then it should be moved to another register
		// for example, the expression 'x = y' produces two variable registers r(x) and r(y),
		// so binding register r(x) = r(y) will override each other and we lose access to variable 'y'
		registerId := g.function.addLocal(decl.Name.Value)
		if leftRegister != registerId {
			g.function.emit(OpcodeMove, registerId, leftRegister)
		}
	}
}

func (g *RVMGenerator) visitFuncDeclaration(decl *ast.FuncDeclarationStatement) {
	parentFunction := g.function
	g.function = &FunctionObject{
		name:         decl.Name.Value,
		parent:       parentFunction,
		instructions: []VMInstruction{},
		locals:       []LocalRegister{},
		constants:    []ConstantValue{},
		scopeDepth:   0,
	}
	parentFunction.addConstant(decl.Name.Value, &OperandValue{
		Kind:  OperandTypeFunctionObject,
		Value: g.function,
	})

	for _, argument := range decl.Params {
		g.function.addLocal(argument.Name.Value)
	}

	for _, bodyStatement := range decl.Body.Statements {
		g.emitStatement(bodyStatement)
	}

	g.function.emit(OpcodeReturn, RegisterAddress(0), 0) // emit default return statement at the end to prevent missing return statement
	g.function = parentFunction
}

func (g *RVMGenerator) emitExpressionAligned(expression ast.Expression) RegisterAddress {
	register := g.emitExpression(expression)
	g.function.freeAllTempRegister()
	return register
}

func (g *RVMGenerator) emitExpression(expression ast.Expression) RegisterAddress {
	switch expr := expression.(type) {
	case *ast.IfExpression:
		condRegister := g.emitExpressionAligned(expr.Condition)
		falseBranch := g.function.emit(OpcodeJumpIf)
		resultRegister := g.function.addTemp()

		g.lastBlockExpressionRegister = -1
		g.emitStatement(expr.ThenBlock)
		if g.lastBlockExpressionRegister != -1 {
			g.function.emit(OpcodeMove, resultRegister, g.lastBlockExpressionRegister)
		}

		thenBranch := g.function.emit(OpcodeJump)
		g.function.PatchInstruction(falseBranch, condRegister, false, len(g.function.instructions))
		if expr.ElseBlock != nil {
			switch expr.ElseBlock.(type) {
			case *ast.BlockStatement:
				g.lastBlockExpressionRegister = -1
				g.emitStatement(expr.ElseBlock)
				if g.lastBlockExpressionRegister != -1 {
					g.function.emit(OpcodeMove, resultRegister, g.lastBlockExpressionRegister)
				}
			case *ast.IfExpression:
				g.emitExpressionAligned(expr.ElseBlock)
			}
		}
		g.function.PatchInstruction(thenBranch, len(g.function.instructions))
		return resultRegister
	case *ast.UnaryExpression:
		targetReg := g.function.addTemp()
		operandReg := g.emitExpression(expr.Right)
		switch expr.Operator.Type {
		case token.BANG:
			g.function.emit(OpcodeNot, targetReg, operandReg)
		case token.MINUS:
			g.function.emit(OpcodeNeg, targetReg, operandReg)
		case token.PLUS:
		default:
			panic(fmt.Sprintf("error: unknown unary operator '%s': %s", expr.Operator.Literal, expr.Span))
		}
		return targetReg
	case *ast.CallExpression:
		calleeReg := g.function.addTemp() // callee register also can be as a return register

		functionRef := g.function.lookupConstant(expr.Function.Symbol.Name)
		if functionRef == nil {
			panic(fmt.Sprintf("error: unresolved function '%s'", expr.Function.Symbol.Name))
		}
		g.function.emit(OpcodeLoadConst, calleeReg, g.function.emitConstantValue(functionRef))

		for _, argumentExpr := range expr.Args {
			register := g.emitExpression(argumentExpr)
			if int(register) < len(g.function.locals) && !g.function.locals[register].temp {
				tempRegister := g.function.addTemp()
				fmt.Printf("emit temporary register? %d\n", register)
				g.function.emit(OpcodeMove, tempRegister, register)
			}
		}

		returns := 0
		if expr.Function.Symbol.Function.ReturnType != symbols.SymbolTypeVoid {
			returns = 1
		}
		g.function.emit(OpcodeCall, calleeReg, len(expr.Args), returns)

		for i := 0; i < len(expr.Args); i++ {
			g.function.popTempRegister()
		}

		return calleeReg
	case *ast.AssignExpression:
		leftReg := g.emitExpression(expr.Left)
		if opcode, ok := mappedAssignOperatorsToOpcodes[expr.Operator.Type]; ok {
			if expr.Operator.Type == token.ASSIGN {
				rightReg := g.emitExpression(expr.Right)
				g.function.emit(OpcodeMove, leftReg, rightReg)
				return leftReg
			}
			rightReg := g.emitExpression(expr.Right)
			g.function.emit(opcode, leftReg, leftReg, rightReg)
			return leftReg
		}

		panic(fmt.Sprintf("unknown assign expression operator '%s'", expr.Operator.Literal))
	case *ast.BinaryExpression:
		targetReg := g.function.addTemp()

		leftReg := g.emitExpression(expr.Left)
		rightReg := g.emitExpression(expr.Right)

		g.function.popTempRegister() // pop last register

		if opcode, ok := mappedBinaryOperatorsToOpcodes[expr.Operator.Type]; ok {
			g.function.emit(opcode, targetReg, leftReg, rightReg)
			return targetReg
		}

		panic(fmt.Sprintf("error: unknown operator '%s'", expr.Operator.Literal))
	case *ast.IntLiteral:
		targetReg := g.function.addTemp()
		g.function.emit(OpcodeLoadConst, targetReg, g.function.emitConstantValue(getOperandValueFromConstant(expr)))
		return targetReg
	case *ast.FloatLiteral:
		targetReg := g.function.addTemp()
		g.function.emit(OpcodeLoadConst, targetReg, g.function.emitConstantValue(getOperandValueFromConstant(expr)))
		return targetReg
	case *ast.BoolLiteral:
		var value bool
		switch expr.Value {
		case "true":
			value = true
		case "false":
			value = false
		default:
			panic(fmt.Sprintf("error: invalid boolean value: %s", expr.Value))
		}
		reg := g.function.addTemp()
		g.function.emit(OpcodeLoadBool, reg, value)
		return reg
	case *ast.StringLiteral:
		reg := g.function.addTemp()
		g.function.emit(OpcodeLoadString, reg, expr.Value)
		return reg
	case *ast.Identifier:
		local := g.function.lookupLocal(expr.Value)
		if local == nil {
			if constant := g.function.lookupConstant(expr.Value); constant == nil {
				panic(fmt.Sprintf("error: unresolved symbol '%s' at %s\n", expr.Value, expr.Token.Position))
			}
			registerId := g.function.addTemp()
			g.function.emit(OpcodeLoadConst, registerId, expr.Value)
			return registerId
		} else {
			return local.address
		}
	}

	panic(fmt.Sprintf("error: unknown expression type: %T", expression))
}

func getOperandValueFromConstant(expr ast.Expression) *OperandValue {
	switch expr := expr.(type) {
	case *ast.IntLiteral:
		value, err := strconv.ParseInt(expr.Value, expr.Base, 64)
		if err != nil {
			panic("getOperandValueFromConstant: invalid int literal")
		}
		return &OperandValue{
			Kind:  OperandTypeInt64,
			Value: value,
		}
	case *ast.FloatLiteral:
		value, err := strconv.ParseFloat(expr.Value, 64)
		if err != nil {
			panic("getOperandValueFromConstant: invalid float literal")
		}
		return &OperandValue{
			Kind:  OperandTypeFloat64,
			Value: value,
		}
	case *ast.BoolLiteral:
		return &OperandValue{
			Kind:  OperandTypeBool,
			Value: expr.Value == "true",
		}
	case *ast.StringLiteral:
		return &OperandValue{
			Kind:  OperandTypeString,
			Value: expr.Value,
		}
	}

	panic("getOperandValueFromConstant: unknown expression type")
}
