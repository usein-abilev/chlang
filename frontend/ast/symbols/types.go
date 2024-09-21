package symbols

type SymbolValueType int

// Build-in types
const (
	SymbolTypeInvalid SymbolValueType = iota
	SymbolTypeInt8                    // i8
	SymbolTypeInt16                   // i16
	SymbolTypeInt32                   // i32
	SymbolTypeInt64                   // i64
	SymbolTypeUint8                   // u8
	SymbolTypeUint16                  // u16
	SymbolTypeUint32                  // u32
	SymbolTypeUint64                  // u64
	SymbolTypeFloat32                 // f32
	SymbolTypeFloat64                 // f64
	SymbolTypeBool                    // true, false
	SymbolTypeString                  // string literal
	SymbolTypeVoid                    // void
)

var langSymbolTypeTag = map[SymbolValueType]string{
	SymbolTypeInt8:    "i8",
	SymbolTypeInt16:   "i16",
	SymbolTypeInt32:   "i32",
	SymbolTypeInt64:   "i64",
	SymbolTypeUint8:   "u8",
	SymbolTypeUint16:  "u16",
	SymbolTypeUint32:  "u32",
	SymbolTypeUint64:  "u64",
	SymbolTypeFloat32: "f32",
	SymbolTypeFloat64: "f64",
	SymbolTypeBool:    "bool",
	SymbolTypeString:  "string",
	SymbolTypeVoid:    "void",

	SymbolTypeInvalid: "<invalid>",
}

var symbolTags = map[string]SymbolValueType{
	"i8":     SymbolTypeInt8,
	"i16":    SymbolTypeInt16,
	"i32":    SymbolTypeInt32,
	"i64":    SymbolTypeInt64,
	"u8":     SymbolTypeUint8,
	"u16":    SymbolTypeUint16,
	"u32":    SymbolTypeUint32,
	"u64":    SymbolTypeUint64,
	"f32":    SymbolTypeFloat32,
	"f64":    SymbolTypeFloat64,
	"bool":   SymbolTypeBool,
	"string": SymbolTypeString,
	"void":   SymbolTypeVoid,
}

func (t SymbolValueType) String() string {
	return langSymbolTypeTag[t]
}

func GetTypeByTag(s string) SymbolValueType {
	return symbolTags[s]
}

func (t SymbolValueType) IsNumeric() bool {
	return t.IsInteger() || t.IsFloat()
}

func (t SymbolValueType) IsInteger() bool {
	return t.IsSigned() || t.IsUnsigned()
}

func (t SymbolValueType) IsFloat() bool {
	switch t {
	case SymbolTypeFloat32, SymbolTypeFloat64:
		return true
	}
	return false
}

func (t SymbolValueType) IsSigned() bool {
	switch t {
	case SymbolTypeInt8, SymbolTypeInt16, SymbolTypeInt32, SymbolTypeInt64:
		return true
	}
	return false
}

func (t SymbolValueType) IsUnsigned() bool {
	switch t {
	case SymbolTypeUint8, SymbolTypeUint16, SymbolTypeUint32, SymbolTypeUint64:
		return true
	}
	return false
}

// GetMaxType returns the type with the highest precedence
// Yes, this is a very naive implementation
func GetMaxType(a, b SymbolValueType) SymbolValueType {
	if a > b {
		return a
	} else {
		return b
	}
}

// IsCompatibleType checks if the left type is compatible with the right type
// This is used for type checking
func IsCompatibleType(left, right SymbolValueType) bool {
	if left == right {
		return true
	}

	if (left.IsFloat() && right.IsFloat()) ||
		(left.IsSigned() && right.IsSigned()) ||
		(left.IsUnsigned() && right.IsUnsigned()) {
		return left >= right
	}

	return false
}
