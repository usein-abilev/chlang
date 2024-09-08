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
		if arg.Kind == OperandTypeInt64 {
			str += fmt.Sprintf("%v", arg.Value)
		} else if arg.Kind == OperandTypeString {
			strValue := arg.Value.(string)
			str += parseStringLiteral(strValue)
		} else {
			str += fmt.Sprintf("%v", arg.Value)
		}
	}
	fmt.Printf("%s\n", str)
}

func parseStringLiteral(value string) string {
	value, err := strconv.Unquote(value)
	if err != nil {
		panic(fmt.Sprintf("parseStringLiteral: invalid string literal: %v", value))
	}
	return value
}
