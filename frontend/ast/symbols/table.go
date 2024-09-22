package symbols

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

	// User defined types
	SymbolEntityUserDefined
)

func (t symbolEntityType) String() string {
	switch t {
	case SymbolEntityVariable:
		return "Variable"
	case SymbolEntityFunction:
		return "Function"
	case SymbolEntityConstant:
		return "Constant"
	case SymbolEntityUserDefined:
		return "UserDefined"
	}
	return "Unknown"
}

type FuncSymbolSignature struct {
	SpreadArgument bool
	Args           []*SymbolEntity
	ReturnType     ChlangType
}

// Represents a single symbol or type specification in the symbol table (environment)
type SymbolEntity struct {
	Name string

	// Whether the symbol is used in the program
	Used bool

	// The type of the symbol: i32, i64, f64, string, etc.
	Type ChlangType

	// The type of the entity: variable, function, constant, etc.
	EntityType symbolEntityType

	// If the symbol is a function,
	// this field will be contain the function's details
	Function *FuncSymbolSignature

	// The position of the symbol in the source code
	Position token.TokenPosition
}

type Scope struct {
	parent  *Scope
	symbols map[string]*SymbolEntity
}

type SymbolTable struct {
	current *Scope
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		current: &Scope{
			symbols: make(map[string]*SymbolEntity),
		},
	}
}

func (st *SymbolTable) GetUnusedSymbols() []*SymbolEntity {
	return getUnusedSymbolsInScope(st.current)
}

func getUnusedSymbolsInScope(scope *Scope) []*SymbolEntity {
	var unused []*SymbolEntity
	for _, symbol := range scope.symbols {
		if !symbol.Used {
			unused = append(unused, symbol)
		}
	}
	if scope.parent != nil {
		unused = append(unused, getUnusedSymbolsInScope(scope.parent)...)
	}
	return unused
}

func (st *SymbolTable) OpenScope() {
	st.current = &Scope{
		parent:  st.current,
		symbols: make(map[string]*SymbolEntity),
	}
}

func (st *SymbolTable) CloseScope() {
	if st.current.parent == nil {
		return
	}
	st.current = st.current.parent
}

func (st *SymbolTable) Insert(symbol *SymbolEntity) {
	if _, ok := st.current.symbols[symbol.Name]; !ok {
		st.current.symbols[symbol.Name] = symbol
		return
	}
	panic(fmt.Sprintf("Symbol already exists %+v", symbol))
}

func (st *SymbolTable) LookupInScope(name string) *SymbolEntity {
	symbol, ok := st.current.symbols[name]
	if !ok {
		return nil
	}
	return symbol
}

// Lookup searches for a symbol in the current scope and all parent scopes
func (st *SymbolTable) Lookup(name string) *SymbolEntity {
	return st.current.lookup(name)
}

func (st *Scope) lookup(name string) *SymbolEntity {
	symbol, ok := st.symbols[name]
	if !ok && st.parent != nil {
		return st.parent.lookup(name)
	}
	if !ok {
		return nil
	}
	return symbol
}

func (st *SymbolTable) Print() {
	header := []string{"Name", "Type", "Entity Type"}
	columnsSize := []int{0, 0, 0}
	rows := make([][]string, 0)
	for name, symbol := range st.current.symbols {
		if len(name) > columnsSize[0] {
			columnsSize[0] = len(name)
		}
		typeString := fmt.Sprintf("%s", symbol.Type)
		if len(typeString) > columnsSize[1] {
			columnsSize[1] = len(typeString)
		}
		if len(symbol.EntityType.String()) > columnsSize[2] {
			columnsSize[2] = len(symbol.EntityType.String())
		}
		rows = append(rows, []string{name, typeString, symbol.EntityType.String()})
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
