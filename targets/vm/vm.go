// Represents the 64-bit register-based virtual machine that will run the compiled code.
package vm

import (
	"fmt"
	"math"
)

const (
	minStackFrameSize = 0xff
	maxStackSize      = 0xFFFF
)

type Stack []OperandValue

// A CallFrame struct represents a function call frame on the stack
type CallFrame struct {
	parent        *CallFrame // parent call record (for nested calls)
	function      *FunctionObject
	base          RegisterAddress // base pointer for the current stack frame
	top           RegisterAddress // top pointer for the current stack frame (inclusive)
	returnAddress RegisterAddress // instruction pointer in the parent call record
	args          uint64          // number of arguments
	results       uint64          // number of return values
	usedSize      uint64          // size of the used stack frame
}

type VMInstruction struct {
	opcode   Opcode
	operands []any
}

// A virtual machine's code based on Tree-Address Code like in RISC-V, V8, Lua VM assemblies
type VM struct {
	ip            uint32     // instruction pointer (program counter)
	stack         Stack      // stack of 64-bit values (registers, parameters for function, locals, etc.)
	stackCapacity int        // current stack capacity
	callRecord    *CallFrame // current call record
	options       *VMOptions // VM options (debug, etc.)
}

type VMOptions struct {
	Debug bool
}

func NewVM(module *FunctionObject, opts *VMOptions) *VM {
	stack := make(Stack, minStackFrameSize)
	record := &CallFrame{
		base:     0,
		top:      minStackFrameSize,
		function: module,
	}
	vm := &VM{
		ip:            0,
		stack:         stack,
		stackCapacity: minStackFrameSize,
		callRecord:    record,
		options:       opts,
	}
	return vm
}

func (vm *VM) debug(format string, args ...interface{}) {
	if vm.options.Debug {
		fmt.Printf(format, args...)
	}
}

func (vm *VM) printStack() {
	fmt.Printf("-----------------------------------\nStack: \n")
	for i, slot := range vm.stack {
		if slot.Kind == OperandTypeUndefined {
			continue
		}
		fmt.Printf("S%d: kind=%s, value=%d", i, slot.Kind, slot.Value)
		idx := RegisterAddress(i)
		if vm.callRecord.parent != nil {
			if idx == vm.callRecord.parent.base {
				fmt.Printf("  <- parent base")
			}
			if idx == vm.callRecord.parent.top-1 {
				fmt.Printf("  <- parent top")
			}
		}
		if idx == vm.callRecord.base {
			fmt.Printf("  <- base")
		}
		if idx == vm.callRecord.top-1 {
			fmt.Printf("  <- top")
		}
		fmt.Printf("\n")
	}
	fmt.Printf("-----------------------------------\n")
}

