package parser

import (
	"fmt"
	"simpyl/ast"
	"simpyl/lexer"
	"simpyl/token"
	"strconv"
)

type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	errors    []string
	spacing   int
	skipFlag  bool

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // list[index]
)

var precedences = map[token.TokenType]int{
	token.IN:       EQUALS,
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.DOT:      CALL,
	token.LBRACKET: INDEX,
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:        l,
		errors:   []string{},
		spacing:  0,
		skipFlag: true,
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.LBRACKET, p.parseListLiteral)
	p.registerPrefix(token.LBRACE, p.parseDictLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.DOT, p.parseObjectMethod)
	p.registerInfix(token.IN, p.parseInExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		if p.skipFlag {
			p.nextToken()
		} else {
			p.skipFlag = true
		}
	}
	return program
}

/*
Statement Parsing
*/

func (p *Parser) parseStatement() ast.Statement {
	switch {
	case p.curToken.Type == token.TAB || p.curToken.Type == token.NEWLINE:
		p.parseSpacing()
		return p.parseStatement()
		// Error if before EOF
		// Error if inside let statement
		// Error if before semicolon

	case p.curToken.Type == token.LET:
		return p.parseLetStatement()

	case p.curToken.Type == token.IDENT && p.peekToken.Type == token.ASSIGN:
		return p.parseIdentStatement()

	case p.curToken.Type == token.RETURN:
		return p.parseReturnStatement()

	case p.curToken.Type == token.FUNCTION:
		return p.parseFunctionStatement()

	case p.curToken.Type == token.FOR:
		return p.parseForStatement()

	case p.curToken.Type == token.WHILE:
		return p.parseWhileStatement()

	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseIdentStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: token.Token{Type: token.LET, Literal: string("let")}}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseFunctionStatement() *ast.FunctionStatement {
	lit := &ast.FunctionStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	lit.Name = p.curToken.Literal

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()
	if !p.expectPeek(token.COLON) {
		return nil
	}

	lit.Body = p.parseBlockStatement()
	return lit
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	loop := &ast.ForStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	loop.Iterator = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.IN) {
		return nil
	}
	p.nextToken()

	loop.Iterable = p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}

	loop.Body = p.parseBlockStatement()
	return loop
}

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	loop := &ast.WhileStatement{Token: p.curToken}

	p.nextToken()
	loop.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}

	loop.Body = p.parseBlockStatement()
	return loop
}

/*
Expression Parsing
*/

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {

	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {

	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {

	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {

	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseListLiteral() ast.Expression {
	array := &ast.ListLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

func (p *Parser) parseDictLiteral() ast.Expression {
	dict := &ast.DictLiteral{Token: p.curToken}

	dict.Pairs = make(map[ast.Expression]ast.Expression)
	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		p.advanceWhitespace()

		key := p.parseExpression(LOWEST)
		p.advanceWhitespace()
		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken()
		p.advanceWhitespace()

		value := p.parseExpression(LOWEST)
		p.advanceWhitespace()
		dict.Pairs[key] = value
		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return dict
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.COLON) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()
	p.advanceWhitespace()

	if p.curTokenIs(token.ELSE) {
		p.skipFlag = true
		p.nextToken()
		if !p.curTokenIs(token.COLON) {
			return nil
		}
		p.nextToken()
		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	indentLevel := p.spacing

	p.nextToken()
	p.advanceWhitespace()

	for p.spacing > indentLevel && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		if p.skipFlag {
			p.nextToken()
		} else {
			p.skipFlag = true
		}
		p.advanceWhitespace()
	}

	p.skipFlag = false

	return block
}

func (p *Parser) advanceWhitespace() {
	for p.curTokenIs(token.TAB) || p.curTokenIs(token.NEWLINE) {
		p.parseSpacing()
	}
}

func (p *Parser) parseSpacing() {
	switch p.curToken.Type {
	case token.TAB:
		p.spacing += 1
		p.nextToken()

	case token.NEWLINE:
		p.spacing = 0
		p.nextToken()
	default:
		msg := fmt.Sprintf("expected next token to be %s, got %s instead",
			token.TAB, p.peekToken.Type)
		p.errors = append(p.errors, msg)
	}
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)

	return exp
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if p.expectPeek(token.COLON) {
		p.nextToken()
		exp.Colon = true
		exp.EndIndex = p.parseExpression(LOWEST)
	} else {
		exp.Colon = false
	}

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	if p.expectPeek(token.ASSIGN) {
		p.nextToken()
		stmt := &ast.IndexAssignExpression{Token: exp.Token, Left: exp.Left, Index: exp.Index,
			Colon: exp.Colon, EndIndex: exp.EndIndex}
		stmt.Value = p.parseExpression(LOWEST)
		return stmt
	}
	return exp
}

func (p *Parser) parseObjectMethod(left ast.Expression) ast.Expression {
	expression := &ast.ObjectMethod{Obj: left, Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	expression.Method = p.parseExpression(LOWEST)

	return expression
}

func (p *Parser) parseInExpression(left ast.Expression) ast.Expression {
	expression := &ast.InExpression{Token: p.curToken, Left: left}

	p.nextToken()
	expression.Right = p.parseExpression(LOWEST)

	return expression
}

/*
Prefix Parsing
*/

func (p *Parser) parsePrefixExpression() ast.Expression {

	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

/*
Infix Parsing
*/

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {

	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

/*
Token functions
*/

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		return false
	}
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

/*
Error functions
*/

func (p *Parser) Errors() []string {
	return p.errors
}

/*
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}
*/

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}
