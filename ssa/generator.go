// This file contains the code for generating the SSA form of the input program.
package ssa

import (
	"github.com/usein-abilev/chlang/frontend/ast"
)

type SSAOperation int

const (
	SSAOperationAdd SSAOperation = iota
	SSAOperationSub
	SSAOperationMul
	SSAOperationDiv
	SSAOperationCall
	SSAOperationReturn
	SSAOperationPhi // Represents a phi node.
)

// For example, the following code:
/*
fn sum (a: i32, b: i32) -> i32 {
	return a + b;
}
fn main() {
	let a = 1
	a += 2 + 3 * 5
	let b = 2
	let c = sum(a, b)
	if a > b {
		println(a)
	} else if a < b {
		println(b)
	} else {
		println(c)
	}
}

// Will be converted to the following SSA form:
main:
	a.0: i32 = 1
	a.1: i32 = a.0 + 2
	a.2: i32 = 3 * 5
	a.3: i32 = a.1 + a.2
	b.0: i32 = 2
	c.0: i32 = call sum(a.3, b.0)
	if a > b goto if1 else goto if2
if1:
	println(a.3)
	goto end
if2:
	println(b.0)
	goto end
end:
	println(c.0)
	return
sum:
	return a + b
*/

// Represents an SSA instruction.
// Each instruction
type SSAInstruction struct {
	Operation SSAOperation
}

type SSABlock struct {
	Instructions []SSAInstruction
}

type SSAGenerator struct {
	Program       *ast.Program
	LabeledBlocks map[string]*SSABlock
}
