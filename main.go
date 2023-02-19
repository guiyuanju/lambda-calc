package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

type tokenType string

const (
	lambda     tokenType = "lambda"
	dot        tokenType = "dot"
	leftParen  tokenType = "leftParen"
	rightParen tokenType = "rightParen"
	whiteSpace tokenType = "whiteSpace" // neccessary for distinguish application
	identifier tokenType = "identifier"
	define     tokenType = "define"
)

type token struct {
	tokenType tokenType
	lexeme    string
}

type scanner struct {
	cur     int
	program []rune
	tokens  []token
}

func (s *scanner) current() rune {
	return s.program[s.cur]
}

func (s *scanner) advance() {
	s.cur += 1
}

func (s *scanner) isEnd() bool {
	return s.cur >= len(s.program)
}

func (s *scanner) addToken(token token) {
	s.tokens = append(s.tokens, token)
}

func (s *scanner) identifier() (token, error) {
	var id string
	isLetter := func(c string) bool {
		return c >= "a" && c <= "z" || c >= "A" && c <= "Z"
	}
	for !s.isEnd() && isLetter(string(s.current())) {
		id += string(s.current())
		s.advance()
	}
	if id == "" {
		return token{}, fmt.Errorf("%v cannot be used in identifier", string(s.current()))
	}
	s.cur -= 1
	return token{identifier, id}, nil
}

func (s *scanner) match(text string) bool {
	prev := s.cur
	reset := func() { s.cur = prev }
	defer reset()
	for _, c := range text {
		if s.isEnd() || c != s.current() {
			return false
		}
		s.advance()
	}
	return true
}

func (s *scanner) consume(text string) string {
	for range text {
		s.advance()
	}
	return text
}

func (s *scanner) binding() token {
	s.consume("def")
	return token{define, "def"}
}

func (s *scanner) scan() ([]token, error) {
	for !s.isEnd() {
		switch cur := s.current(); cur {
		case ' ', '\t':
			s.addToken(token{whiteSpace, string(cur)})
		case '𝞴', '\\':
			s.addToken(token{lambda, "𝞴"})
		case '.':
			s.addToken(token{dot, "."})
		case '(':
			s.addToken(token{leftParen, "("})
		case ')':
			s.addToken(token{rightParen, ")"})
		default:
			// extra space to avoid confliciton with identifier starting with "def"
			if s.match("def ") {
				s.addToken(s.binding())
			}
			if t, err := s.identifier(); err != nil {
				return nil, err
			} else {
				s.addToken(t)
			}
		}
		s.advance()
	}
	return s.tokens, nil
}

type expression interface {
	isExpression()
	String() string
}

type binding struct {
	name  variable
	value expression
}

func (binding) isExpression() {}
func (b binding) String() string {
	return fmt.Sprintf("(def %v %v)", b.name, b.value)
}

type abstraction struct {
	param variable
	expr  expression
}

func (abstraction) isExpression() {}
func (a abstraction) String() string {
	return fmt.Sprintf("(𝞴%v.%v)", a.param, a.expr)
}

type application struct {
	left  expression
	right expression
}

func (application) isExpression() {}
func (a application) String() string {
	return fmt.Sprintf("(%v %v)", a.left, a.right)
}

type variable struct {
	identifier string
}

func (variable) isExpression() {}
func (v variable) String() string {
	return fmt.Sprintf("%v", v.identifier)
}

type parser struct {
	cur    int
	tokens []token
}

func (p *parser) current() token {
	if p.isEnd() {
		panic("unexpected eof")
	}
	return p.tokens[p.cur]
}

func (p *parser) advance() {
	p.cur += 1
}

func (p *parser) consume(tt tokenType) {
	if p.isEnd() {
		panic(fmt.Sprintf("expect %v, but got eof", tt))
	}
	if p.current().tokenType != tt {
		panic(fmt.Sprintf("expect %v, but got %v %v", tt, p.current().tokenType, p.current().lexeme))
	}
	p.advance()
}

func (p *parser) isEnd() bool {
	return p.cur >= len(p.tokens)
}

