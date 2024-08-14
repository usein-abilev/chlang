package vm

import (
	"fmt"
	"strings"
)

type Code struct {
	instructions []VMInstruction
}

type ASMInstruction struct {
	opcode   Opcode
	operands []any
}

type ASMBuilder struct {
	labels         map[string]int32
	resolvedLabels map[int32]string
	instructions   []ASMInstruction
}

func NewASMBuilder() *ASMBuilder {
	return &ASMBuilder{
		labels:         map[string]int32{},
		resolvedLabels: map[int32]string{},
		instructions:   []ASMInstruction{},
	}
}

func (builder *ASMBuilder) Label(label string) {
	idx := int32(len(builder.instructions))
	builder.labels[label] = idx
	builder.resolvedLabels[idx] = label
}

func (builder *ASMBuilder) Add(opcode Opcode, operands ...any) {
	builder.instructions = append(builder.instructions, ASMInstruction{
		opcode:   opcode,
		operands: operands,
	})
}

func (builder *ASMBuilder) resolveOperand(operand any) uint64 {
	if label, ok := operand.(string); ok {
		if addr, ok := builder.labels[label]; ok {
			return uint64(addr)
		} else {
			panic(fmt.Sprintf("error: unknown label: %s", label))
		}
	} else if value, ok := operand.(int); ok {
		return uint64(value)
	} else if value, ok := operand.(int64); ok {
		return uint64(value)
	}
	fmt.Printf("error: unknown operand: %v\n", operand)
	return 0
}

func (builder *ASMBuilder) Print() {
	fmt.Printf("-----------------------------------\nGenerated Code: \n")
	opcodeWidth := 10
	operandWidth := 3
	for i, instruction := range builder.instructions {
		if label, ok := builder.resolvedLabels[int32(i)]; ok {
			fmt.Printf("%v: %v\n", label, strings.Repeat("-", 20))
		}
		fmt.Printf("\t%v: \033[36m%v\033[0m%s", i, instruction.opcode, strings.Repeat(" ", opcodeWidth-len(instruction.opcode.String())))
		for _, operand := range instruction.operands {
			fmt.Printf("%s%v,", strings.Repeat(" ", operandWidth), operand)
		}
		fmt.Println()
	}
	fmt.Printf("-----------------------------------\n")
}

func (builder *ASMBuilder) Build() Code {
	code := Code{}
	for _, asmInstruction := range builder.instructions {
		var operands []uint64
		for _, operand := range asmInstruction.operands {
			operands = append(operands, builder.resolveOperand(operand))
		}
		fmt.Printf("convert Instruction: %v\n", operands)
		code.instructions = append(code.instructions, VMInstruction{
			opcode:   asmInstruction.opcode,
			operands: operands,
		})
	}
	return code
}
