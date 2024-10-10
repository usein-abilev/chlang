package env

import "fmt"

// Composite type interface
type ChlangType interface {
	Type()
	String() string
}

type ChlangTraitType struct {
	Name         string
	Signatures   []*ChlangFunctionType
	Declarations []*EnvSymbolEntity
}

func (ChlangTraitType) Type() {}
func (c ChlangTraitType) String() string {
	return "trait " + c.Name
}

type ChlangStructField struct {
	Name string
	Type ChlangType
}

// Struct type, e.g. struct { a: i32, b: i32 }
type ChlangStructType struct {
	Name    string
	Fields  []*ChlangStructField
	Methods map[string]*EnvSymbolEntity
}

func (s *ChlangStructType) LookupField(name string) *ChlangStructField {
	for _, field := range s.Fields {
		if field.Name == name {
			return field
		}
	}
	return nil
}

func (s *ChlangStructType) LookupMethod(name string) *EnvSymbolEntity {
	return s.Methods[name]
}

func (ChlangStructType) Type() {}
func (c ChlangStructType) String() string {
	return "struct " + c.Name
}

// Array type, e.g. i32[10], i32[]
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

// Represents a function type in the language
// Example: (i32, i32) -> i32
// Example 1: (MyOwnType, i32) -> (i32, MyOwnType)
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
