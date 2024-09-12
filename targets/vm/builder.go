package vm

import (
	"fmt"
	"strings"
)

type ASMBuilder struct {
	function *FunctionObject
}

func NewASMBuilder(function *FunctionObject) *ASMBuilder {
	return &ASMBuilder{function}
}

func (builder *ASMBuilder) Emit(opcode Opcode, operands ...any) int {
	builder.function.instructions = append(builder.function.instructions, VMInstruction{
		opcode:   opcode,
		operands: operands,
	})
	return len(builder.function.instructions) - 1
}

func (builder *ASMBuilder) PatchInstruction(opcodeAddress int, operands ...any) {
	instruction := builder.function.instructions[opcodeAddress]
	instruction.operands = operands
	builder.function.instructions[opcodeAddress] = instruction
}

func (builder *ASMBuilder) Print() {
	fmt.Printf("-----------------------------------\nGenerated Code of \"%s\": \n", builder.function.name)
	opcodeWidth := 10
	operandWidth := 3
	for i, instruction := range builder.function.instructions {
		fmt.Printf("\t%v: \033[36m%v\033[0m%s", i, instruction.opcode, strings.Repeat(" ", opcodeWidth-len(instruction.opcode.String())))
		lastOperandIdx := len(instruction.operands) - 1
		for idx, operand := range instruction.operands {
			fmt.Printf("%s", strings.Repeat(" ", operandWidth))
			if _, ok := operand.(RegisterAddress); ok {
				fmt.Printf("\033[33mr%v\033[0m", operand)
			} else {
				fmt.Printf("%v", operand)
			}
			if idx != lastOperandIdx {
				fmt.Printf(", ")
			}
		}
		fmt.Println()
	}
	fmt.Printf("-----------------------------------\n")
}
