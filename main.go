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
	whiteSpace tokenType = "whiteSpace"
	identifier tokenType = "identifier"
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

func (s *scanner) identifier() token {
	var id string
	isLetter := func(c string) bool {
		return c >= "a" && c <= "z" || c >= "A" && c <= "Z"
	}
	for !s.isEnd() && isLetter(string(s.current())) {
		id += string(s.current())
		s.advance()
	}
	s.cur -= 1
	return token{identifier, id}
}

func (s *scanner) scan() []token {
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
			s.addToken(s.identifier())
		}
		s.advance()
	}
	return s.tokens
}

type expression interface {
	isExpression()
	textify() string
}

type abstraction struct {
	param variable
	expr  expression
}

func (abstraction) isExpression() {}
func (a abstraction) String() string {
	return fmt.Sprintf("(Abs %v %v)", a.param, a.expr)
}
func (a abstraction) textify() string {
	return fmt.Sprintf("(𝞴%v.%v)", a.param.textify(), a.expr.textify())
}

type application struct {
	left  expression
	right expression
}

func (application) isExpression() {}
func (a application) String() string {
	return fmt.Sprintf("(App %v %v)", a.left, a.right)
}
func (a application) textify() string {
	return fmt.Sprintf("(%v %v)", a.left.textify(), a.right.textify())
}

type variable struct {
	identifier string
}

func (variable) isExpression() {}
func (v variable) String() string {
	return fmt.Sprintf("%v", v.identifier)
}
func (v variable) textify() string {
	return v.String()
}

type parser struct {
	cur    int
	tokens []token
}

func (p *parser) current() token {
	return p.tokens[p.cur]
}

func (p *parser) advance() {
	p.cur += 1
}

func (p *parser) isEnd() bool {
	return p.cur >= len(p.tokens)
}

func (p *parser) parse() expression {
	return p.expression()
}

func (p *parser) expression() expression {
	return p.abstraction()
}

func (p *parser) abstraction() expression {
	if p.current().tokenType == lambda {
		p.advance()
		vars := p.variables()
		p.advance()
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
		p.advance()
		expr = application{expr, p.atom()}
	}
	return expr
}

func (p *parser) atom() expression {
	if p.current().tokenType == identifier {
		return p.variable()
	}
	// TODO: error handling
	p.advance()
	exp := p.expression()
	p.advance()
	return exp
}

func (p *parser) variables() []variable {
	variables := []variable{p.variable()}
	for p.current().tokenType == whiteSpace {
		p.advance()
		variables = append(variables, p.variable())
	}
	return variables
}

func (p *parser) variable() variable {
	v := p.current().lexeme
	p.advance()
	return variable{v}
}

type binding struct {
	left  variable
	right expression
}

type environment struct {
	bindings []binding
}

func (e environment) bind(left variable, right expression) environment {
	newE := append([]binding{}, e.bindings...)
	return environment{
		append(newE, binding{left, right}),
	}
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
		parser := parser{tokens: scanner.scan()}
		interpreter := interpreter{ast: parser.parse()}
		fmt.Println(interpreter.interpret().textify())
		fmt.Print("> ")
	}
	if err := s.Err(); err != nil {
		log.Println(err)
	}

}

func main() {
	repl()
}
