package lambda

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type tokenType string

const (
	lambda     tokenType = "lambda"
	dot        tokenType = "dot"
	leftParen  tokenType = "leftParen"
	rightParen tokenType = "rightParen"
	whiteSpace tokenType = "whiteSpace" // neccessary for distinguish application
	identifier tokenType = "identifier"
	let        tokenType = "let"
	equal      tokenType = "equal"
	in         tokenType = "in"
	quote      tokenType = "'"
)

type token struct {
	tokenType tokenType
	lexeme    string
}

type Scanner struct {
	cur     int
	Program []rune
	tokens  []token
}

func (s *Scanner) current() rune {
	return s.Program[s.cur]
}

func (s *Scanner) advance() {
	s.cur += 1
}

func (s *Scanner) isEnd() bool {
	return s.cur >= len(s.Program)
}

func (s *Scanner) addToken(token token) {
	s.tokens = append(s.tokens, token)
}

func (s *Scanner) identifier() (token, error) {
	var id string
	isLetter := func(c string) bool {
		return c >= "a" && c <= "z" || c >= "A" && c <= "Z"
	}
	isDigit := func(c string) bool {
		return c >= "0" && c <= "9"
	}
	isArithmatic := func(c string) bool {
		return c == "+" || c == "-" || c == "*" || c == "/"
	}
	for !s.isEnd() &&
		(isLetter(string(s.current())) ||
			isDigit(string(s.current())) ||
			isArithmatic(string(s.current()))) {
		id += string(s.current())
		s.advance()
	}
	if id == "" {
		return token{}, fmt.Errorf("%v cannot be used in identifier", string(s.current()))
	}
	return token{identifier, id}, nil
}

