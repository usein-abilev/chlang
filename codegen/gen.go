package codegen

import (
	"fmt"
	"strconv"

	"github.com/usein-abilev/chlang/frontend/ast"
	"github.com/usein-abilev/chlang/frontend/token"
	"github.com/usein-abilev/chlang/vm"
)

type generator struct {
	program *ast.Program
	builder *vm.ASMBuilder

	// Register allocation
	mappedRegisters map[string]int
	registers       map[int]bool // true means 'busy'
}

func GenerateVMAssembly(program *ast.Program, builder *vm.ASMBuilder) {
	gen := &generator{
		program:         program,
		builder:         builder,
		mappedRegisters: make(map[string]int),
		registers:       make(map[int]bool),
	}
	for _, statement := range program.Statements {
		gen.genStatement(statement)
	}
	builder.Add(vm.Halt)
}

func (g *generator) genStatement(statement ast.Statement) {
	switch stmt := statement.(type) {
	case *ast.VarDeclarationStatement:
		varRegIndex := g.bindRegister(stmt.Name.Value)
		resultReg := g.generateExpression(stmt.Value)
		g.builder.Add(vm.Move, varRegIndex, resultReg)
		g.freeRegister(resultReg)
	case *ast.FuncDeclarationStatement:
		g.builder.Label(stmt.Name.Value)
		// How to handle function arguments in register-based vm?
		// Should we use the stack?

		// Generate function body
		for _, statement := range stmt.Body.Statements {
			g.genStatement(statement)
		}
	}
}

func (g *generator) generateExpression(expr ast.Expression) int {
	switch e := expr.(type) {
	case *ast.BinaryExpression:
		leftReg := g.generateExpression(e.Left)
		rightReg := g.generateExpression(e.Right)
		switch e.Operator.Type {
		case token.PLUS:
			g.builder.Add(vm.Add, leftReg, leftReg, rightReg)
		case token.MINUS:
			g.builder.Add(vm.Sub, leftReg, leftReg, rightReg)
		case token.ASTERISK:
			g.builder.Add(vm.Mul, leftReg, leftReg, rightReg)
		case token.SLASH:
			g.builder.Add(vm.Div, leftReg, leftReg, rightReg)
		default:
			fmt.Printf("Unknown operator: %v\n", e.Operator.Type)
		}
		g.freeRegister(rightReg)
		return leftReg
	case *ast.IntLiteral:
		reg := g.allocateRegister()
		value, err := strconv.ParseInt(e.Value, 10, 64)
		if err != nil {
			fmt.Printf("Error parsing int: %v\n", err)
			return -1
		}
		g.builder.Add(vm.LoadImm, reg, value)
		return reg
	case *ast.Identifier:
		if index, ok := g.mappedRegisters[e.Value]; ok {
			return index
		}
		panic("Unresolved identifier")
	case *ast.CallExpression:
		panic("CallExpression not implemented")
	}

	return -1
}

func (g *generator) bindRegister(name string) int {
	if index, ok := g.mappedRegisters[name]; ok {
		return index
	}
	index := g.allocateRegister()
	g.mappedRegisters[name] = index
	g.registers[index] = true
	return index
}

// Allocates a new free register
func (g *generator) allocateRegister() int {
	for index := 0; index < 0xFF; index++ {
		reg := g.registers[index]
		if reg {
			continue
		}
		g.registers[index] = true
		return index
	}

	panic("allocateRegister(): no registers available")
}

// Frees a register to be used by another variable/value
func (g *generator) freeRegister(idx int) {
	g.registers[idx] = false
}
