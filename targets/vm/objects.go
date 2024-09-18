package vm

type OperandValueType uint8

const (
	OperandTypeUndefined OperandValueType = iota
	OperandTypeInt8
	OperandTypeInt16
	OperandTypeInt32
	OperandTypeInt64
	OperandTypeFloat32
	OperandTypeFloat64
	OperandTypeBool
	OperandTypeString
	OperandTypeFunctionObject
	OperandTypeBuildInFunction
)

type OperandValue struct {
	Kind  OperandValueType
	Value any
}

func (ovt OperandValueType) IsNumeric() bool {
	switch ovt {
	case OperandTypeInt8, OperandTypeInt16, OperandTypeInt32, OperandTypeInt64, OperandTypeFloat32, OperandTypeFloat64:
		return true
	}
	return false
}

func (ovt OperandValueType) String() string {
	switch ovt {
	case OperandTypeInt8:
		return "int8"
	case OperandTypeInt16:
		return "int16"
	case OperandTypeInt32:
		return "int32"
	case OperandTypeInt64:
		return "int64"
	case OperandTypeFloat32:
		return "float32"
	case OperandTypeFloat64:
		return "float64"
	case OperandTypeBool:
		return "bool"
	case OperandTypeString:
		return "string"
	case OperandTypeFunctionObject:
		return "function"
	case OperandTypeBuildInFunction:
		return "build-in-function"
	}
	return "undefined"
}

type ConstantValueIdx int
type ConstantValue struct {
	Name  string
	Value *OperandValue
}

type RegisterAddress int

func (r RegisterAddress) AsInt() int {
	return int(r)
}

// Represents an allocated local variable register in the VM stack.
type LocalRegister struct {
	name    string
	address RegisterAddress
	depth   int
	temp    bool
}
