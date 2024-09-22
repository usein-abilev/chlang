package symbols

import "fmt"

// Composite type interface
type ChlangType interface {
	Type()
	String() string
}

type ChlangArrayType struct {
	ElementType ChlangType
	Length      int
}

func (ChlangArrayType) Type() {}
func (c ChlangArrayType) String() string {
	element := c.ElementType.String()
	if c.Length > 0 {
		return fmt.Sprintf("%s[%d]", element, c.Length)
	}
	return element + "[]"
}

type ChlangFunctionType struct {
	SpreadType ChlangType // nil if not spread
	Return     ChlangType
	Args       []ChlangType
}

func (c ChlangFunctionType) String() string {
	args := ""
	for i, arg := range c.Args {
		if i > 0 {
			args += ", "
		}
		args += arg.String()
	}
	return "(" + args + ") -> " + c.Return.String()
}
func (ChlangFunctionType) Type() {}

func (ChlangPrimitiveType) Type() {}

// IsLeftCompatibleType checks if the left type is compatible with the right type
// This is used for type checking
func IsLeftCompatibleType(left, right ChlangType) bool {
	if left == right {
		return true
	}

	leftPrimitive, leftIsPrimitive := left.(ChlangPrimitiveType)
	rightPrimitive, rightIsPrimitive := right.(ChlangPrimitiveType)
	if leftIsPrimitive && rightIsPrimitive {
		if (leftPrimitive.IsFloat() && rightPrimitive.IsFloat()) ||
			(leftPrimitive.IsSigned() && rightPrimitive.IsSigned()) ||
			(leftPrimitive.IsUnsigned() && rightPrimitive.IsUnsigned()) {
			return leftPrimitive >= rightPrimitive
		}
	}

	return false
}

// IsCompatibleType checks if the left type is compatible with the right type
func IsCompatibleType(left, right ChlangType) bool {
	if left == right {
		return true
	}

	leftPrimitive, leftIsPrimitive := left.(*ChlangPrimitiveType)
	rightPrimitive, rightIsPrimitive := right.(*ChlangPrimitiveType)
	if leftIsPrimitive && rightIsPrimitive {
		if leftPrimitive.IsNumeric() && rightPrimitive.IsNumeric() {
			return true
		}
	}

	return false
}
