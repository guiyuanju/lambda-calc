from enum import Enum
import sys

Token = Enum('Token',
 ['LAMBDA', 'DOT', 'LEFT_PAREN', 'RIGHT_PAREN', 'ID'])

Term = Enum('Term', ['ABS', 'APP', 'VAR'])

def scan(text):
    tokens = []
    for c in text:
        match c:
            case ' ':
                continue
            case '𝛌' | '\\':
                tokens.append((Token.LAMBDA, ''))
            case '.':
                tokens.append((Token.DOT, ''))
            case '(':
                tokens.append((Token.LEFT_PAREN, ''))
            case ')':
                tokens.append((Token.RIGHT_PAREN, ''))
            case _:
                tokens.append((Token.ID, c))
    return tokens

def parse(tokens):
    i = 0
    def var():
        nonlocal i
        res = (Term.VAR, tokens[i][1])
        i += 1
        return res
    def abs():
        nonlocal i
        i += 1
        v = var()
        i += 1
        return (Term.ABS, v, term())
    def term():
        nonlocal i
        while i < len(tokens):
            match tokens[i][0]:
                case Token.ID:
                    return var()
                case Token.LEFT_PAREN:
                    if tokens[i + 1][0] == Token.LAMBDA:
                        i += 1
                        res = abs()
                        i += 1
                        return res
                    else:
                        i += 1
                        res1 = term()
                        res2 = term()
                        i += 1
                        return (Term.APP, res1, res2)
    return term()

def interpret(ast, env):
    def find(env, var):
        for d in reversed(env):
            if var in d:
                return (True, d[var])
        return (False, '')
    match ast[0]:
        case Term.VAR:
            founded, res = find(env, ast)
            if founded:
                return res
            return ast
        case Term.ABS:
            # env + ...: variable shaowing
            return (ast[0], ast[1], interpret(ast[2], env + [{ast[1]: ast[1]}]))
        case Term.APP:
            if ast[1][0] == Term.ABS:
                return interpret(ast[1][2], env + [{ast[1][1]: interpret(ast[2], env)}])
            return ast

def textify(term):
    match term[0]:
        case Term.VAR:
            return term[1]
        case Term.ABS:
            return f"(𝛌{term[1][1]}.{textify(term[2])})"
        case Term.APP:
            return f"({textify(term[1])}{textify(term[2])})"

def repl():
    print("> ", end = "", flush = True)
    for line in sys.stdin:
        print(textify(interpret(parse(scan(line)), [])))
        print("> ", end = "", flush = True)
    
repl()
    
# program = "((𝛌x.(𝛌y.x))z)"
# lexmes = scan(program)
# ast = parse(lexmes)
# value = interpret(ast, [])
# print(textify(value))
# print(textify(ast))