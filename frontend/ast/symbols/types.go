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

	switch leftType := left.(type) {
	case ChlangPrimitiveType:
		if rightPrimitive, ok := right.(ChlangPrimitiveType); ok {
			if (leftType.IsFloat() && rightPrimitive.IsFloat()) ||
				(leftType.IsSigned() && rightPrimitive.IsSigned()) ||
				(leftType.IsUnsigned() && rightPrimitive.IsUnsigned()) {
				return leftType >= rightPrimitive
			}
		}
	case *ChlangArrayType:
		if rightArray, ok := right.(*ChlangArrayType); ok {
			return IsLeftCompatibleType(leftType.ElementType, rightArray.ElementType) &&
				(leftType.Length == rightArray.Length || leftType.Length == 0)
		}
	}

	return false
}

// IsCompatibleType checks if the left type is compatible with the right type
func IsCompatibleType(left, right ChlangType) bool {
	if left == right {
		return true
	}

	switch leftType := left.(type) {
	case ChlangPrimitiveType:
		if rightPrimitive, ok := right.(ChlangPrimitiveType); ok {
			if leftType.IsNumeric() && rightPrimitive.IsNumeric() {
				return true
			}
		}
	case *ChlangArrayType:
		if rightArray, ok := right.(*ChlangArrayType); ok {
			return IsCompatibleType(leftType.ElementType, rightArray.ElementType) &&
				(leftType.Length == rightArray.Length || leftType.Length == 0 || rightArray.Length == 0)
		}
	case *ChlangFunctionType:
		if rightFunction, ok := right.(*ChlangFunctionType); ok {
			if len(leftType.Args) != len(rightFunction.Args) {
				return false
			}
			for i, arg := range leftType.Args {
				if !IsCompatibleType(arg, rightFunction.Args[i]) {
					return false
				}
			}
			return IsCompatibleType(leftType.Return, rightFunction.Return)
		}
	}

	return false
}
