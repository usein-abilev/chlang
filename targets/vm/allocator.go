package vm

import "fmt"

const maxRegisters = 0xFF
const tempVariableName = "<temp>"

// Dummy register allocator which not consider the lifetime of the variables
// TODO: Need to be replaced by Graph-coloring algorithm
type RegisterAllocator struct {
	// The register table contains the mapping of variable names to register addresses.
	table *RegisterTable

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
	idx := len(table) - 1
	for ; table[idx].depth == r.scopeDepth; idx-- {
	}
	*r.table = table[:idx+1]

	r.scopeDepth--
}

func (r *RegisterAllocator) IsTempRegister(address RegisterAddress) bool {
	r.checkRegisterBoundary(address)
	return (*r.table)[address.AsInt()].name == tempVariableName
}

// AllocateTempRegister allocates a temporary register for intermediate values.
func (r *RegisterAllocator) AllocateTempRegister() RegisterAddress {
	return r.AllocateRegister(tempVariableName)
}

// Binds a register to variable name, the register must be a temporary or free.
// It safe to call this method on bound address (varRegister)
func (r *RegisterAllocator) BindRegister(name string, address RegisterAddress) bool {
	r.checkRegisterBoundary(address)
	local := &(*r.table)[address.AsInt()]
	if local.name != tempVariableName {
		return false
	}
	local.name = name
	return true
}

// AllocateRegister allocates a register for the given variable name.
func (r *RegisterAllocator) AllocateRegister(name string) RegisterAddress {
	if name != tempVariableName {
		local, exists := r.table.LookupName(name)
		if exists {
			return local.register
		}
	}

	address := RegisterAddress(len(*r.table))
	local := &LocalRegister{
		name:     name,
		register: address,
		depth:    r.scopeDepth,
	}
	*r.table = append(*r.table, *local)
	return local.register
}

// LookupVariable returns the register address for the given variable name.
func (r *RegisterAllocator) LookupVariable(name string) (*LocalRegister, bool) {
	return r.table.LookupName(name)
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
