// Represents the 64-bit register-based virtual machine that will run the compiled code.
package vm

import (
	"fmt"
)

type StackSlotKind int

const (
	StackSlotKindRegister StackSlotKind = iota
	StackSlotKindLocal
	StackSlotKindParameter
	StackSlotKindReturn
)

const (
	minStackFrameSize = 0x7f
	maxStackSize      = 0xFFFF
)

type StackSlot struct {
	kind  StackSlotKind
	value int64 // complete 64-bit value
}
type Stack []StackSlot

type CallRecord struct {
	parent   *CallRecord // parent call record (for nested calls)
	savedIp  uint32      // instruction pointer in the parent call record
	base     uint64      // base pointer for the current stack frame
	top      uint64      // top pointer for the current stack frame (inclusive)
	nargs    uint64      // number of arguments
	nrets    uint64      // number of return values
	usedSize uint64      // size of the used stack frame
}

type VMInstruction struct {
	opcode   Opcode
	operands []uint64
}

type VM struct {
	// Set of instructions to be executed
	// A virtual machine's code based on Tree-Address Code like in RISC-V assemblies
	//
	// Each instruction is represented as unsigned 64-bit integer
	// ISA Table (Little Endian):
	// Opcode = 8 bits
	// | Opcode   | Operand 1 | Operand 2 | Operand 3 |
	// | LoadInt4 |    8bit   |   		48bit         |
	// | Add	  |    8bit   |    8bit    |  8bit    |
	// | Move     |    8bit   |    8bit    |          |
	// | Halt     |           |            |          |
	instructions  []VMInstruction // program instructions (actual vm code)
	ip            uint32          // instruction pointer (program counter)
	stack         Stack           // stack of 64-bit values (registers, parameters for function, locals, etc.)
	stackCapacity int             // current stack capacity
	callRecord    *CallRecord     // current call record
}

func New() *VM {
	stack := make(Stack, minStackFrameSize)
	record := &CallRecord{
		base: 0,
		top:  minStackFrameSize,
	}
	vm := &VM{
		ip:            0,
		stack:         stack,
		stackCapacity: minStackFrameSize,
		callRecord:    record,
		instructions:  []VMInstruction{},
	}
	return vm
}

func (vm *VM) LoadProgram(code Code) {
	vm.instructions = code.instructions
	vm.ip = 0
}

func (vm VM) Print() {
	fmt.Printf("-----------------------------------\nStack: \n")
	for i, slot := range vm.stack {
		fmt.Printf("S%d: kind=%d, value=%d\n", i, slot.kind, slot.value)
	}
	fmt.Printf("-----------------------------------\n")
}

func (vm *VM) Run() {
	for {
		base := vm.callRecord.base
		instruction := vm.fetch()
		if instruction == nil {
			return
		}
		opcode, operands := decode(instruction)
		switch opcode {
		case Halt:
			return
		case LoadImm:
			register, value := operands[0], operands[1]
			vm.setStackValue(base+register, &StackSlot{
				kind:  StackSlotKindLocal,
				value: int64(value),
			})
		case Add:
		case Sub:
		case Mul:
		case Div:
			register, x, y := operands[0], operands[1], operands[2]
			vm.performBinaryOperation(opcode, base+register, base+x, base+y)
		case Move:
			to, from := operands[0], operands[1]
			vm.setStackValue(base+to, &StackSlot{
				kind:  StackSlotKindRegister,
				value: vm.stack[base+from].value,
			})
		case Call:
			address, nargs, nrets := operands[0], operands[1], operands[2]
			vm.callFunc(address, nargs, nrets)
		default:
			fmt.Printf("error: unknown opcode: %v\n", opcode)
		}
	}
}

func (vm *VM) performBinaryOperation(opcode Opcode, register uint64, x uint64, y uint64) {
	var result int64
	switch opcode {
	case Add:
		result = vm.stack[x].value + vm.stack[y].value
	case Sub:
		result = vm.stack[x].value - vm.stack[y].value
	case Mul:
		result = vm.stack[x].value * vm.stack[y].value
	case Div:
		result = vm.stack[x].value / vm.stack[y].value
	}
	vm.setStackValue(register, &StackSlot{
		kind:  StackSlotKindRegister,
		value: result,
	})
}

func (vm *VM) setStackValue(index uint64, slot *StackSlot) {
	if index+1 > vm.callRecord.usedSize {
		vm.callRecord.usedSize = index + 1
	}
	vm.stack[index] = *slot
}

func (vm *VM) callFunc(address uint64, nargs uint64, nrets uint64) {
	if address >= uint64(len(vm.instructions)) {
		panic(fmt.Sprintf("Invalid function address %d", address))
	}
	fmt.Printf("Calling a function at address %d (args=%d, rets=%d, used=%d)\n", address, nargs, nrets, vm.callRecord.usedSize)
	parent := vm.callRecord
	vm.callRecord = &CallRecord{
		savedIp: vm.ip,
		nargs:   nargs,
		nrets:   nrets,
		parent:  parent,
		base:    vm.callRecord.top + 1,
		top:     vm.callRecord.top + 1 + minStackFrameSize,
	}
	// resize the stack if necessary
	if vm.callRecord.top > uint64(vm.stackCapacity) {
		vm.stackCapacity = int(vm.callRecord.top)
		stack := make(Stack, vm.stackCapacity)
		copy(stack, vm.stack)
		vm.stack = stack
	}

	if nargs > 0 {
		argsBegin := parent.base + parent.usedSize - nargs
		for i := uint64(0); i < nargs; i++ {
			offset := argsBegin + i
			fmt.Printf("Coping the argument %d-th to the next stack, value: '%d'\n", i, vm.stack[offset].value)
			vm.setStackValue(vm.callRecord.base+i, &StackSlot{
				kind:  StackSlotKindParameter,
				value: vm.stack[offset].value,
			})
		}
	}

	vm.ip = uint32(address)
}

func (vm *VM) fetch() *VMInstruction {
	if vm.ip >= uint32(len(vm.instructions)) {
		return nil
	}
	instruction := vm.instructions[vm.ip]
	vm.ip++
	return &instruction
}

func decode(instruction *VMInstruction) (Opcode, []uint64) {
	// opcode := Opcode(instruction & 0xff)
	// operands := []int32{}
	// args := opcode.Args()
	// for i, arg := range args {
	// 	operands = append(operands, int32(instruction>>(arg+uint8(8*i)))&0xff)
	// }
	return instruction.opcode, instruction.operands
}
