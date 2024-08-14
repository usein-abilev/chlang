package vm

// Opcode represents the operation code of an instruction. Maximum number of opcodes is 256 (0-255)
type Opcode uint8

const (
	// Halt the execution of the virtual machine
	Halt Opcode = iota

	// Moves the value of register R(y) to register R(x). Order like in x86 assembly
	Move // MOV R(x), R(y)

	// Loads immediate 32-bit signed integer to register R(x)
	LoadImm // R(x) = y, LoadImm4 x y

	// Adds two registers and stores the result in register R(x)
	Add // R(x) = R(y) + R(z), AddInt4 x y z

	// Adds 32-bit signed immediate value and register value and stores the result in register R(X)
	AddImm // R(x) = R(y) + z

	// Subtracts two integers and stores the result in register R(x)
	Sub // R(x) = R(y) - R(z), SubInt4 x y z

	// Multiplies two integers and stores the result in register R(x)
	Mul // R(x) = R(y) * R(z), MulInt4 x y z

	// Divides two integers and stores the result in register R(x)
	Div // R(x) = R(y) / R(z), DivInt4 x y z

	// Call a function. This instruction accepts a 3 operand: address (index of instruction to moved to), number of arguments, and number of return values
	Call

	// Return from a function
	Ret
)

var opcodeArgs = map[Opcode][]uint8{
	Halt:    {},
	Move:    {8, 8},     // R(x), R(y)
	LoadImm: {8, 32},    // R(x), y
	Add:     {8, 8, 8},  // R(x), R(y), R(z)
	AddImm:  {8, 8, 32}, // R(x), R(y), z
	Sub:     {8, 8, 8},  // R(x), R(y), R(z)
	Mul:     {8, 8, 8},  // R(x), R(y), R(z)
	Div:     {8, 8, 8},  // R(x), R(y), R(z)
	Call:    {32, 8, 8}, // Call(func, argCount, returnCount)
}

var opcodeNames = map[Opcode]string{
	Move:    "Move",
	LoadImm: "LoadImm4",
	Add:     "Add",
	AddImm:  "AddImm4",
	Sub:     "Sub",
	Mul:     "Mul",
	Div:     "Div",
	Call:    "Call",
	Ret:     "Ret",
	Halt:    "Halt",
}

func (op Opcode) String() string {
	return opcodeNames[op]
}

func (op Opcode) Args() []uint8 {
	return opcodeArgs[op]
}

// // Creates a new instruction (Little Endian order)
// func (op Opcode) Build(operands ...int32) VMInstruction {
// 	args := op.Args()
// 	instruction := VMInstruction(op)
// 	for i, arg := range args {
// 		instruction |= VMInstruction(operands[i]) << (arg + 8*uint8(i))
// 	}
// 	fmt.Printf(
// 		"Build Instruction: (%b %b %b %b) %v\n",
// 		instruction&0xff,
// 		(instruction>>8)&0xff,
// 		(instruction>>16)&0xff,
// 		(instruction>>24)&0xff,
// 		operands,
// 	)
// 	return instruction
// }
