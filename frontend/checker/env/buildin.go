package env

type ChlangPrimitiveType int

// Build-in types
const (
	SymbolTypeInvalid ChlangPrimitiveType = iota
	SymbolTypeInt8                        // i8
	SymbolTypeInt16                       // i16
	SymbolTypeInt32                       // i32
	SymbolTypeInt64                       // i64
	SymbolTypeUint8                       // u8
	SymbolTypeUint16                      // u16
	SymbolTypeUint32                      // u32
	SymbolTypeUint64                      // u64
	SymbolTypeFloat32                     // f32
	SymbolTypeFloat64                     // f64
	SymbolTypeBool                        // true, false
	SymbolTypeString                      // string literal
	SymbolTypeVoid                        // void
)

var langSymbolTypeTag = map[ChlangPrimitiveType]string{
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

var symbolTags = map[string]ChlangPrimitiveType{
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

func (t ChlangPrimitiveType) String() string {
	return langSymbolTypeTag[t]
}

func GetPrimitiveTypeByTag(s string) (ChlangPrimitiveType, bool) {
	symbol, ok := symbolTags[s]
	return symbol, ok
}

func (t ChlangPrimitiveType) IsNumeric() bool {
	return t.IsInteger() || t.IsFloat()
}

func (t ChlangPrimitiveType) IsInteger() bool {
	return t.IsSigned() || t.IsUnsigned()
}

func (t ChlangPrimitiveType) IsFloat() bool {
	switch t {
	case SymbolTypeFloat32, SymbolTypeFloat64:
		return true
	}
	return false
}

func (t ChlangPrimitiveType) IsSigned() bool {
	switch t {
	case SymbolTypeInt8, SymbolTypeInt16, SymbolTypeInt32, SymbolTypeInt64:
		return true
	}
	return false
}

func (t ChlangPrimitiveType) IsUnsigned() bool {
	switch t {
	case SymbolTypeUint8, SymbolTypeUint16, SymbolTypeUint32, SymbolTypeUint64:
		return true
	}
	return false
}

func (t ChlangPrimitiveType) GetNumberBitSize() int {
	switch t {
	case SymbolTypeInt8, SymbolTypeUint8:
		return 8
	case SymbolTypeInt16, SymbolTypeUint16:
		return 16
	case SymbolTypeInt32, SymbolTypeUint32:
		return 32
	case SymbolTypeInt64, SymbolTypeUint64:
		return 64
	}
	return 0
}

// GetMaxType returns the type with the highest precedence
// Yes, this is a very naive implementation
func GetMaxType(a, b ChlangPrimitiveType) ChlangPrimitiveType {
	if a > b {
		return a
	} else {
		return b
	}
}
