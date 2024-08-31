package main

import (
	"fmt"

	"github.com/usein-abilev/chlang/frontend"
	"github.com/usein-abilev/chlang/targets/vm"
)

func main() {
	program := frontend.Build("./examples/astdebug/basic.chl")
	codegen := vm.NewRVMGenerator(program)
	module := codegen.Generate()
	fmt.Printf("Generated module: %+v\n", module)
	machine := vm.NewVM(module)
	machine.Run()
	machine.Print()
}
