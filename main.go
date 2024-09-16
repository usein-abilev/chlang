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
	machine := vm.NewVM(module, &vm.VMOptions{Debug: false})
	fmt.Printf("\n======== Program output ========\n\n")
	machine.Run()
}
