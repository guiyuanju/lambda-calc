package main

import "testing"

var cases = []struct {
	program string
	textify string
	value   string
}{
	{
		"𝞴x.x",
		"(𝞴x.x)",
		"(𝞴x.x)",
	},
	{
		"(𝞴x.x)",
		"(𝞴x.x)",
		"(𝞴x.x)",
	},
	{
		"𝞴x.f x",
		"(𝞴x.(f x))",
		"(𝞴x.(f x))",
	},
	{
		"𝞴x.(f x)",
		"(𝞴x.(f x))",
		"(𝞴x.(f x))",
	},
	{
		"𝞴x y.x y",
		"(𝞴x.(𝞴y.(x y)))",
		"(𝞴x.(𝞴y.(x y)))",
	},
	{
		"x y z",
		"((x y) z)",
		"((x y) z)",
	},
	{
		"x (y z)",
		"(x (y z))",
		"(x (y z))",
	},
	{
		"x",
		"x",
		"x",
	},
	{
		"(x)",
		"x",
		"x",
	},
	{
		"(𝞴x.x) y",
		"((𝞴x.x) y)",
		"y",
	},
	{
		"(𝞴x y x.x x) z",
		"((𝞴x.(𝞴y.(𝞴x.(x x)))) z)",
		"(𝞴y.(𝞴x.(x x)))",
	},
	{
		"(𝞴x.x x) y",
		"((𝞴x.(x x)) y)",
		"(y y)",
	},
	// {
	// 	"(x",
	// 	"(x",
	// 	"(x",
	// },
	// {
	// 	"(def x y)",
	// 	"()"
	// }
}

func TestScanner(t *testing.T) {
	for _, tt := range cases {
		scanner := scanner{program: []rune(tt.program)}
		tokens, _ := scanner.scan()
		parser := parser{tokens: tokens}
		res := parser.parse().String()
		t.Run(tt.program, func(t *testing.T) {
			if res != tt.textify {
				t.Errorf("expected %v, but got %v", tt.textify, res)
			}
		})
	}
}

func TestInterpreter(t *testing.T) {
	for _, tt := range cases {
		scanner := scanner{program: []rune(tt.program)}
		tokens, _ := scanner.scan()
		parser := parser{tokens: tokens}
		ast := parser.parse()
		interpreter := interpreter{ast: ast}
		value := interpreter.interpret()
		t.Run(tt.program, func(t *testing.T) {
			if value.String() != tt.value {
				t.Errorf("expected %v, but got %v", tt.value, value.String())
			}
		})
	}
}
