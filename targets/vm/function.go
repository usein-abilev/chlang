package vm

import (
	"fmt"
	"strings"
)

// FunctionObject represents a function or module object in the VM.
// It contains the function's instructions, constants, registers, and parent context.
type FunctionObject struct {
	// Name of the function for debugging purposes.
	name string

	// The list of instructions that are executed by the VM.
	instructions []VMInstruction

	// The list of local registers mapped to the variable names.
	locals []LocalRegister

	// Constants are the values that are used in the function which known at compile time.
	// Like strings, numbers, and functions references.
	constants []ConstantValue

	// The current scope depth of the function.
	scopeDepth int

	// The parent context of the function.
	// If function declared inside another function, the parent is the outer function.
	parent *FunctionObject
}

func (fn *FunctionObject) addConstant(name string, value *OperandValue) ConstantValueIdx {
	if exists := fn.lookupConstant(name); exists != nil {
		panic(fmt.Sprintf("constant '%s' already exists", name))
	}
	fn.constants = append(fn.constants, ConstantValue{
		Name:  name,
		Value: value,
	})
	return ConstantValueIdx(len(fn.constants) - 1)
}

func (fn *FunctionObject) lookupConstant(name string) *OperandValue {
	self := fn
	for self != nil {
		for _, constant := range self.constants {
			if constant.Name == name {
				return constant.Value
			}
		}
		self = self.parent
	}
	return nil
}

func (fn *FunctionObject) emitConstantValue(value *OperandValue) ConstantValueIdx {
	for idx, constant := range fn.constants {
		if constant.Value.Kind == value.Kind && constant.Value.Value == value.Value {
			return ConstantValueIdx(idx)
		}
	}
	fn.constants = append(fn.constants, ConstantValue{
		Name:  fmt.Sprintf("const#%d", len(fn.constants)),
		Value: value,
	})
	return ConstantValueIdx(len(fn.constants) - 1)
}

func (fn *FunctionObject) enterScope() {
	fn.scopeDepth++
}

func (fn *FunctionObject) leaveScope() {
	if len(fn.locals) == 0 {
		fn.scopeDepth--
		return
	}

	// Remove the local registers that are in the current scope.
	idx := len(fn.locals) - 1
	for ; idx >= 0; idx-- {
		if fn.locals[idx].depth != fn.scopeDepth {
			break
		}
	}
	fn.locals = fn.locals[:idx+1]
	fn.scopeDepth--
}

func (fn *FunctionObject) bindLocal(register RegisterAddress, name string) bool {
	if int(register) >= len(fn.locals) {
		return false
	}
	local := &fn.locals[register]
	if !local.temp {
		return false
	}
	local.name = name
	local.temp = false
	return true
}

func (fn *FunctionObject) addTemp() RegisterAddress {
	address := RegisterAddress(len(fn.locals))
	fn.locals = append(fn.locals, LocalRegister{
		name:    fmt.Sprintf("<temp#%d>", len(fn.locals)),
		depth:   fn.scopeDepth,
		temp:    true,
		address: address,
	})
	return address
}

func (fn *FunctionObject) freeAllTempRegister() {
	idx := len(fn.locals) - 1
	for ; idx >= 0; idx-- {
		if !fn.locals[idx].temp {
			break
		}
	}
	fn.locals = fn.locals[:idx+1]
}
func (fn *FunctionObject) popTempRegister() *LocalRegister {
	if len(fn.locals) == 0 {
		return nil
	}
	idx := len(fn.locals) - 1
	last := &fn.locals[idx]
	if !last.temp {
		fmt.Printf("[WARN]: Trying to pop a non-temp register: %v\n", last)
		return nil
	}
	fn.locals = fn.locals[:idx]
	return last
}

func (fn *FunctionObject) addLocal(name string) RegisterAddress {
	address := RegisterAddress(len(fn.locals))
	fn.locals = append(fn.locals, LocalRegister{
		name:    name,
		depth:   fn.scopeDepth,
		temp:    false,
		address: address,
	})
	return address
}

func (fn *FunctionObject) lookupLocal(name string) *LocalRegister {
	for i := len(fn.locals) - 1; i >= 0; i-- {
		if fn.locals[i].name == name {
			return &fn.locals[i]
		}
	}
	return nil
}

func (fn *FunctionObject) emit(opcode Opcode, operands ...any) int {
	fn.instructions = append(fn.instructions, VMInstruction{
		opcode:   opcode,
		operands: operands,
	})
	return len(fn.instructions) - 1
}

func (fn *FunctionObject) PatchInstruction(opcodeAddress int, operands ...any) {
	instruction := fn.instructions[opcodeAddress]
	instruction.operands = operands
	fn.instructions[opcodeAddress] = instruction
}

func (fn *FunctionObject) Print() {
	padding := strings.Repeat("-", 15)
	fmt.Printf("%s function_object=%s %s\n", padding, fn.name, padding)

	fmt.Printf("Constants (%d): \n", len(fn.constants))
	nameOffset := 40
	for i, constant := range fn.constants {
		value := constant.Value.Value
		if constant.Value.Kind == OperandTypeFunctionObject {
			value = fmt.Sprintf("%p", value)
		}
		left := fmt.Sprintf("\t%v: \033[33m<%s>\033[0m%v", i, constant.Value.Kind, value)
		offset := 0
		if nameOffset > len(left) {
			offset = nameOffset - len(left)
		}
		fmt.Printf("%s%s(%s)\n", left, strings.Repeat(" ", offset), constant.Name)
	}

	fn.printLocals()
	fn.printInstructions()
}

func (fn *FunctionObject) printLocals() {
	fmt.Printf("Registers (%d):\n", len(fn.locals))
	for i, local := range fn.locals {
		fmt.Printf("\t%v: %s (temp=%v, scope=%d)\n", i, local.name, local.temp, local.depth)
	}
}

func (fn *FunctionObject) printInstructions() {
	fmt.Printf("Instructions (%d): \n", len(fn.instructions))
	opcodeWidth := 10
	operandWidth := 3
	for i, instruction := range fn.instructions {
		fmt.Printf("\t%v: \033[36m%v\033[0m%s", i, instruction.opcode, strings.Repeat(" ", opcodeWidth-len(instruction.opcode.String())))
		lastOperandIdx := len(instruction.operands) - 1
		for idx, operand := range instruction.operands {
			fmt.Printf("%s", strings.Repeat(" ", operandWidth))
			if _, ok := operand.(RegisterAddress); ok {
				fmt.Printf("\033[33mr%v\033[0m", operand)
			} else if _, ok := operand.(ConstantValueIdx); ok {
				fmt.Printf("\033[34mconst#%v\033[0m", operand)
			} else {
				fmt.Printf("%v", operand)
			}
			if idx != lastOperandIdx {
				fmt.Printf(", ")
			}
		}
		fmt.Println()
	}
}
