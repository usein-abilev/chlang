package vm

// Opcode represents the operation code of an instruction. Maximum number of opcodes is 256 (0-255)
type Opcode uint8

const (
	// OpcodeHalt the execution of the virtual machine
	OpcodeHalt Opcode = iota

	// Moves the value of register R(y) to register R(x). Order like in x86 assembly
	OpcodeMove // MOV R(x), R(y)

	// Loads constant to register R(x)
	OpcodeLoadConst

	// Loads boolean value to register R(x)
	OpcodeLoadBool

	// Loads string to register R(x)
	OpcodeLoadString

	// Loads immediate 32-bit signed integer to register R(x)
	OpcodeLoadImm32 // R(x) = y, LoadImm4 x y

	// Adds two registers and stores the result in register R(x)
	OpcodeAdd // R(x) = R(y) + R(z), AddInt4 x y z

	// Adds 32-bit signed immediate value and register value and stores the result in register R(X)
	OpcodeAddImm32 // R(x) = R(y) + z

	// Subtracts two integers and stores the result in register R(x)
	OpcodeSub // R(x) = R(y) - R(z), SubInt4 x y z

	// Multiplies two integers and stores the result in register R(x)
	OpcodeMul // R(x) = R(y) * R(z), MulInt4 x y z

	// Divides two integers and stores the result in register R(x)
	OpcodeDiv // R(x) = R(y) / R(z), DivInt4 x y z

	// Performs bitwise AND operation
	OpcodeBitwiseAnd // R(x) = R(y) & R(z)

	// Performs bitwise OR operation
	OpcodeBitwiseOr // R(x) = R(y) | R(z)

	// Performs equality check
	OpcodeEq // R(x) = R(y) == R(z)

	// Performs greater than operation
	OpcodeGt // R(x) = R(y) > R(z)

	// Greater than or equal
	OpcodeGte // R(x) = R(y) >= R(z)

	// Less than
	OpcodeLt // R(x) = R(y) < R(z)

	// Less than or equal
	OpcodeLte // R(x) = R(y) <= R(z)

	// Inverts the boolean value of a register
	OpcodeNot // R(x) = !R(y)

	// Jump to a specific address
	OpcodeJump // Jump [address]

	// Conditional jump
	OpcodeJumpIf // JumpIf R(x), #bool, [address]

	// OpcodeCall a function. This instruction accepts a 3 operand: address (function reference stored in register), number of arguments, and number of return values
	OpcodeCall

	// OpcodeReturn from a function
	OpcodeReturn

	// No operation
	OpcodeNop
)

var opcodeNames = map[Opcode]string{
	OpcodeMove:       "Move",
	OpcodeLoadImm32:  "LoadImm32",
	OpcodeLoadBool:   "LoadBool",
	OpcodeLoadConst:  "LoadConst",
	OpcodeLoadString: "LoadString",
	OpcodeAdd:        "Add",
	OpcodeAddImm32:   "AddImm4",
	OpcodeSub:        "Sub",
	OpcodeMul:        "Mul",
	OpcodeDiv:        "Div",
	OpcodeBitwiseAnd: "And",
	OpcodeBitwiseOr:  "Or",
	OpcodeEq:         "Eq",
	OpcodeGt:         "Gt",
	OpcodeGte:        "Gte",
	OpcodeLt:         "Lt",
	OpcodeLte:        "Lte",
	OpcodeNot:        "Not",
	OpcodeJump:       "Jump",
	OpcodeJumpIf:     "JumpIf",
	OpcodeCall:       "Call",
	OpcodeReturn:     "Return",
	OpcodeHalt:       "Halt",
	OpcodeNop:        "Nop",
}

func (op Opcode) String() string {
	return opcodeNames[op]
}
