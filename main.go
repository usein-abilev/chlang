package main

import (
	// "github.com/usein-abilev/chlang/codegen"
	// "github.com/usein-abilev/chlang/frontend"
	"github.com/usein-abilev/chlang/vm"
)

func main() {
	// program := frontend.Build("./examples/astdebug/basic.chl")
	builder := vm.NewASMBuilder()

	// codegen.GenerateVMAssembly(program, builder)
	builder.Label("main")
	builder.Add(vm.LoadImm, 0, 111111) // pushes 111111 to stack
	builder.Add(vm.LoadImm, 1, 222222) // pushes 222222 to stack
	builder.Add(vm.LoadImm, 5, 10)     // argument of the function sum
	builder.Add(vm.LoadImm, 6, 20)     // argument of the function sum
	builder.Add(vm.LoadImm, 7, 30)     // argument of the function sum
	builder.Add(vm.LoadImm, 8, 40)     // argument of the function sum
	builder.Add(vm.Call, "sum", 4, 1)  // 4 arguments, 1 return values
	builder.Add(vm.Call, "sum", 1, 1)  // 4 arguments, 1 return values
	builder.Label("sum")
	builder.Add(vm.Add, 0, 0, 1)
	builder.Add(vm.Add, 0, 0, 2)
	builder.Add(vm.Add, 0, 0, 3)
	builder.Add(vm.Halt)
	code := builder.Build()
	builder.Print()

	machine := vm.New()
	machine.LoadProgram(code)
	machine.Run()
	machine.Print()
}