func (s *Scanner) match(text string) bool {
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

func (s *Scanner) consume(text string) error {
	for _, c := range text {
		if c != s.current() {
			return fmt.Errorf("expected %v in %v, but got %v", c, text, s.current())
		}
		s.advance()
	}
	return nil
}

func (s *Scanner) consumeOneOf(xs []rune) error {
	for _, c := range xs {
		if s.current() == c {
			s.advance()
			return nil
		}
	}
	return fmt.Errorf("expected one of %v, got %v", xs, s.current())
}

func (s *Scanner) Scan() ([]token, error) {
	s.Program = []rune(strings.Trim(string(s.Program), " \t\n"))
	for !s.isEnd() {
		switch cur := s.current(); cur {
		case ' ', '\t', '\n':
			s.consumeOneOf([]rune{' ', '\t', '\n'})
			s.addToken(token{whiteSpace, " "})
		case '𝞴', 'λ', '\\':
			s.consumeOneOf([]rune{'𝞴', 'λ', '\\'})
			s.addToken(token{lambda, "𝞴"})
		case '.':
			s.consume(".")
			s.addToken(token{dot, "."})
		case '(':
			s.consume("(")
			s.addToken(token{leftParen, "("})
		case ')':
			s.consume(")")
			s.addToken(token{rightParen, ")"})
		case '=':
			s.consume("=")
			s.addToken(token{equal, "="})
		case '\'':
			s.consume("'")
			s.addToken(token{quote, "'"})
		default:
			// extra space to avoid confliciton with identifier starting with "let"
			if s.match("let") {
				s.consume("let")
				s.addToken(token{let, "let"})
			} else if s.match("in") {
				s.consume("in")
				s.addToken(token{in, "in"})
			} else if t, err := s.identifier(); err != nil {
				return nil, err
			} else {
				s.addToken(t)
			}
		}
	}
	whiteSpaceCollaped := []token{}
	prevIsWhiteSpace := false
	for _, t := range s.tokens {
		if prevIsWhiteSpace {
			if t.tokenType == whiteSpace {
				continue
			} else {
				prevIsWhiteSpace = false
			}
		} else {
			if t.tokenType == whiteSpace {
				prevIsWhiteSpace = true
			}
		}
		whiteSpaceCollaped = append(whiteSpaceCollaped, t)
	}
	s.tokens = whiteSpaceCollaped
	return s.tokens, nil
}

type expression interface {
	isExpression()
	String() string
}

type binding struct {
	name  variable
	value expression
	body  expression
}

func (binding) isExpression() {}
func (b binding) String() string {
	return fmt.Sprintf("let %v = %v in %v", b.name, b.value, b.body)
}

type replBinding struct {
	name  variable
	value expression
}

func (replBinding) isExpression() {}
func (b replBinding) String() string {
	return fmt.Sprintf("let %v = %v", b.name, b.value)
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

type freeVariable struct {
	identifier string
}

func (freeVariable) isExpression() {}
func (v freeVariable) String() string {
	return fmt.Sprintf("%v", v.identifier)
}

type Parser struct {
	cur    int
	Tokens []token
}

func (p *Parser) current() token {
	if p.isEnd() {
		panic("unexpected eof")
	}
	return p.Tokens[p.cur]
}

func (p *Parser) advance() {
	p.cur += 1
}

func (p *Parser) consume(tt tokenType) {
	if p.isEnd() {
		panic(fmt.Sprintf("expect %v, but got eof", tt))
	}
	if p.current().tokenType != tt {
		panic(fmt.Sprintf("expect %v, but got %v %v", tt, p.current().tokenType, p.current().lexeme))
	}
	p.advance()
}

// func (p *Parser) consumeAll(tt tokenType) {
// 	for !p.isEnd() && p.current().tokenType == tt {
// 		p.consume(tt)
// 	}
// }

func (p *Parser) consumeMaybe(tt tokenType) {
	if !p.isEnd() && p.current().tokenType == tt {
		p.consume(tt)
	}
}

func (p *Parser) isEnd() bool {
	return p.cur >= len(p.Tokens)
}

func (p *Parser) Parse() expression {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()
	return p.expression()
}

func (p *Parser) expression() expression {
	if p.current().tokenType == quote {
		return p.replBinding()
	}
	return p.binding()
}

func (p *Parser) replBinding() expression {
	p.consume(quote)
	p.consumeMaybe(whiteSpace)
	v := p.variable()
	p.consumeMaybe(whiteSpace)
	p.consume(equal)
	p.consumeMaybe(whiteSpace)
	abs := p.abstraction()
	return replBinding{name: v, value: abs}
}

func (p *Parser) binding() expression {
	if p.current().tokenType == let {
		p.consume(let)
		p.consume(whiteSpace)
		v := p.variable()
		p.consumeMaybe(whiteSpace)
		p.consume(equal)
		p.consumeMaybe(whiteSpace)
		abs := p.abstraction()
		p.consume(whiteSpace)
		p.consume(in)
		p.consume(whiteSpace)
		body := p.binding()
		return binding{name: v, value: abs, body: body}
	}
	return p.abstraction()
}

func (p *Parser) abstraction() expression {
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

func (p *Parser) application() expression {
	expr := p.atom()
	for !p.isEnd() && p.current().tokenType == whiteSpace {
		// TODO: error handling
		if p.Tokens[p.cur+1].tokenType == in {
			return expr
		}
		p.consume(whiteSpace)
		expr = application{expr, p.atom()}
	}
	return expr
}

func (p *Parser) atom() expression {
	if p.current().tokenType == identifier {
		return p.variable()
	}
	p.consume(leftParen)
	exp := p.expression()
	p.consume(rightParen)
	return exp
}

func (p *Parser) variables() []variable {
	variables := []variable{p.variable()}
	for p.current().tokenType == whiteSpace {
		p.consume(whiteSpace)
		variables = append(variables, p.variable())
	}
	return variables
}

func (p *Parser) variable() variable {
	v := p.current().lexeme
	p.consume(identifier)
	return variable{v}
}

type envBinding struct {
	name  variable
	value expression
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

type Interpreter struct {
	Ast expression
}

func (i *Interpreter) Interpret(env environment) expression {
	return eval(i.Ast, env)
}

func eval(exp expression, env environment) expression {
	// fmt.Println(exp)
	switch exp := exp.(type) {
	case binding:
		return eval(exp.body, env.bind(exp.name, eval(exp.value, env)))
	case replBinding:
		return replBinding{name: exp.name, value: eval(exp.value, env)}
	case abstraction:
		// variable shadowing
		return abstraction{exp.param, eval(exp.expr, env.bind(exp.param, exp.param))}
	case application:
		// left := exp.left
		// right := eval(exp.right, env)
		// switch left := left.(type) {
		// case abstraction:
		// 	return eval(left.expr, env.bind(left.param, right))
		// case application:
		// 	return eval(application{eval(left, env), right}, env)
		// default:
		// 	return application{eval(left, env), right}
		// }

		left := eval(exp.left, env)
		right := eval(exp.right, env)
		switch left := left.(type) {
		case abstraction:
			return eval(left.expr, env.bind(left.param, right))
		default:
			return application{left, right}
		}
	// case freeVariable:
	// 	return exp
	case variable:
		if right, ok := env.find(exp); ok {
			return right
		}
		return freeVariable(exp)
	default:
		return exp
	}
}

func Repl() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("> ")
	env := environment{}
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			break
		}
		if len(text) == 1 {
			fmt.Print("> ")
			continue
		}
		text = text[:len(text)-1]
		scanner := Scanner{Program: []rune(text)}
		tokens, err := scanner.Scan()
		if err != nil {
			fmt.Println(err)
			fmt.Print("> ")
			continue
		}
		parser := Parser{Tokens: tokens}
		interpreter := Interpreter{Ast: parser.Parse()}
		value := interpreter.Interpret(env)
		switch v := value.(type) {
		case replBinding:
			env = env.bind(v.name, v.value)
			fmt.Printf("%v => %v\n", v.name, v.value)
		default:
			fmt.Println(value)
		}
		fmt.Print("> ")
	}
}
