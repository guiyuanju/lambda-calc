package main

import (
	"june/lambda/lambda"
)

func main() {
	lambda.Repl()

	// program := "let \na = b in c"
	// scanner := lambda.Scanner{Program: []rune(program)}
	// tokens, err := scanner.Scan()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println(tokens)
	// parser := lambda.Parser{Tokens: tokens}
	// ast := parser.Parse()
	// fmt.Println(ast)
	// interpreter := lambda.Interpreter{Ast: ast}
	// fmt.Println(interpreter.Interpret())
}
