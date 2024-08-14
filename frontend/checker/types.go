package checker

type LangSymbolType uint32

// Build-in types
const (
	LangTypeInvalid LangSymbolType = iota
	LangTypeInt32                  // i32
	LangTypeInt64                  // i64
	LangTypeUint32                 // u32
	LangTypeUint64                 // u64
	LangTypeFloat32                // f32
	LangTypeFloat64                // f64
	LangTypeBool                   // true, false
	LangTypeString                 // string literal
	LangTypeVoid                   // void
)

var langSymbolTypeTag = map[LangSymbolType]string{
	LangTypeInt32:   "i32",
	LangTypeInt64:   "i64",
	LangTypeUint32:  "u32",
	LangTypeUint64:  "u64",
	LangTypeFloat32: "f32",
	LangTypeFloat64: "f64",
	LangTypeBool:    "bool",
	LangTypeString:  "string",
	LangTypeVoid:    "void",

	LangTypeInvalid: "<invalid>",
}

var langSymbolTypeFromStr = map[string]LangSymbolType{
	"i32":    LangTypeInt32,
	"i64":    LangTypeInt64,
	"u32":    LangTypeUint32,
	"u64":    LangTypeUint64,
	"f32":    LangTypeFloat32,
	"f64":    LangTypeFloat64,
	"bool":   LangTypeBool,
	"string": LangTypeString,
	"void":   LangTypeVoid,
}

func (t LangSymbolType) String() string {
	return langSymbolTypeTag[t]
}

func LangSymbolTypeFromStr(s string) LangSymbolType {
	return langSymbolTypeFromStr[s]
}

func (t LangSymbolType) IsNumeric() bool {
	switch t {
	case LangTypeInt32, LangTypeInt64, LangTypeUint32, LangTypeUint64, LangTypeFloat32, LangTypeFloat64:
		return true
	}
	return false
}

func (t LangSymbolType) IsInteger() bool {
	switch t {
	case LangTypeInt32, LangTypeInt64, LangTypeUint32, LangTypeUint64:
		return true
	}
	return false
}

func (t LangSymbolType) IsFloat() bool {
	switch t {
	case LangTypeFloat32, LangTypeFloat64:
		return true
	}
	return false
}

func (t LangSymbolType) IsSigned() bool {
	switch t {
	case LangTypeInt32, LangTypeInt64:
		return true
	}
	return false
}

func (t LangSymbolType) IsUnsigned() bool {
	switch t {
	case LangTypeUint32, LangTypeUint64:
		return true
	}
	return false
}
