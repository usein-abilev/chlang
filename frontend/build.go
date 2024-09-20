package frontend

import (
	"fmt"
	"os"

	"github.com/usein-abilev/chlang/frontend/ast"
	"github.com/usein-abilev/chlang/frontend/checker"
	"github.com/usein-abilev/chlang/frontend/errors"
	"github.com/usein-abilev/chlang/frontend/scanner"
	"github.com/usein-abilev/chlang/frontend/transformer"
)

func Build(filepath string) *ast.Program {
	bytes := readFileBytes(filepath)
	source := string(bytes)
	lexer, err := scanner.New(source)
	if err != nil {
		panic(fmt.Sprintf("Cannot create lexer: %s", err))
	}
	program, parserErrors := ast.Init(lexer).Parse()
	checker := checker.Check(program)
	checker.SymbolTable.Print()

	unusedSymbols := checker.SymbolTable.GetUnusedSymbols()
	if len(unusedSymbols) > 0 {
		fmt.Println("[warn] Unused symbols:")
		for _, symbol := range unusedSymbols {
			fmt.Printf("  %s\n", symbol.Name)
		}
	}

	hasErrors := len(*parserErrors) > 0 || len(checker.Errors) > 0
	if hasErrors {
		fmt.Println("Errors found:")
		if len(*parserErrors) > 0 {
			for _, err := range *parserErrors {
				switch e := err.(type) {
				case *errors.SyntaxError:
					e.Write(os.Stdout)
				default:
					fmt.Printf("Unknown error type: %T", e)
				}
			}
		}
		for _, err := range checker.Errors {
			switch e := err.(type) {
			case *errors.SemanticError:
				e.Write(os.Stdout)
			default:
				fmt.Printf("Unknown error type: %T", e)
			}
		}

		os.Exit(1)
	}

	// AST optimization phase
	program = transformer.Transform(program)

	return program
}

func readFileBytes(path string) []byte {
	bytes, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("Cannot read file: %s", path))
	}
	return bytes
}
