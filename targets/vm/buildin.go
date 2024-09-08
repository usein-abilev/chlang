package vm

import "fmt"

var BuildInFunctions = map[string]func([]*OperandValue){
	"println": buildInPrintln,
}

func buildInPrintln(args []*OperandValue) {
	str := ""
	for idx, arg := range args {
		if idx > 0 {
			str += " "
		}
		if arg.Kind == OperandTypeInt64 {
			str += fmt.Sprintf("%v", arg.Value)
		} else if arg.Kind == OperandTypeString {
			str += fmt.Sprintf("%s", arg.Value)
		} else {
			str += fmt.Sprintf("%v", arg.Value)
		}
	}
	fmt.Printf("%s\n", str)
}