func (vm *VM) Run() {
main_loop:
	for {
		base := RegisterAddress(vm.callRecord.base)
		instruction := vm.fetch()
		if instruction == nil {
			break
		}
		opcode, operands := instruction.opcode, instruction.operands
		switch opcode {
		case OpcodeHalt:
			break main_loop
		case OpcodeLoadImm32:
			address := operands[0].(RegisterAddress)
			operand := operands[1].(int64) // TODO: For now, keep it as int64
			vm.setStackValue(base+address, &OperandValue{
				Kind:  OperandTypeInt64,
				Value: operand,
			})
		case OpcodeLoadBool:
			address := operands[0].(RegisterAddress)
			operand := operands[1].(bool)
			vm.setStackValue(base+address, &OperandValue{
				Kind:  OperandTypeBool,
				Value: operand,
			})
		case OpcodeLoadConst:
			target := operands[0].(RegisterAddress)
			value := operands[1].(ConstantValueIdx)
			constant := vm.callRecord.function.constants[value]
			vm.setStackValue(base+target, constant.Value)
		case OpcodeLoadString:
			target := operands[0].(RegisterAddress)
			value := operands[1].(string)
			vm.setStackValue(base+target, &OperandValue{
				Kind:  OperandTypeString,
				Value: value,
			})
		case OpcodeNot:
			target := operands[0].(RegisterAddress)
			operand := operands[1].(RegisterAddress)
			vm.setStackValue(base+target, &OperandValue{
				Kind:  OperandTypeBool,
				Value: !vm.stack[base+operand].Value.(bool),
			})
		case OpcodeNeg:
			target := operands[0].(RegisterAddress)
			operand := operands[1].(RegisterAddress)
			stackValue := vm.stack[base+operand]
			if !stackValue.Kind.IsNumeric() {
				panic(fmt.Sprintf("vm: invalid operand type '%s' for negation", vm.stack[base+operand].Kind))
			}
			switch stackValue.Kind {
			case OperandTypeInt8, OperandTypeInt16, OperandTypeInt32, OperandTypeInt64:
				vm.setStackValue(base+target, &OperandValue{
					Kind:  OperandTypeInt64,
					Value: -stackValue.Value.(int64),
				})
			case OperandTypeFloat32, OperandTypeFloat64:
				vm.setStackValue(base+target, &OperandValue{
					Kind:  OperandTypeFloat64,
					Value: -stackValue.Value.(float64),
				})
			}
		case OpcodeAdd, OpcodeSub, OpcodeMul, OpcodeDiv,
			OpcodeEq, OpcodeGt, OpcodeGte, OpcodeLt, OpcodeLte, OpcodeNeq,
			OpcodeXor, OpcodeAnd, OpcodeOr, OpcodeShl, OpcodeShr,
			OpcodeMod, OpcodePow:
			target := operands[0].(RegisterAddress)
			operand1 := operands[1].(RegisterAddress)
			operand2 := operands[2].(RegisterAddress)
			vm.performBinaryOperation(opcode, base+target, base+operand1, base+operand2)
		case OpcodeMove:
			dest := operands[0].(RegisterAddress)
			source := operands[1].(RegisterAddress)
			sourceSlot := vm.stack[base+source]
			vm.setStackValue(base+dest, &OperandValue{
				Kind:  sourceSlot.Kind,
				Value: sourceSlot.Value,
			})
		case OpcodeJump:
			address := operands[0].(int)
			vm.ip = uint32(address)
		case OpcodeJumpIf:
			value := operands[0].(RegisterAddress)
			condition := operands[1].(bool)
			address := operands[2].(int)
			if vm.stack[base+value].Value == condition {
				vm.ip = uint32(address)
			}
		case OpcodeCall:
			function := operands[0].(RegisterAddress)
			args := operands[1].(int)
			results := operands[2].(int)
			vm.callFunc(function, args, results)
		case OpcodeReturn:
			from := operands[0].(RegisterAddress)
			count := operands[1].(int)

			// return values are stored after the arguments
			returnStartIdx := vm.callRecord.base - 1
			for i := 0; i < count; i++ {
				vm.setStackValue(returnStartIdx+RegisterAddress(i), &vm.stack[base+from+RegisterAddress(i)])
			}

			vm.ip = uint32(vm.callRecord.returnAddress) // go back to parent
			vm.callRecord = vm.callRecord.parent
		default:
			panic(fmt.Sprintf("error: unknown opcode: %v\n", opcode))
		}
	}

	if vm.options.Debug {
		vm.printStack()
	}
}

func (vm *VM) performBinaryOperation(opcode Opcode, register, x, y RegisterAddress) {
	operandX := vm.stack[x]
	operandY := vm.stack[y]

	if opcode.IsComparison() {
		if operandX.Kind != operandY.Kind {
			panic(fmt.Sprintf("vm: invalid operand type '%s' and '%s'", operandX.Kind, operandY.Kind))
		}

		var result bool

		switch x := operandX.Value.(type) {
		case int64:
			y := operandY.Value.(int64)

			switch opcode {
			case OpcodeEq:
				result = x == y
			case OpcodeGt:
				result = x > y
			case OpcodeGte:
				result = x >= y
			case OpcodeLt:
				result = x < y
			case OpcodeLte:
				result = x <= y
			case OpcodeNeq:
				result = x != y
			}
		case float64:
			y := operandY.Value.(float64)

			switch opcode {
			case OpcodeEq:
				result = x == y
			case OpcodeGt:
				result = x > y
			case OpcodeGte:
				result = x >= y
			case OpcodeLt:
				result = x < y
			case OpcodeLte:
				result = x <= y
			case OpcodeNeq:
				result = x != y
			}

		default:
			panic(fmt.Sprintf("vm: unsupported operand type '%T'", operandX.Value))
		}

		vm.setStackValue(register, &OperandValue{
			Kind:  OperandTypeBool,
			Value: result,
		})
		return
	}

	if operandX.Kind.IsNumeric() {
		var result interface{}

		switch x := operandX.Value.(type) {
		case int64:
			var y int64
			if _, ok := operandY.Value.(int64); !ok {
				y = int64(operandY.Value.(float64))
			} else {
				y = operandY.Value.(int64)
			}

			switch opcode {
			case OpcodeAdd:
				result = x + y
			case OpcodeSub:
				result = x - y
			case OpcodeMul:
				result = x * y
			case OpcodeDiv:
				if y == 0 {
					panic("vm: division by zero")
				}
				result = x / y
			case OpcodePow:
				result = int64(math.Pow(float64(x), float64(y)))
			case OpcodeMod:
				result = x % y
			case OpcodeShl:
				result = x << y
			case OpcodeShr:
				result = x >> y
			case OpcodeAnd:
				result = x & y
			case OpcodeOr:
				result = x | y
			case OpcodeXor:
				result = x ^ y
			default:
				panic(fmt.Sprintf("vm: unknown binary opcode '%s'", opcode))
			}

		case float64:
			var y float64
			if _, ok := operandY.Value.(float64); !ok {
				y = float64(operandY.Value.(int64))
			} else {
				y = operandY.Value.(float64)
			}

			switch opcode {
			case OpcodeAdd:
				result = x + y
			case OpcodeSub:
				result = x - y
			case OpcodeMul:
				result = x * y
			case OpcodeDiv:
				if y == 0.0 {
					panic("vm: division by zero")
				}
				result = x / y
			case OpcodePow:
				result = math.Pow(x, y)
			default:
				panic(fmt.Sprintf("vm: unsupported float64 operation for '%s'", opcode))
			}

		default:
			panic(fmt.Sprintf("vm: unsupported operand type '%T'", operandX.Value))
		}

		// Set the appropriate type based on the result
		var kind OperandValueType
		switch result.(type) {
		case int64:
			kind = OperandTypeInt64
		case float64:
			kind = OperandTypeFloat64
		}

		vm.setStackValue(register, &OperandValue{
			Kind:  kind,
			Value: result,
		})
	} else if operandX.Kind == OperandTypeBool {
		var result, valueX, valueY bool
		valueX = operandX.Value.(bool)
		valueY = operandY.Value.(bool)

		switch opcode {
		case OpcodeAnd:
			result = valueX && valueY
		case OpcodeOr:
			result = valueX || valueY
		default:
			panic(fmt.Sprintf("vm: unknown binary opcode '%s'", opcode))
		}

		vm.setStackValue(register, &OperandValue{
			Kind:  OperandTypeBool,
			Value: result,
		})
	} else {
		panic(fmt.Sprintf("vm: invalid operand type '%s', operand y = '%s', opcode '%s' (current instruction: %d)", operandX.Kind, operandY.Kind, opcode, vm.ip))
	}
}

