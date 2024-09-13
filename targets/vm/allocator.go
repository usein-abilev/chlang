package vm

import "fmt"

const maxRegisters = 0xFF
const tempVariableName = "<temp>"

// Dummy register allocator which not consider the lifetime of the variables
// TODO: Need to be replaced by Graph-coloring algorithm
type RegisterAllocator struct {
	// The register table contains the mapping of variable names to register addresses.
	table *RegisterTable

	// The list of registers that are freed and can be reused.
	freed []int

	// Determinate the current scope depth
	scopeDepth int
}

func NewRegisterAllocator(table *RegisterTable) *RegisterAllocator {
	binding := &RegisterAllocator{
		scopeDepth: 0,
		table:      table,
	}
	return binding
}

func (r *RegisterAllocator) EnterScope() {
	r.scopeDepth++
}

func (r *RegisterAllocator) LeaveScope() {
	if r.scopeDepth == 0 {
		panic("Cannot leave the root scope")
	}

	// remove all variables in the current scope
	table := *r.table
	for idx := len(table) - 1; idx >= 0 && table[idx].depth == r.scopeDepth; idx-- {
		r.freed = append(r.freed, table[idx].register.AsInt())
		table[idx].name = ""
	}

	r.scopeDepth--
}

func (r *RegisterAllocator) IsTempRegister(address RegisterAddress) bool {
	r.checkRegisterBoundary(address)
	return (*r.table)[address.AsInt()].name == tempVariableName
}

// AllocateTemp allocates a temporary register for intermediate values.
func (r *RegisterAllocator) AllocateTemp() RegisterAddress {
	return r.Allocate(tempVariableName)
}

// FreeTemp frees the temporary register.
func (r *RegisterAllocator) FreeTemp(address RegisterAddress) {
	if r.IsTempRegister(address) {
		r.freed = append(r.freed, address.AsInt())
		table := *r.table
		table[address.AsInt()].name = ""
	}
}

// Binds a register to variable name, the register must be a temporary or free.
// It safe to call this method on bound address (varRegister)
func (r *RegisterAllocator) BindRegister(name string, address RegisterAddress) bool {
	r.checkRegisterBoundary(address)
	local := &(*r.table)[address.AsInt()]
	if local.name == tempVariableName || local.name == "" {
		r.takeFreeRegister(address)
		local.name = name
		return true
	}
	panic("cannot bind a register that is already bound to " + local.name)
}

// Allocate allocates a register for the given variable name.
func (r *RegisterAllocator) Allocate(variable string) RegisterAddress {
	if variable != tempVariableName {
		local, exists := r.table.LookupName(variable)
		if exists {
			return local.register
		}
	}

	local := &LocalRegister{
		name:  variable,
		depth: r.scopeDepth,
	}

	if len(r.freed) > 0 {
		index := r.freed[len(r.freed)-1]
		local.register = RegisterAddress(index)
		r.freed = r.freed[:len(r.freed)-1]
		(*r.table)[index] = *local
	} else {
		local.register = RegisterAddress(len(*r.table))
		*r.table = append(*r.table, *local)
	}

	return local.register
}

// Free frees the register for the given variable name.
func (r *RegisterAllocator) Free(variable string) {
	local, exists := r.table.LookupName(variable)
	if !exists {
		return
	}
	local.name = ""
	r.freed = append(r.freed, local.register.AsInt())
}

// LookupVariable returns the register address for the given variable name.
func (r *RegisterAllocator) LookupVariable(name string) (*LocalRegister, bool) {
	return r.table.LookupName(name)
}

// Take the register out of the freed list
func (r *RegisterAllocator) takeFreeRegister(address RegisterAddress) {
	removeIdx := -1
	for i := len(r.freed) - 1; i >= 0; i-- {
		if r.freed[i] == address.AsInt() {
			removeIdx = i
			break
		}
	}
	if removeIdx != -1 {
		r.freed = append(r.freed[:removeIdx], r.freed[removeIdx+1:]...)
	}
}

func (r *RegisterAllocator) checkRegisterBoundary(address RegisterAddress) {
	if address.AsInt() >= len(*r.table) {
		panic(fmt.Sprintf("error: invalid register boundary addr=%d, max=%d", address, len(*r.table)))
	}
}

func (r RegisterTable) String() string {
	str := ""
	for idx, reg := range r {
		str += fmt.Sprintf("%s: R%d, Idx: %d\n", reg.name, reg.register, idx)
	}
	return str
}

func (r *RegisterTable) LookupName(name string) (*LocalRegister, bool) {
	for i := len(*r) - 1; i >= 0; i-- {
		if (*r)[i].name == name {
			return &(*r)[i], true
		}
	}
	return nil, false
}
