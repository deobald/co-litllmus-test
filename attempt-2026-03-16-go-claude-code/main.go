package main

import (
	"container/heap"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// Token types
// ============================================================================

type TokenType int

const (
	// Literals
	TokInt TokenType = iota
	TokString
	TokTrue
	TokFalse
	TokNull

	// Keywords
	TokVar
	TokIf
	TokWhile
	TokFunction
	TokReturn
	TokYield
	TokSpawn

	// Identifiers
	TokIdent

	// Operators
	TokPlus
	TokMinus
	TokStar
	TokSlash
	TokEq
	TokNeq
	TokLt
	TokGt
	TokAssign
	TokArrowLeft  // <-
	TokArrowRight // ->

	// Punctuation
	TokLParen
	TokRParen
	TokLBrace
	TokRBrace
	TokSemicolon
	TokComma

	TokEOF
)

type Token struct {
	Type    TokenType
	Value   string
	Line    int
	Col     int
	IntVal  *big.Int // for TokInt
}

// ============================================================================
// Lexer
// ============================================================================

type Lexer struct {
	source []rune
	pos    int
	line   int
	col    int
}

func NewLexer(source string) *Lexer {
	return &Lexer{source: []rune(source), pos: 0, line: 1, col: 1}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.source) {
		return 0
	}
	return l.source[l.pos]
}

func (l *Lexer) advance() rune {
	ch := l.source[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) skipWhitespaceAndComments() {
	for l.pos < len(l.source) {
		ch := l.peek()
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			l.advance()
			continue
		}
		if ch == '/' && l.pos+1 < len(l.source) {
			next := l.source[l.pos+1]
			if next == '/' {
				// line comment
				l.advance()
				l.advance()
				for l.pos < len(l.source) && l.peek() != '\n' {
					l.advance()
				}
				continue
			}
			if next == '*' {
				// block comment
				l.advance()
				l.advance()
				for l.pos < len(l.source) {
					if l.peek() == '*' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '/' {
						l.advance()
						l.advance()
						break
					}
					l.advance()
				}
				continue
			}
		}
		break
	}
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

var keywords = map[string]TokenType{
	"null":     TokNull,
	"true":     TokTrue,
	"false":    TokFalse,
	"var":      TokVar,
	"if":       TokIf,
	"while":    TokWhile,
	"function": TokFunction,
	"return":   TokReturn,
	"yield":    TokYield,
	"spawn":    TokSpawn,
}

func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token

	for {
		l.skipWhitespaceAndComments()
		if l.pos >= len(l.source) {
			tokens = append(tokens, Token{Type: TokEOF, Line: l.line, Col: l.col})
			break
		}

		startLine, startCol := l.line, l.col
		ch := l.peek()

		switch {
		case isLetter(ch):
			start := l.pos
			for l.pos < len(l.source) && (isLetter(l.peek()) || isDigit(l.peek())) {
				l.advance()
			}
			word := string(l.source[start:l.pos])
			if tt, ok := keywords[word]; ok {
				tokens = append(tokens, Token{Type: tt, Value: word, Line: startLine, Col: startCol})
			} else {
				tokens = append(tokens, Token{Type: TokIdent, Value: word, Line: startLine, Col: startCol})
			}

		case isDigit(ch):
			start := l.pos
			for l.pos < len(l.source) && isDigit(l.peek()) {
				l.advance()
			}
			numStr := string(l.source[start:l.pos])
			n := new(big.Int)
			n.SetString(numStr, 10)
			tokens = append(tokens, Token{Type: TokInt, Value: numStr, Line: startLine, Col: startCol, IntVal: n})

		case ch == '"':
			l.advance() // skip opening quote
			var sb strings.Builder
			for l.pos < len(l.source) && l.peek() != '"' {
				c := l.advance()
				if c == '\\' && l.pos < len(l.source) {
					esc := l.advance()
					switch esc {
					case 'n':
						sb.WriteByte('\n')
					case 't':
						sb.WriteByte('\t')
					case 'r':
						sb.WriteByte('\r')
					case '"':
						sb.WriteByte('"')
					case '\\':
						sb.WriteByte('\\')
					default:
						sb.WriteByte('\\')
						sb.WriteRune(esc)
					}
				} else {
					sb.WriteRune(c)
				}
			}
			if l.pos < len(l.source) {
				l.advance() // skip closing quote
			}
			tokens = append(tokens, Token{Type: TokString, Value: sb.String(), Line: startLine, Col: startCol})

		case ch == '+':
			l.advance()
			tokens = append(tokens, Token{Type: TokPlus, Value: "+", Line: startLine, Col: startCol})
		case ch == '*':
			l.advance()
			tokens = append(tokens, Token{Type: TokStar, Value: "*", Line: startLine, Col: startCol})
		case ch == '/':
			l.advance()
			tokens = append(tokens, Token{Type: TokSlash, Value: "/", Line: startLine, Col: startCol})
		case ch == '(':
			l.advance()
			tokens = append(tokens, Token{Type: TokLParen, Value: "(", Line: startLine, Col: startCol})
		case ch == ')':
			l.advance()
			tokens = append(tokens, Token{Type: TokRParen, Value: ")", Line: startLine, Col: startCol})
		case ch == '{':
			l.advance()
			tokens = append(tokens, Token{Type: TokLBrace, Value: "{", Line: startLine, Col: startCol})
		case ch == '}':
			l.advance()
			tokens = append(tokens, Token{Type: TokRBrace, Value: "}", Line: startLine, Col: startCol})
		case ch == ';':
			l.advance()
			tokens = append(tokens, Token{Type: TokSemicolon, Value: ";", Line: startLine, Col: startCol})
		case ch == ',':
			l.advance()
			tokens = append(tokens, Token{Type: TokComma, Value: ",", Line: startLine, Col: startCol})
		case ch == '-':
			l.advance()
			if l.pos < len(l.source) && l.peek() == '>' {
				l.advance()
				tokens = append(tokens, Token{Type: TokArrowRight, Value: "->", Line: startLine, Col: startCol})
			} else {
				// Check if this is a negative number literal
				// Negative number: after certain tokens, '-' followed by digit is unary minus
				// Per spec: "There is no unary minus operator; negative numbers are parsed as signed integer literals."
				if l.pos < len(l.source) && isDigit(l.peek()) {
					// Check if previous token makes this a negative literal
					isNegLit := true
					if len(tokens) > 0 {
						prev := tokens[len(tokens)-1].Type
						// If prev is a value-producing token, this is subtraction
						if prev == TokInt || prev == TokString || prev == TokTrue || prev == TokFalse ||
							prev == TokNull || prev == TokIdent || prev == TokRParen {
							isNegLit = false
						}
					}
					if isNegLit {
						start := l.pos
						for l.pos < len(l.source) && isDigit(l.peek()) {
							l.advance()
						}
						numStr := "-" + string(l.source[start:l.pos])
						n := new(big.Int)
						n.SetString(numStr, 10)
						tokens = append(tokens, Token{Type: TokInt, Value: numStr, Line: startLine, Col: startCol, IntVal: n})
					} else {
						tokens = append(tokens, Token{Type: TokMinus, Value: "-", Line: startLine, Col: startCol})
					}
				} else {
					tokens = append(tokens, Token{Type: TokMinus, Value: "-", Line: startLine, Col: startCol})
				}
			}
		case ch == '<':
			l.advance()
			if l.pos < len(l.source) && l.peek() == '-' {
				l.advance()
				tokens = append(tokens, Token{Type: TokArrowLeft, Value: "<-", Line: startLine, Col: startCol})
			} else {
				tokens = append(tokens, Token{Type: TokLt, Value: "<", Line: startLine, Col: startCol})
			}
		case ch == '>':
			l.advance()
			tokens = append(tokens, Token{Type: TokGt, Value: ">", Line: startLine, Col: startCol})
		case ch == '=':
			l.advance()
			if l.pos < len(l.source) && l.peek() == '=' {
				l.advance()
				tokens = append(tokens, Token{Type: TokEq, Value: "==", Line: startLine, Col: startCol})
			} else {
				tokens = append(tokens, Token{Type: TokAssign, Value: "=", Line: startLine, Col: startCol})
			}
		case ch == '!':
			l.advance()
			if l.pos < len(l.source) && l.peek() == '=' {
				l.advance()
				tokens = append(tokens, Token{Type: TokNeq, Value: "!=", Line: startLine, Col: startCol})
			} else {
				return nil, fmt.Errorf("unexpected character '!' at line %d col %d", startLine, startCol)
			}
		default:
			return nil, fmt.Errorf("unexpected character '%c' at line %d col %d", ch, startLine, startCol)
		}
	}

	return tokens, nil
}

