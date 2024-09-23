package vm

import (
	"fmt"
	"strconv"
)

var BuildInFunctions = map[string]func([]*OperandValue){
	"println": buildInPrintln,
}

func buildInPrintln(args []*OperandValue) {
	str := ""
	for idx, arg := range args {
		if arg == nil {
			panic("buildInPrintln: nil argument")
		}
		if idx > 0 {
			str += " "
		}
		str += stringifyOperandValue(arg)
	}
	fmt.Printf("%s\n", str)
}

func stringifyOperandValue(operand *OperandValue) string {
	if operand == nil {
		return "nil"
	}
	switch operand.Kind {
	case OperandTypeInt8, OperandTypeInt16, OperandTypeInt32, OperandTypeInt64:
		return fmt.Sprintf("%v", operand.Value)
	case OperandTypeFloat32, OperandTypeFloat64:
		return fmt.Sprintf("%v", operand.Value)
	case OperandTypeBool:
		return fmt.Sprintf("%v", operand.Value)
	case OperandTypeString:
		return parseStringLiteral(operand.Value.(string))
	case OperandTypeArray:
		arr := operand.Value.([]OperandValue)
		str := "["
		for idx, item := range arr {
			if idx > 0 {
				str += ", "
			}
			str += stringifyOperandValue(&item)
		}
		str += "]"
		return str
	case OperandTypeFunctionObject:
		return "function"
	case OperandTypeBuildInFunction:
		return "build-in-function"
	}
	return "undefined"
}

func parseStringLiteral(value string) string {
	value, err := strconv.Unquote(value)
	if err != nil {
		panic(fmt.Sprintf("parseStringLiteral: invalid string literal: %v", value))
	}
	return value
}
