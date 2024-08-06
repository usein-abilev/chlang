package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/usein-abilev/chlang/parser"
	"github.com/usein-abilev/chlang/scanner"
	"github.com/usein-abilev/chlang/token"
)

func readFileBytes(path string) []byte {
	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Cannot read file: %s", path)
	}
	return bytes
}

func executeExp(mem *map[string]interface{}, exp parser.Expression) string {
	switch exp := exp.(type) {
	case *parser.IntLiteral:
		return exp.Value
	case *parser.FloatLiteral:
		return exp.Value
	case *parser.Identifier:
		return (*mem)[exp.Value].(string)
	case *parser.UnaryExp:
		right := executeExp(mem, exp.Right)
		rightInt, err := strconv.Atoi(right)
		if err != nil {
			log.Fatalf("Cannot convert right to int: %s", right)
		}
		switch exp.Op.Type {
		case token.MINUS:
			return fmt.Sprint(-rightInt)
		}
	case *parser.BinaryExp:
		left := executeExp(mem, exp.Left)
		right := executeExp(mem, exp.Right)

		leftInt, err := strconv.Atoi(left)
		if err != nil {
			log.Fatalf("Cannot convert left to int: %s", left)
		}
		rightInt, err := strconv.Atoi(right)
		if err != nil {
			log.Fatalf("Cannot convert right to int: %s %T", right, exp.Right)
		}

		switch exp.Op.Type {
		case token.PLUS:
			return fmt.Sprint(leftInt + rightInt)
		case token.MINUS:
			return fmt.Sprint(leftInt - rightInt)
		case token.ASTERISK:
			return fmt.Sprint(leftInt * rightInt)
		case token.SLASH:
			return fmt.Sprint(leftInt / rightInt)
		case token.PERCENT:
			return fmt.Sprint(leftInt % rightInt)
		case token.EXPONENT:
			return fmt.Sprint(int64(math.Pow(float64(leftInt), float64(rightInt))))
		}
	default:
		log.Fatalf("Unknown expression type: %v", exp)

	}
	return "<unknown>"
}

func executeProgram(program *parser.Program) {
	log.Printf("Executing program: %v", program)
	var fakememory = make(map[string]interface{})
	for _, statement := range program.Statements {
		if varDecl, ok := statement.(*parser.VarDeclarationStatement); ok {
			fakememory[varDecl.Name.Value] = executeExp(&fakememory, varDecl.Value)
		}
	}

	log.Printf("Memory: %v", fakememory)
}

func main() {
	bytes := readFileBytes("./examples/astdebug/variables.chl")
	source := string(bytes)
	lexer, err := scanner.New(source)
	if err != nil {
		log.Fatalf("Cannot create lexer: %s", err)
	}
	parser := parser.New(lexer)
	program := parser.Parse()
	b, err := json.Marshal(program)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Printf(string(b))
	executeProgram(program)
	log.Printf("Program: %v", program)
}