// ============================================================================
// AST
// ============================================================================

type Node interface {
	nodeType() string
}

// Expressions
type IntLiteral struct{ Value *big.Int }
type StringLiteral struct{ Value string }
type BoolLiteral struct{ Value bool }
type NullLiteral struct{}
type IdentExpr struct{ Name string }
type BinaryExpr struct {
	Op    string
	Left  Node
	Right Node
}
type CallExpr struct {
	Callee Node
	Args   []Node
}
type LambdaExpr struct {
	Params []string
	Body   []Node
}
type ReceiveExpr struct {
	Expr Node
}

// Statements
type VarStmt struct {
	Name string
	Init Node
}
type AssignStmt struct {
	Name string
	Expr Node
}
type IfStmt struct {
	Cond Node
	Body []Node
}
type WhileStmt struct {
	Cond Node
	Body []Node
}
type FuncStmt struct {
	Name   string
	Params []string
	Body   []Node
}
type ReturnStmt struct {
	Expr Node // nil means return null
}
type YieldStmt struct{}
type SpawnStmt struct {
	Expr Node
}
type SendStmt struct {
	Value   Node
	Channel Node
}
type ExprStmt struct {
	Expr Node
}

func (n *IntLiteral) nodeType() string   { return "IntLiteral" }
func (n *StringLiteral) nodeType() string { return "StringLiteral" }
func (n *BoolLiteral) nodeType() string   { return "BoolLiteral" }
func (n *NullLiteral) nodeType() string   { return "NullLiteral" }
func (n *IdentExpr) nodeType() string     { return "IdentExpr" }
func (n *BinaryExpr) nodeType() string    { return "BinaryExpr" }
func (n *CallExpr) nodeType() string      { return "CallExpr" }
func (n *LambdaExpr) nodeType() string    { return "LambdaExpr" }
func (n *ReceiveExpr) nodeType() string   { return "ReceiveExpr" }
func (n *VarStmt) nodeType() string       { return "VarStmt" }
func (n *AssignStmt) nodeType() string    { return "AssignStmt" }
func (n *IfStmt) nodeType() string        { return "IfStmt" }
func (n *WhileStmt) nodeType() string     { return "WhileStmt" }
func (n *FuncStmt) nodeType() string      { return "FuncStmt" }
func (n *ReturnStmt) nodeType() string    { return "ReturnStmt" }
func (n *YieldStmt) nodeType() string     { return "YieldStmt" }
func (n *SpawnStmt) nodeType() string     { return "SpawnStmt" }
func (n *SendStmt) nodeType() string      { return "SendStmt" }
func (n *ExprStmt) nodeType() string      { return "ExprStmt" }

// ============================================================================
// Parser
// ============================================================================

type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) peek() Token {
	return p.tokens[p.pos]
}

func (p *Parser) advance() Token {
	t := p.tokens[p.pos]
	p.pos++
	return t
}

func (p *Parser) expect(tt TokenType) Token {
	t := p.peek()
	if t.Type != tt {
		runtimeError("Parse error: expected %d but got %d ('%s') at line %d", tt, t.Type, t.Value, t.Line)
	}
	return p.advance()
}

func (p *Parser) ParseProgram() []Node {
	var stmts []Node
	for p.peek().Type != TokEOF {
		stmts = append(stmts, p.parseStmt())
	}
	return stmts
}

func (p *Parser) parseStmt() Node {
	t := p.peek()

	switch t.Type {
	case TokIf:
		return p.parseIfStmt()
	case TokWhile:
		return p.parseWhileStmt()
	case TokVar:
		return p.parseVarStmt()
	case TokYield:
		return p.parseYieldStmt()
	case TokSpawn:
		return p.parseSpawnStmt()
	case TokReturn:
		return p.parseReturnStmt()
	case TokFunction:
		// Could be function declaration or lambda expression statement
		if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == TokIdent {
			// Check if next is an identifier (not a keyword used as ident) - this is a function declaration
			return p.parseFuncStmt()
		}
		// Lambda expression as statement
		return p.parseExprOrSendStmt()
	case TokIdent:
		// Could be assignment or expression or send
		if p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].Type == TokAssign {
			return p.parseAssignStmt()
		}
		return p.parseExprOrSendStmt()
	default:
		return p.parseExprOrSendStmt()
	}
}