func (p *parser) parse() expression {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	return p.expression()
}

func (p *parser) expression() expression {
	return p.abstraction()
}

// func (p *parser) binding() expression {
// 	if p.current().tokenType == define {
// 		p.consume(define)
// 		p.consume(whiteSpace)
// 		v := p.variable()
// 		p.consume(whiteSpace)
// 		abs := p.abstraction()
// 		return binding{name: v, value: abs}
// 	}
// 	return p.abstraction()
// }

func (p *parser) abstraction() expression {
	if p.current().tokenType == lambda {
		p.consume(lambda)
		vars := p.variables()
		p.consume(dot)
		exp := p.expression()
		// build nested abstraction
		res := abstraction{vars[len(vars)-1], exp}
		if len(vars) > 1 {
			for i := len(vars) - 2; i >= 0; i-- {
				res = abstraction{vars[i], res}
			}
		}
		return res
	}
	return p.application()
}

func (p *parser) application() expression {
	expr := p.atom()
	for !p.isEnd() && p.current().tokenType == whiteSpace {
		p.consume(whiteSpace)
		expr = application{expr, p.atom()}
	}
	return expr
}

func (p *parser) atom() expression {
	if p.current().tokenType == identifier {
		return p.variable()
	}
	// TODO: error handling
	p.consume(leftParen)
	exp := p.expression()
	p.consume(rightParen)
	return exp
}

func (p *parser) variables() []variable {
	variables := []variable{p.variable()}
	for p.current().tokenType == whiteSpace {
		p.consume(whiteSpace)
		variables = append(variables, p.variable())
	}
	return variables
}

func (p *parser) variable() variable {
	v := p.current().lexeme
	p.consume(identifier)
	return variable{v}
}

type environment struct {
	bindings []struct {
		left  variable
		right expression
	}
}

func (e environment) clone() environment {
	return environment{
		bindings: append([]struct {
			left  variable
			right expression
		}{}, e.bindings...),
	}
}

func (e environment) bind(left variable, right expression) environment {
	newE := e.clone()
	newE.bindings = append(newE.bindings, struct {
		left  variable
		right expression
	}{left, right})
	return newE
}

func (e environment) find(left variable) (expression, bool) {
	for i := len(e.bindings) - 1; i >= 0; i-- {
		if e.bindings[i].left.identifier == left.identifier {
			return e.bindings[i].right, true
		}
	}
	return variable{}, false
}

type interpreter struct {
	ast expression
}

func (i *interpreter) interpret() expression {
	return eval(i.ast, environment{})
}

func eval(exp expression, env environment) expression {
	fmt.Println(exp, env)
	switch exp := exp.(type) {
	case abstraction:
		// variable shadowing
		return abstraction{exp.param, eval(exp.expr, env.bind(exp.param, exp.param))}
	case application:
		left := eval(exp.left, env)
		right := eval(exp.right, env)
		switch left := left.(type) {
		case abstraction:
			return eval(left.expr, env.bind(left.param, right))
		default:
			return application{left, right}
		}
	case variable:
		if right, ok := env.find(exp); ok {
			return right
		}
		return exp
	default:
		return exp
	}
}

func repl() {
	s := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for s.Scan() {
		scanner := scanner{program: []rune(s.Text())}
		tokens, err := scanner.scan()
		if err != nil {
			fmt.Println(err)
			fmt.Print("> ")
			continue
		}
		parser := parser{tokens: tokens}
		interpreter := interpreter{ast: parser.parse()}
		fmt.Println(interpreter.interpret())
		fmt.Print("> ")
	}
	if err := s.Err(); err != nil {
		log.Println(err)
	}

}

func main() {
	repl()

	// program := "asdf saf)"
	// scanner := scanner{program: []rune(program)}
	// tokens, err := scanner.scan()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// fmt.Println(tokens)
	// parser := parser{tokens: tokens}
	// ast := parser.parse()
	// fmt.Println(ast)
	// interpreter := interpreter{ast: ast}
	// fmt.Println(interpreter.interpret())
}