func (vm *VM) setStackNullValue(index uint64) {
	vm.stack[index] = OperandValue{
		Kind:  OperandTypeUndefined,
		Value: nil,
	}
}

func (vm *VM) setStackValue(index RegisterAddress, slot *OperandValue) {
	idx := uint64(index)
	if idx+1 > vm.callRecord.usedSize {
		vm.callRecord.usedSize = idx + 1
	}
	vm.stack[index] = *slot
}

// Perform a function call. First argument is the address register R(Addr) of the function.
// This address register should contain the address of the function to be called.
// The method will create a new call record where 'base' starts from reference register R(Addr)
// LoadSym 0, "sum"	; Load symbol 'sum' to the register 0
// LoadImm 1, 10	; Load value 10 to the register 1
// LoadImm 2, 20	; Load value 20 to the register 2
// Call 0, 2, 1 	; Call sum function with 2 arguments and 1 return value
func (vm *VM) callFunc(function RegisterAddress, args int, results int) {
	vm.debug("[func call]: Calling a function at address %d (args=%d, rets=%d, used=%d)\n", function, args, results, vm.callRecord.usedSize)
	parentFrame := vm.callRecord
	functionBasePointer := parentFrame.base + function + 1 // base starts at the first argument if exists (addressReg + 1)
	functionObj := vm.stack[parentFrame.base+function]
	if functionObj.Kind == OperandTypeBuildInFunction {
		builtinFunction, ok := BuildInFunctions[functionObj.Value.(string)]
		if !ok {
			panic(fmt.Sprintf("Invalid build-in function '%s'", functionObj.Value))
		}
		var operands []*OperandValue
		for i := 0; i < args; i++ {
			operands = append(operands, &vm.stack[functionBasePointer+RegisterAddress(i)])
		}
		builtinFunction(operands)
		vm.setStackNullValue(uint64(functionBasePointer) - 1)
		return
	} else if functionObj.Kind != OperandTypeFunctionObject {
		panic(fmt.Sprintf("Invalid function object to perform call: %T (ip=%d, caller=%s)", functionObj, vm.ip-1, vm.callRecord.function.name))
	}
	vm.callRecord = &CallFrame{
		args:          uint64(args),
		results:       uint64(results),
		parent:        parentFrame,
		base:          functionBasePointer,
		top:           functionBasePointer + minStackFrameSize,
		function:      functionObj.Value.(*FunctionObject),
		returnAddress: RegisterAddress(vm.ip),
	}
	vm.ip = 0
	vm.debug("[func call]: New function call record: base=%d, top=%d, return=%d\n", vm.callRecord.base, vm.callRecord.top, vm.callRecord.returnAddress)

	// resize the stack if necessary
	if vm.callRecord.top.AsInt() > vm.stackCapacity {
		if vm.stackCapacity > maxStackSize {
			panic(fmt.Sprintf("\033[31mStack overflow during function call '%d'\033[0m", function))
		}
		vm.stackCapacity = int(vm.callRecord.top)
		stack := make(Stack, vm.stackCapacity)
		copy(stack, vm.stack)
		vm.stack = stack
	}
}

func (vm *VM) fetch() *VMInstruction {
	if vm.ip >= uint32(len(vm.callRecord.function.instructions)) {
		return nil
	}
	instruction := vm.callRecord.function.instructions[vm.ip]
	vm.ip++
	return &instruction
}
