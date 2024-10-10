// Package env provides functionality for managing a symbol table (environment)
// for a programming language. It includes definitions for symbols, types, and
// scopes, as well as methods for inserting, looking up, and managing these entities.
//
// The package defines the following types:
//
//   - TableEntity: An interface for entities that can be stored in the symbol table.
//   - EnvSymbolEntity: Represents a symbol in the symbol table, including its name,
//     usage status, type, entity type, function arguments, and source code position.
//   - EnvTypeEntity: Represents a type in the symbol table, including its name,
//     usage status, type specification, and source code position.
//   - EnvScope: Represents a lexical scope in the symbol table, containing symbols
//     and types defined in that scope and a reference to its parent scope.
//   - Env: Represents the entire symbol table, including methods for managing scopes
//     and inserting/looking up symbols and types.
//
// The package provides the following methods:
//
// - NewEnv: Creates a new symbol table environment with default types initialized.
// - GetUnusedSymbols: Returns a list of unused symbols in the symbol table.
// - OpenScope: Opens a new lexical scope in the symbol table.
// - CloseScope: Closes the current lexical scope in the symbol table.
// - InsertSymbol: Inserts a new symbol into the current scope.
// - InsertType: Inserts a new type into the current scope.
// - LookupSymbolLocal: Looks up a symbol in the current scope only.
// - LookupSymbol: Looks up a symbol in the current scope and all parent scopes.
// - LookupType: Looks up a type in the current scope and all parent scopes.
// - Print: Prints the current symbol table to the console.
package env

import (
	"fmt"

	"github.com/usein-abilev/chlang/frontend/token"
)

type symbolEntityType int

// Symbol types for the symbol table
const (
	_ symbolEntityType = iota
	SymbolEntityVariable
	SymbolEntityFunction
	SymbolEntityConstant
)

func (t symbolEntityType) String() string {
	switch t {
	case SymbolEntityVariable:
		return "Variable"
	case SymbolEntityFunction:
		return "Function"
	case SymbolEntityConstant:
		return "Constant"
	}
	return "Unknown"
}

type TableEntity interface {
	GetName() string
}

// Represents a single symbol or type specification in the symbol table (environment)
type EnvSymbolEntity struct {
	// Name of the symbol or type
	Name string

	// Whether the symbol is used in the program
	Used bool

	// The type of the symbol: i32, i64, f64, string, etc.
	Type ChlangType

	// The type of the entity: variable, function, constant, etc.
	EntityType symbolEntityType

	// Arguments of the function, if it's a function
	FunctionArgs []*EnvSymbolEntity

	// The position of the symbol in the source code
	Span *token.Span
}

func (s *EnvSymbolEntity) GetName() string {
	return s.Name
}

type EnvTypeEntity struct {
	// Name of the symbol or type
	Name string

	// Whether the symbol is used in the program
	Used bool

	// Actual type specification
	Spec ChlangType

	// The code region of the type in the source code
	Span *token.Span
}

func (t *EnvTypeEntity) GetName() string {
	return t.Name
}

// Represents a lexical scope in the table
// Each scope is a leaf in the scopes tree of the symbol table
type EnvScope struct {
	parent *EnvScope

	// The symbols (variables, functions, etc) defined in the current scope
	symbols map[string]*EnvSymbolEntity

	// The types (struct, user-defined type, enums) defined in the current scope
	types map[string]*EnvTypeEntity
}

type Env struct {
	Local *EnvScope
}

func NewEnv() *Env {
	local := &EnvScope{
		symbols: make(map[string]*EnvSymbolEntity, 0),
		types:   make(map[string]*EnvTypeEntity, 0),
	}
	env := &Env{Local: local}
	env.initDefaultTypes()
	return env
}

// initDefaultTypes initializes the default types in the symbol table
func (st *Env) initDefaultTypes() {
	for name, typeId := range symbolTags {
		st.InsertType(&EnvTypeEntity{
			Name: name,
			Used: true,
			Spec: typeId,
		})
	}
}

