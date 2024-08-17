package symbols

type SymbolValueType int

// Build-in types
const (
	SymbolTypeInvalid SymbolValueType = iota
	SymbolTypeInt32                   // i32
	SymbolTypeInt64                     // i64
	SymbolTypeUint32                    // u32
	SymbolTypeUint64                    // u64
	SymbolTypeFloat32                   // f32
	SymbolTypeFloat64                   // f64
	SymbolTypeBool                      // true, false
	SymbolTypeString                    // string literal
	SymbolTypeVoid                      // void
)

var langSymbolTypeTag = map[SymbolValueType]string{
	SymbolTypeInt32: "i32",
	SymbolTypeInt64:   "i64",
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
	"i32":    SymbolTypeInt32,
	"i64":    SymbolTypeInt64,
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
	switch t {
	case SymbolTypeInt32, SymbolTypeInt64, SymbolTypeUint32, SymbolTypeUint64, SymbolTypeFloat32, SymbolTypeFloat64:
		return true
	}
	return false
}

func (t SymbolValueType) IsInteger() bool {
	switch t {
	case SymbolTypeInt32, SymbolTypeInt64, SymbolTypeUint32, SymbolTypeUint64:
		return true
	}
	return false
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
	case SymbolTypeInt32, SymbolTypeInt64:
		return true
	}
	return false
}

func (t SymbolValueType) IsUnsigned() bool {
	switch t {
	case SymbolTypeUint32, SymbolTypeUint64:
		return true
	}
	return false
}