func (p *Parser) parseExprOrSendStmt() Node {
	expr := p.parseExpr()
	if p.peek().Type == TokArrowRight {
		p.advance() // consume ->
		channel := p.parseExpr()
		p.expect(TokSemicolon)
		return &SendStmt{Value: expr, Channel: channel}
	}
	p.expect(TokSemicolon)
	return &ExprStmt{Expr: expr}
}

func (p *Parser) parseIfStmt() Node {
	p.expect(TokIf)
	p.expect(TokLParen)
	cond := p.parseExpr()
	p.expect(TokRParen)
	p.expect(TokLBrace)
	var body []Node
	for p.peek().Type != TokRBrace {
		body = append(body, p.parseStmt())
	}
	p.expect(TokRBrace)
	return &IfStmt{Cond: cond, Body: body}
}

func (p *Parser) parseWhileStmt() Node {
	p.expect(TokWhile)
	p.expect(TokLParen)
	cond := p.parseExpr()
	p.expect(TokRParen)
	p.expect(TokLBrace)
	var body []Node
	for p.peek().Type != TokRBrace {
		body = append(body, p.parseStmt())
	}
	p.expect(TokRBrace)
	return &WhileStmt{Cond: cond, Body: body}
}

func (p *Parser) parseVarStmt() Node {
	p.expect(TokVar)
	name := p.expect(TokIdent).Value
	p.expect(TokAssign)
	expr := p.parseExpr()
	p.expect(TokSemicolon)
	return &VarStmt{Name: name, Init: expr}
}

func (p *Parser) parseAssignStmt() Node {
	name := p.expect(TokIdent).Value
	p.expect(TokAssign)
	expr := p.parseExpr()
	p.expect(TokSemicolon)
	return &AssignStmt{Name: name, Expr: expr}
}

func (p *Parser) parseYieldStmt() Node {
	p.expect(TokYield)
	p.expect(TokSemicolon)
	return &YieldStmt{}
}

func (p *Parser) parseSpawnStmt() Node {
	p.expect(TokSpawn)
	expr := p.parseExpr()
	p.expect(TokSemicolon)
	return &SpawnStmt{Expr: expr}
}

func (p *Parser) parseReturnStmt() Node {
	p.expect(TokReturn)
	if p.peek().Type == TokSemicolon {
		p.advance()
		return &ReturnStmt{Expr: nil}
	}
	expr := p.parseExpr()
	p.expect(TokSemicolon)
	return &ReturnStmt{Expr: expr}
}

func (p *Parser) parseFuncStmt() Node {
	p.expect(TokFunction)
	name := p.expect(TokIdent).Value
	p.expect(TokLParen)
	params := p.parseParams()
	p.expect(TokRParen)
	p.expect(TokLBrace)
	var body []Node
	for p.peek().Type != TokRBrace {
		body = append(body, p.parseStmt())
	}
	p.expect(TokRBrace)
	return &FuncStmt{Name: name, Params: params, Body: body}
}

func (p *Parser) parseParams() []string {
	var params []string
	if p.peek().Type == TokRParen {
		return params
	}
	params = append(params, p.expect(TokIdent).Value)
	for p.peek().Type == TokComma {
		p.advance()
		params = append(params, p.expect(TokIdent).Value)
	}
	return params
}

// Expression parsing with precedence climbing
func (p *Parser) parseExpr() Node {
	return p.parseEquality()
}