func (st *Env) GetUnusedSymbols() []TableEntity {
	return getUnusedSymbolsInScopes(st.Local)
}

func getUnusedSymbolsInScopes(scope *EnvScope) []TableEntity {
	var unused []TableEntity
	for _, symbol := range scope.symbols {
		if !symbol.Used {
			unused = append(unused, symbol)
		}
	}
	for _, t := range scope.types {
		if !t.Used {
			unused = append(unused, t)
		}
	}
	if scope.parent != nil {
		symbols := getUnusedSymbolsInScopes(scope.parent)
		unused = append(unused, symbols...)
	}
	return unused
}

func (st *Env) OpenScope() {
	st.Local = &EnvScope{
		parent:  st.Local,
		symbols: map[string]*EnvSymbolEntity{},
		types:   map[string]*EnvTypeEntity{},
	}
}

func (st *Env) CloseScope() {
	if st.Local.parent == nil {
		return
	}
	st.Local = st.Local.parent
}

func (st *Env) InsertSymbol(symbol *EnvSymbolEntity) bool {
	if _, ok := st.Local.symbols[symbol.Name]; !ok {
		st.Local.symbols[symbol.Name] = symbol
		return true
	}
	return false
}

func (st *Env) InsertType(t *EnvTypeEntity) bool {
	if _, ok := st.Local.types[t.Name]; !ok {
		st.Local.types[t.Name] = t
		return true
	}
	return false
}

// LookupSymbolLocal searches for a symbol in the current scope only
func (st *Env) LookupSymbolLocal(name string) *EnvSymbolEntity {
	symbol, ok := st.Local.symbols[name]
	if !ok {
		return nil
	}
	return symbol
}

// LookupSymbol searches for a symbol in the current scope and all parent scopes
func (st *Env) LookupSymbol(name string) *EnvSymbolEntity {
	return st.Local.lookupSymbol(name)
}

func (st *Env) LookupType(name string) *EnvTypeEntity {
	return st.Local.lookupType(name)
}

func (st *EnvScope) lookupSymbol(name string) *EnvSymbolEntity {
	symbol, ok := st.symbols[name]
	if !ok && st.parent != nil {
		return st.parent.lookupSymbol(name)
	}
	if !ok {
		return nil
	}
	return symbol
}

func (st *EnvScope) lookupType(name string) *EnvTypeEntity {
	t, ok := st.types[name]
	if !ok && st.parent != nil {
		return st.parent.lookupType(name)
	}
	if !ok {
		return nil
	}
	return t
}

func (st *Env) Print() {
	header := []string{"Name", "Type", "Entity Type"}
	columnsSize := []int{0, 0, 0}
	rows := make([][]string, 0)
	for name, symbol := range st.Local.symbols {
		if len(name) > columnsSize[0] {
			columnsSize[0] = len(name)
		}
		typeString := symbol.Type.String()
		if len(typeString) > columnsSize[1] {
			columnsSize[1] = len(typeString)
		}
		if len(symbol.EntityType.String()) > columnsSize[2] {
			columnsSize[2] = len(symbol.EntityType.String())
		}
		rows = append(rows, []string{name, typeString, symbol.EntityType.String()})
	}

	for name, t := range st.Local.types {
		if len(name) > columnsSize[0] {
			columnsSize[0] = len(name)
		}
		typeString := t.Spec.String()
		if len(typeString) > columnsSize[1] {
			columnsSize[1] = len(typeString)
		}
		rows = append(rows, []string{name, typeString, "Type"})
	}

	fmt.Print("\n---------------Symbol Table------------\n")
	for i, h := range header {
		fmt.Printf("%-*s", columnsSize[i]+2, h)
	}
	fmt.Print("\n")
	for _, row := range rows {
		for i, col := range row {
			fmt.Printf("%-*s", columnsSize[i]+2, col)
		}
		fmt.Print("\n")
	}
	fmt.Print("---------------------------------------\n\n")
}
