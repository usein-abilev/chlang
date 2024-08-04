package main

import (
	"github.com/usein-abilev/chlang/scanner"
	"github.com/usein-abilev/chlang/token"
	"log"
	"os"
)

func readFileBytes(path string) []byte {
	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Cannot read file: %s", path)
	}
	return bytes
}

func main() {
	bytes := readFileBytes("./examples/hodgepodge.chl")
	source := string(bytes)
	lexer, err := scanner.New(source)
	if err != nil {
		log.Fatalf("Cannot create lexer: %s", err)
	}
	transformed := ""
	for {
		tok := lexer.Scan()
		if tok.Type == token.EOF {
			break
		}
		transformed += tok.Literal + " "
		log.Printf("Scanned token: (%s) -> '%s'", token.TokenSymbolNames[tok.Type], tok.Literal)
	}

	log.Printf("Read %d bytes", len(bytes))
	log.Printf("Transformed: %s", transformed)
}