func (p *Parser) parseEquality() Node {
	left := p.parseComparison()
	for p.peek().Type == TokEq || p.peek().Type == TokNeq {
		op := p.advance().Value
		right := p.parseComparison()
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseComparison() Node {
	left := p.parseAddition()
	for p.peek().Type == TokLt || p.peek().Type == TokGt {
		op := p.advance().Value
		right := p.parseAddition()
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseAddition() Node {
	left := p.parseMultiplication()
	for p.peek().Type == TokPlus || p.peek().Type == TokMinus {
		op := p.advance().Value
		right := p.parseMultiplication()
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseMultiplication() Node {
	left := p.parseUnary()
	for p.peek().Type == TokStar || p.peek().Type == TokSlash {
		op := p.advance().Value
		right := p.parseUnary()
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseUnary() Node {
	if p.peek().Type == TokArrowLeft {
		p.advance()
		expr := p.parseUnary()
		return &ReceiveExpr{Expr: expr}
	}
	return p.parseCall()
}

func (p *Parser) parseCall() Node {
	expr := p.parsePrimary()
	for p.peek().Type == TokLParen {
		p.advance() // (
		var args []Node
		if p.peek().Type != TokRParen {
			args = append(args, p.parseExpr())
			for p.peek().Type == TokComma {
				p.advance()
				args = append(args, p.parseExpr())
			}
		}
		p.expect(TokRParen)
		expr = &CallExpr{Callee: expr, Args: args}
	}
	return expr
}

func (p *Parser) parsePrimary() Node {
	t := p.peek()

	switch t.Type {
	case TokNull:
		p.advance()
		return &NullLiteral{}
	case TokTrue:
		p.advance()
		return &BoolLiteral{Value: true}
	case TokFalse:
		p.advance()
		return &BoolLiteral{Value: false}
	case TokInt:
		p.advance()
		return &IntLiteral{Value: t.IntVal}
	case TokString:
		p.advance()
		return &StringLiteral{Value: t.Value}
	case TokIdent:
		p.advance()
		return &IdentExpr{Name: t.Value}
	case TokFunction:
		return p.parseLambdaExpr()
	case TokLParen:
		p.advance()
		expr := p.parseExpr()
		p.expect(TokRParen)
		return expr
	default:
		runtimeError("Parse error: unexpected token '%s' at line %d", t.Value, t.Line)
		return nil
	}
}

func (p *Parser) parseLambdaExpr() Node {
	p.expect(TokFunction)
	p.expect(TokLParen)
	params := p.parseParams()
	p.expect(TokRParen)
	p.expect(TokLBrace)
	var body []Node
	for p.peek().Type != TokRBrace {
		body = append(body, p.parseStmt())
	}
	p.expect(TokRBrace)
	return &LambdaExpr{Params: params, Body: body}
}

// ============================================================================
// Scope Analysis
// ============================================================================

// ScopeAnalyzer walks the AST before execution to check:
// - Variable defined before use
// - No redefinition in same scope
// - while-local vars exception
// - Function/lambda parameter validation
// - return not at global scope

type ScopeKind int

const (
	ScopeGlobal ScopeKind = iota
	ScopeFunction
)

type ScopeInfo struct {
	kind      ScopeKind
	vars      map[string]bool // variables defined in this scope
	whileVars map[string]bool // variables first defined inside a while body in this scope
	parent    *ScopeInfo
}

func newScope(kind ScopeKind, parent *ScopeInfo) *ScopeInfo {
	return &ScopeInfo{
		kind:      kind,
		vars:      make(map[string]bool),
		whileVars: make(map[string]bool),
		parent:    parent,
	}
}

func (s *ScopeInfo) isDefined(name string) bool {
	for scope := s; scope != nil; scope = scope.parent {
		if scope.vars[name] {
			return true
		}
	}
	return false
}

func (s *ScopeInfo) isDefinedInCurrent(name string) bool {
	return s.vars[name]
}

type ScopeAnalyzer struct {
	current          *ScopeInfo
	inWhile          bool              // whether we're currently inside a while body
	currentWhileVars map[string]bool   // vars first defined in the innermost while body
}

func NewScopeAnalyzer() *ScopeAnalyzer {
	global := newScope(ScopeGlobal, nil)
	// Add builtins
	global.vars["print"] = true
	global.vars["newChannel"] = true
	global.vars["newBufferedChannel"] = true
	global.vars["sleep"] = true
	global.vars["getCurrentMillis"] = true
	return &ScopeAnalyzer{current: global, currentWhileVars: make(map[string]bool)}
}

func (sa *ScopeAnalyzer) Analyze(stmts []Node) {
	for _, stmt := range stmts {
		sa.analyzeStmt(stmt)
	}
}

func (sa *ScopeAnalyzer) analyzeStmt(node Node) {
	switch n := node.(type) {
	case *VarStmt:
		// Check initializer first (before defining the var)
		sa.analyzeExpr(n.Init)
		// Check for redefinition - check ALL ancestor scopes
		if sa.current.isDefined(n.Name) {
			if sa.inWhile && sa.currentWhileVars[n.Name] {
				// while-local var from this exact while body can be redefined across iterations
				return
			}
			runtimeError("Variable already defined: %s", n.Name)
		}
		sa.current.vars[n.Name] = true
		if sa.inWhile {
			sa.currentWhileVars[n.Name] = true
		}

	case *AssignStmt:
		if !sa.current.isDefined(n.Name) {
			runtimeError("Unknown variable: %s", n.Name)
		}
		sa.analyzeExpr(n.Expr)

	case *ExprStmt:
		sa.analyzeExpr(n.Expr)

	case *IfStmt:
		sa.analyzeExpr(n.Cond)
		// if/while do NOT create new scope
		for _, stmt := range n.Body {
			sa.analyzeStmt(stmt)
		}

	case *WhileStmt:
		sa.analyzeExpr(n.Cond)
		prevInWhile := sa.inWhile
		prevWhileVars := sa.currentWhileVars
		sa.inWhile = true
		sa.currentWhileVars = make(map[string]bool)
		for _, stmt := range n.Body {
			sa.analyzeStmt(stmt)
		}
		sa.inWhile = prevInWhile
		sa.currentWhileVars = prevWhileVars

	case *FuncStmt:
		// Check param-name conflicts
		for _, param := range n.Params {
			if param == n.Name {
				runtimeError("Function name same as parameter name: %s", param)
			}
		}
		// Check duplicate params
		checkDuplicateParams(n.Params)
		// Check for redefinition in ALL scopes
		if sa.current.isDefined(n.Name) {
			if sa.inWhile && sa.currentWhileVars[n.Name] {
				// OK - while-local from this exact while body
			} else {
				runtimeError("Variable already defined: %s", n.Name)
			}
		}
		sa.current.vars[n.Name] = true
		if sa.inWhile {
			sa.currentWhileVars[n.Name] = true
		}
		// Analyze body in new function scope
		funcScope := newScope(ScopeFunction, sa.current)
		// The function name is available for self-reference
		funcScope.vars[n.Name] = true
		for _, param := range n.Params {
			funcScope.vars[param] = true
		}
		prevScope := sa.current
		prevInWhile := sa.inWhile
		prevWhileVars := sa.currentWhileVars
		sa.current = funcScope
		sa.inWhile = false
		sa.currentWhileVars = make(map[string]bool)
		for _, stmt := range n.Body {
			sa.analyzeStmt(stmt)
		}
		sa.current = prevScope
		sa.inWhile = prevInWhile
		sa.currentWhileVars = prevWhileVars

	case *ReturnStmt:
		// Check if in function scope
		if !sa.inFunctionScope() {
			runtimeError("Cannot return from global scope")
		}
		if n.Expr != nil {
			sa.analyzeExpr(n.Expr)
		}

	case *YieldStmt:
		// Valid anywhere

	case *SpawnStmt:
		sa.analyzeExpr(n.Expr)

	case *SendStmt:
		// Channel expression is evaluated first, then value
		sa.analyzeExpr(n.Channel)
		sa.analyzeExpr(n.Value)
	}
}

func (sa *ScopeAnalyzer) inFunctionScope() bool {
	for scope := sa.current; scope != nil; scope = scope.parent {
		if scope.kind == ScopeFunction {
			return true
		}
	}
	return false
}

func (sa *ScopeAnalyzer) analyzeExpr(node Node) {
	switch n := node.(type) {
	case *IntLiteral, *StringLiteral, *BoolLiteral, *NullLiteral:
		// nothing to check

	case *IdentExpr:
		if !sa.current.isDefined(n.Name) {
			runtimeError("Unknown variable: %s", n.Name)
		}

	case *BinaryExpr:
		sa.analyzeExpr(n.Left)
		sa.analyzeExpr(n.Right)

	case *CallExpr:
		sa.analyzeExpr(n.Callee)
		for _, arg := range n.Args {
			sa.analyzeExpr(arg)
		}

	case *LambdaExpr:
		checkDuplicateParams(n.Params)
		lambdaScope := newScope(ScopeFunction, sa.current)
		for _, param := range n.Params {
			lambdaScope.vars[param] = true
		}
		prevScope := sa.current
		prevInWhile := sa.inWhile
		prevWhileVars := sa.currentWhileVars
		sa.current = lambdaScope
		sa.inWhile = false
		sa.currentWhileVars = make(map[string]bool)
		for _, stmt := range n.Body {
			sa.analyzeStmt(stmt)
		}
		sa.current = prevScope
		sa.inWhile = prevInWhile
		sa.currentWhileVars = prevWhileVars

	case *ReceiveExpr:
		sa.analyzeExpr(n.Expr)
	}
}

func checkDuplicateParams(params []string) {
	// Count occurrences
	counts := make(map[string]int)
	for _, p := range params {
		counts[p]++
	}
	// Collect all occurrences of duplicated names
	var hasDups bool
	for _, c := range counts {
		if c > 1 {
			hasDups = true
			break
		}
	}
	if hasDups {
		var dups []string
		for _, p := range params {
			if counts[p] > 1 {
				dups = append(dups, p)
			}
		}
		runtimeError("Duplicate paramater names: %s", strings.Join(dups, " "))
	}
}

// ============================================================================
// Values
// ============================================================================

type ValueType int

const (
	ValNull ValueType = iota
	ValBool
	ValNumber
	ValString
	ValFunction
	ValChannel
	ValBuiltin
)

type Value struct {
	Type    ValueType
	Bool    bool
	Number  *big.Int
	Str     string
	Func    *FuncValue
	Channel *Channel
	Builtin *BuiltinFunc
}

type FuncValue struct {
	Name   string // "" for lambda
	Params []string
	Body   []Node
	Env    *Environment
}

type BuiltinFunc struct {
	Name  string
	Arity int
	Fn    func(args []*Value, interp *Interpreter) *Value
}

var valNull = &Value{Type: ValNull}
var valTrue = &Value{Type: ValBool, Bool: true}
var valFalse = &Value{Type: ValBool, Bool: false}

func numVal(n *big.Int) *Value {
	return &Value{Type: ValNumber, Number: new(big.Int).Set(n)}
}

func strVal(s string) *Value {
	return &Value{Type: ValString, Str: s}
}

func boolVal(b bool) *Value {
	if b {
		return valTrue
	}
	return valFalse
}

func (v *Value) String() string {
	switch v.Type {
	case ValNull:
		return "null"
	case ValBool:
		if v.Bool {
			return "true"
		}
		return "false"
	case ValNumber:
		return v.Number.String()
	case ValString:
		return v.Str
	case ValFunction:
		if v.Func.Name == "" {
			return "function <lambda>"
		}
		return "function " + v.Func.Name
	case ValChannel:
		return "Channel"
	case ValBuiltin:
		return "function " + v.Builtin.Name
	default:
		return "unknown"
	}
}

// ErrorString returns the value representation for use in error messages.
// Strings are shown with surrounding quotes.
func (v *Value) ErrorString() string {
	if v.Type == ValString {
		return "\"" + v.Str + "\""
	}
	return v.String()
}

func (v *Value) IsTruthy() bool {
	if v.Type == ValNull {
		return false
	}
	if v.Type == ValBool && !v.Bool {
		return false
	}
	return true
}

func (v *Value) TypeName() string {
	switch v.Type {
	case ValNull:
		return "null"
	case ValBool:
		return "Boolean"
	case ValNumber:
		return "Number"
	case ValString:
		return "String"
	case ValFunction, ValBuiltin:
		return "Function"
	case ValChannel:
		return "Channel"
	default:
		return "unknown"
	}
}

func valuesEqual(a, b *Value) bool {
	if a.Type == ValNull && b.Type == ValNull {
		return true
	}
	if a.Type == ValNull || b.Type == ValNull {
		return false
	}
	if a.Type != b.Type {
		return false
	}
	switch a.Type {
	case ValBool:
		return a.Bool == b.Bool
	case ValNumber:
		return a.Number.Cmp(b.Number) == 0
	case ValString:
		return a.Str == b.Str
	default:
		return false // functions and channels always false
	}
}

// ============================================================================
// Environment
// ============================================================================

type Environment struct {
	vars   map[string]*Value
	parent *Environment
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		vars:   make(map[string]*Value),
		parent: parent,
	}
}

func (e *Environment) Get(name string) (*Value, bool) {
	if v, ok := e.vars[name]; ok {
		return v, true
	}
	if e.parent != nil {
		return e.parent.Get(name)
	}
	return nil, false
}

func (e *Environment) Set(name string, val *Value) bool {
	if _, ok := e.vars[name]; ok {
		e.vars[name] = val
		return true
	}
	if e.parent != nil {
		return e.parent.Set(name, val)
	}
	return false
}

func (e *Environment) Define(name string, val *Value) {
	e.vars[name] = val
}

// ============================================================================
// Channel
// ============================================================================

type Channel struct {
	mu          sync.Mutex
	buffer      []*Value
	capacity    int
	sendQueue   []*ChannelWaiter
	recvQueue   []*ChannelWaiter
}

type ChannelWaiter struct {
	value     *Value // for senders: value being sent; for receivers: will be set
	coroutine *Coroutine
}

func NewChannel(capacity int) *Channel {
	return &Channel{
		capacity: capacity,
	}
}

// ============================================================================
// Coroutine and Scheduler
// ============================================================================

type Coroutine struct {
	id        int
	fn        func()
	wakeCh    chan struct{}  // scheduler signals coroutine to run
	yieldCh   chan struct{}  // coroutine signals it's yielding back
	result    chan *Value    // for channel operations
	done      bool
	isMain    bool
	schedTime time.Time
	index     int // for heap
}

type SchedulerQueue []*Coroutine

func (sq SchedulerQueue) Len() int { return len(sq) }
func (sq SchedulerQueue) Less(i, j int) bool {
	return sq[i].schedTime.Before(sq[j].schedTime)
}
func (sq SchedulerQueue) Swap(i, j int) {
	sq[i], sq[j] = sq[j], sq[i]
	sq[i].index = i
	sq[j].index = j
}
func (sq *SchedulerQueue) Push(x interface{}) {
	n := len(*sq)
	c := x.(*Coroutine)
	c.index = n
	*sq = append(*sq, c)
}
func (sq *SchedulerQueue) Pop() interface{} {
	old := *sq
	n := len(old)
	c := old[n-1]
	old[n-1] = nil
	c.index = -1
	*sq = old[:n-1]
	return c
}

// ============================================================================
// Interpreter
// ============================================================================

type ReturnSignal struct {
	Value *Value
}

type Interpreter struct {
	globalEnv    *Environment
	scheduler    *SchedulerQueue
	coroutineID  int
	currentCoro  *Coroutine
	allCoros     []*Coroutine
}

func NewInterpreter() *Interpreter {
	env := NewEnvironment(nil)
	interp := &Interpreter{
		globalEnv: env,
	}
	sq := &SchedulerQueue{}
	heap.Init(sq)
	interp.scheduler = sq

	// Define builtins
	env.Define("print", &Value{Type: ValBuiltin, Builtin: &BuiltinFunc{
		Name: "print", Arity: 1,
		Fn: func(args []*Value, interp *Interpreter) *Value {
			fmt.Println(args[0].String())
			return valNull
		},
	}})
	env.Define("newChannel", &Value{Type: ValBuiltin, Builtin: &BuiltinFunc{
		Name: "newChannel", Arity: 0,
		Fn: func(args []*Value, interp *Interpreter) *Value {
			return &Value{Type: ValChannel, Channel: NewChannel(0)}
		},
	}})
	env.Define("newBufferedChannel", &Value{Type: ValBuiltin, Builtin: &BuiltinFunc{
		Name: "newBufferedChannel", Arity: 1,
		Fn: func(args []*Value, interp *Interpreter) *Value {
			arg := args[0]
			if arg.Type != ValNumber || arg.Number.Sign() < 0 {
				runtimeError("newBufferedChannel argument is not a non-negative number: %s", arg.ErrorString())
			}
			cap := int(arg.Number.Int64())
			return &Value{Type: ValChannel, Channel: NewChannel(cap)}
		},
	}})
	env.Define("sleep", &Value{Type: ValBuiltin, Builtin: &BuiltinFunc{
		Name: "sleep", Arity: 1,
		Fn: func(args []*Value, interp *Interpreter) *Value {
			arg := args[0]
			if arg.Type != ValNumber || arg.Number.Sign() < 0 {
				runtimeError("sleep argument is not a non-negative number: %s", arg.ErrorString())
			}
			ms := arg.Number.Int64()
			if ms > 0 {
				interp.sleepCoroutine(time.Duration(ms) * time.Millisecond)
			} else {
				// sleep(0) just yields
				interp.yieldCoroutine()
			}
			return valNull
		},
	}})
	env.Define("getCurrentMillis", &Value{Type: ValBuiltin, Builtin: &BuiltinFunc{
		Name: "getCurrentMillis", Arity: 0,
		Fn: func(args []*Value, interp *Interpreter) *Value {
			ms := time.Now().UnixMilli()
			return numVal(big.NewInt(ms))
		},
	}})

	return interp
}

func runtimeError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}

func (interp *Interpreter) Run(stmts []Node) {
	// Create main coroutine
	mainCoro := &Coroutine{
		id:     0,
		wakeCh:  make(chan struct{}, 1),
		yieldCh: make(chan struct{}, 1),
		result:  make(chan *Value, 1),
		isMain:  true,
	}
	interp.currentCoro = mainCoro
	interp.coroutineID = 1

	// Run main coroutine in a goroutine
	mainDone := make(chan struct{})
	go func() {
		// Wait for start signal
		<-mainCoro.wakeCh
		interp.execStmts(stmts, interp.globalEnv)
		mainCoro.done = true
		mainCoro.yieldCh <- struct{}{} // signal scheduler
		close(mainDone)
	}()

	// Start main
	mainCoro.wakeCh <- struct{}{}
	// Wait for main to yield or finish
	<-mainCoro.yieldCh

	// Scheduler loop
	for {
		if mainCoro.done && interp.scheduler.Len() == 0 {
			break
		}

		if interp.scheduler.Len() == 0 {
			// No runnable coroutines
			if !mainCoro.done {
				// Main is blocked on something
				runtimeError("Deadlock: main coroutine blocked on channel with no runnable coroutines")
			}
			break
		}

		// Get next coroutine
		coro := heap.Pop(interp.scheduler).(*Coroutine)

		// Wait for scheduled time
		now := time.Now()
		if coro.schedTime.After(now) {
			time.Sleep(coro.schedTime.Sub(now))
		}

		if coro.done {
			continue
		}

		interp.currentCoro = coro
		coro.wakeCh <- struct{}{} // wake coroutine
		<-coro.yieldCh            // wait for it to yield back
	}

	<-mainDone // ensure main goroutine is done
}

func (interp *Interpreter) yieldCoroutine() {
	coro := interp.currentCoro
	coro.schedTime = time.Now()
	heap.Push(interp.scheduler, coro)
	coro.yieldCh <- struct{}{} // signal scheduler
	<-coro.wakeCh              // wait for scheduler to wake us
}

func (interp *Interpreter) sleepCoroutine(d time.Duration) {
	coro := interp.currentCoro
	coro.schedTime = time.Now().Add(d)
	heap.Push(interp.scheduler, coro)
	coro.yieldCh <- struct{}{} // signal scheduler
	<-coro.wakeCh              // wait for scheduler to wake us
}

func (interp *Interpreter) spawnCoroutine(fn func()) {
	coro := &Coroutine{
		id:        interp.coroutineID,
		wakeCh:    make(chan struct{}, 1),
		yieldCh:   make(chan struct{}, 1),
		result:    make(chan *Value, 1),
		schedTime: time.Now(),
	}
	interp.coroutineID++
	interp.allCoros = append(interp.allCoros, coro)

	go func() {
		<-coro.wakeCh
		// Save/restore current coroutine in the goroutine
		fn()
		coro.done = true
		coro.yieldCh <- struct{}{}
	}()

	heap.Push(interp.scheduler, coro)
}

func (interp *Interpreter) sendToChannel(ch *Channel, val *Value) {
	ch.mu.Lock()

	// Priority 1: if a receiver is waiting, deliver directly
	if len(ch.recvQueue) > 0 {
		waiter := ch.recvQueue[0]
		ch.recvQueue = ch.recvQueue[1:]
		waiter.value = val
		// Schedule the receiver
		waiter.coroutine.schedTime = time.Now()
		heap.Push(interp.scheduler, waiter.coroutine)
		ch.mu.Unlock()
		return
	}

	// Priority 2: if buffer not full, enqueue
	if len(ch.buffer) < ch.capacity {
		ch.buffer = append(ch.buffer, val)
		ch.mu.Unlock()
		return
	}

	// Priority 3: block sender
	if len(ch.sendQueue) >= 4 {
		ch.mu.Unlock()
		runtimeError("Channel send queue is full")
	}

	coro := interp.currentCoro
	waiter := &ChannelWaiter{value: val, coroutine: coro}
	ch.sendQueue = append(ch.sendQueue, waiter)
	ch.mu.Unlock()

	// Block: yield to scheduler and wait
	coro.yieldCh <- struct{}{}
	<-coro.wakeCh
}

func (interp *Interpreter) receiveFromChannel(ch *Channel) *Value {
	ch.mu.Lock()

	// Priority 1: pending senders (unbuffered) - take value directly
	if len(ch.sendQueue) > 0 && len(ch.buffer) == 0 {
		waiter := ch.sendQueue[0]
		ch.sendQueue = ch.sendQueue[1:]
		val := waiter.value
		// Schedule the sender
		waiter.coroutine.schedTime = time.Now()
		heap.Push(interp.scheduler, waiter.coroutine)
		ch.mu.Unlock()
		return val
	}

	// Priority 2: pending senders and buffer has values
	if len(ch.sendQueue) > 0 && len(ch.buffer) > 0 {
		val := ch.buffer[0]
		ch.buffer = ch.buffer[1:]
		// Move sender's value to buffer
		waiter := ch.sendQueue[0]
		ch.sendQueue = ch.sendQueue[1:]
		ch.buffer = append(ch.buffer, waiter.value)
		// Schedule the sender
		waiter.coroutine.schedTime = time.Now()
		heap.Push(interp.scheduler, waiter.coroutine)
		ch.mu.Unlock()
		return val
	}

	// Priority 3: buffer has values, no senders
	if len(ch.buffer) > 0 {
		val := ch.buffer[0]
		ch.buffer = ch.buffer[1:]
		ch.mu.Unlock()
		return val
	}

	// Priority 4: block receiver
	if len(ch.recvQueue) >= 4 {
		ch.mu.Unlock()
		runtimeError("Channel receive queue is full")
	}

	coro := interp.currentCoro
	waiter := &ChannelWaiter{coroutine: coro}
	ch.recvQueue = append(ch.recvQueue, waiter)
	ch.mu.Unlock()

	// Block: yield to scheduler and wait
	coro.yieldCh <- struct{}{}
	<-coro.wakeCh

	return waiter.value
}

func (interp *Interpreter) execStmts(stmts []Node, env *Environment) {
	for _, stmt := range stmts {
		interp.execStmt(stmt, env)
	}
}

func (interp *Interpreter) execStmt(node Node, env *Environment) {
	switch n := node.(type) {
	case *VarStmt:
		val := interp.evalExpr(n.Init, env)
		env.Define(n.Name, val)

	case *AssignStmt:
		val := interp.evalExpr(n.Expr, env)
		if !env.Set(n.Name, val) {
			runtimeError("Unknown variable: %s", n.Name)
		}

	case *ExprStmt:
		interp.evalExpr(n.Expr, env)

	case *IfStmt:
		cond := interp.evalExpr(n.Cond, env)
		if cond.IsTruthy() {
			interp.execStmts(n.Body, env)
		}

	case *WhileStmt:
		for {
			cond := interp.evalExpr(n.Cond, env)
			if !cond.IsTruthy() {
				break
			}
			interp.execStmts(n.Body, env)
		}

	case *FuncStmt:
		fv := &FuncValue{
			Name:   n.Name,
			Params: n.Params,
			Body:   n.Body,
			Env:    env,
		}
		val := &Value{Type: ValFunction, Func: fv}
		env.Define(n.Name, val)

	case *ReturnStmt:
		var val *Value
		if n.Expr != nil {
			val = interp.evalExpr(n.Expr, env)
		} else {
			val = valNull
		}
		panic(ReturnSignal{Value: val})

	case *YieldStmt:
		interp.yieldCoroutine()

	case *SpawnStmt:
		// Evaluate the expression in a new coroutine
		// The spawn captures current env
		capturedEnv := env
		expr := n.Expr
		interp.spawnCoroutine(func() {
			interp.evalExpr(expr, capturedEnv)
		})

	case *SendStmt:
		// Channel is evaluated first, then value
		chVal := interp.evalExpr(n.Channel, env)
		if chVal.Type != ValChannel {
			runtimeError("Cannot send to a non-channel: %s", chVal.String())
		}
		val := interp.evalExpr(n.Value, env)
		interp.sendToChannel(chVal.Channel, val)
	}
}

func (interp *Interpreter) evalExpr(node Node, env *Environment) *Value {
	switch n := node.(type) {
	case *IntLiteral:
		return numVal(n.Value)
	case *StringLiteral:
		return strVal(n.Value)
	case *BoolLiteral:
		return boolVal(n.Value)
	case *NullLiteral:
		return valNull
	case *IdentExpr:
		val, ok := env.Get(n.Name)
		if !ok {
			runtimeError("Unknown variable: %s", n.Name)
		}
		return val

	case *BinaryExpr:
		left := interp.evalExpr(n.Left, env)
		right := interp.evalExpr(n.Right, env)
		return interp.evalBinaryOp(n.Op, left, right)

	case *CallExpr:
		callee := interp.evalExpr(n.Callee, env)
		var args []*Value
		for _, arg := range n.Args {
			args = append(args, interp.evalExpr(arg, env))
		}
		return interp.callFunction(callee, args, n.Callee)

	case *LambdaExpr:
		return &Value{Type: ValFunction, Func: &FuncValue{
			Name:   "",
			Params: n.Params,
			Body:   n.Body,
			Env:    env,
		}}

	case *ReceiveExpr:
		chVal := interp.evalExpr(n.Expr, env)
		if chVal.Type != ValChannel {
			runtimeError("Cannot receive from a non-channel: %s", chVal.String())
		}
		return interp.receiveFromChannel(chVal.Channel)

	default:
		runtimeError("Unknown expression type: %T", node)
		return nil
	}
}

func (interp *Interpreter) callFunction(callee *Value, args []*Value, calleeExpr Node) *Value {
	switch callee.Type {
	case ValBuiltin:
		b := callee.Builtin
		if len(args) != b.Arity {
			runtimeError("%s called with wrong number of arguments: got %d, expected %d", b.Name, len(args), b.Arity)
		}
		return b.Fn(args, interp)

	case ValFunction:
		f := callee.Func
		if len(args) != len(f.Params) {
			name := f.Name
			if name == "" {
				name = "<lambda>"
			}
			runtimeError("%s called with wrong number of arguments: got %d, expected %d", name, len(args), len(f.Params))
		}
		funcEnv := NewEnvironment(f.Env)
		// Self-reference for named functions
		if f.Name != "" {
			funcEnv.Define(f.Name, callee)
		}
		for i, param := range f.Params {
			funcEnv.Define(param, args[i])
		}
		// Execute body, catch return
		var result *Value
		func() {
			defer func() {
				if r := recover(); r != nil {
					if ret, ok := r.(ReturnSignal); ok {
						result = ret.Value
					} else {
						panic(r)
					}
				}
			}()
			interp.execStmts(f.Body, funcEnv)
		}()
		if result != nil {
			return result
		}
		return valNull

	default:
		// Format callee description
		desc := describeCallee(calleeExpr)
		runtimeError("Cannot call a non-function: %s is %s", desc, callee.String())
		return nil
	}
}

func describeCallee(node Node) string {
	switch n := node.(type) {
	case *IdentExpr:
		return fmt.Sprintf("Variable \"%s\"", n.Name)
	default:
		return "Expression"
	}
}

func (interp *Interpreter) evalBinaryOp(op string, left, right *Value) *Value {
	switch op {
	case "+":
		if left.Type == ValNumber && right.Type == ValNumber {
			result := new(big.Int).Add(left.Number, right.Number)
			return numVal(result)
		}
		if left.Type == ValString && right.Type == ValString {
			return strVal(left.Str + right.Str)
		}
		if left.Type == ValString {
			return strVal(left.Str + right.String())
		}
		if right.Type == ValString {
			return strVal(left.String() + right.Str)
		}
		runtimeError("Cannot add or append: %s and %s", left.String(), right.String())

	case "-":
		if left.Type != ValNumber || right.Type != ValNumber {
			runtimeError("Cannot subtract non-numbers: %s and %s", left.String(), right.String())
		}
		return numVal(new(big.Int).Sub(left.Number, right.Number))

	case "*":
		if left.Type != ValNumber || right.Type != ValNumber {
			runtimeError("Cannot multiply non-numbers: %s and %s", left.String(), right.String())
		}
		return numVal(new(big.Int).Mul(left.Number, right.Number))

	case "/":
		if left.Type != ValNumber || right.Type != ValNumber {
			runtimeError("Cannot divide non-numbers: %s and %s", left.String(), right.String())
		}
		if right.Number.Sign() == 0 {
			runtimeError("Division by zero")
		}
		// Integer division truncated toward negative infinity
		result := new(big.Int)
		result.Div(left.Number, right.Number) // Go's big.Int Div truncates toward negative infinity (Euclidean)
		// Actually big.Int.Div is Euclidean division. We want floor division.
		// big.Int.Quo truncates toward zero. We want toward negative infinity.
		// Use Div which is floor division (toward -inf) for big.Int... let me verify.
		// Actually big.Int.Div is "Euclidean" - result is >= 0 for positive divisor.
		// big.Int.Quo truncates toward zero.
		// Floor division: we need to use Quo and then adjust.
		// Let me just use Quo and adjust manually.
		result.Quo(left.Number, right.Number)
		// If there's a remainder and the signs differ, subtract 1
		mod := new(big.Int).Rem(left.Number, right.Number)
		if mod.Sign() != 0 && (left.Number.Sign() < 0) != (right.Number.Sign() < 0) {
			result.Sub(result, big.NewInt(1))
		}
		return numVal(result)

	case "<":
		if left.Type != ValNumber || right.Type != ValNumber {
			runtimeError("Cannot compare non-numbers: %s and %s", left.String(), right.String())
		}
		return boolVal(left.Number.Cmp(right.Number) < 0)

	case ">":
		if left.Type != ValNumber || right.Type != ValNumber {
			runtimeError("Cannot compare non-numbers: %s and %s", left.String(), right.String())
		}
		return boolVal(left.Number.Cmp(right.Number) > 0)

	case "==":
		return boolVal(valuesEqual(left, right))

	case "!=":
		return boolVal(!valuesEqual(left, right))
	}

	runtimeError("Unknown operator: %s", op)
	return nil
}

// ============================================================================
// Reserved word check for identifiers
// ============================================================================

var reservedWords = map[string]bool{
	"null": true, "true": true, "false": true,
	"var": true, "if": true, "while": true,
	"function": true, "return": true, "yield": true, "spawn": true,
}

// ============================================================================
// Main
// ============================================================================

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: co <filename>")
		os.Exit(1)
	}

	filename := os.Args[1]
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Cannot read file: %s\n", err)
		os.Exit(1)
	}

	lexer := NewLexer(string(source))
	tokens, err := lexer.Tokenize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}

	// Check for reserved words used as identifiers in var/function declarations
	checkReservedWordUsage(tokens)

	parser := NewParser(tokens)
	program := parser.ParseProgram()

	// Scope analysis
	analyzer := NewScopeAnalyzer()
	analyzer.Analyze(program)

	// Execute
	interp := NewInterpreter()
	interp.Run(program)
}

func checkReservedWordUsage(tokens []Token) {
	for i := 0; i < len(tokens); i++ {
		if tokens[i].Type == TokVar && i+1 < len(tokens) {
			next := tokens[i+1]
			if reservedWords[next.Value] && next.Type != TokIdent {
				runtimeError("Reserved word cannot be used as identifier: %s", next.Value)
			}
		}
	}
}
