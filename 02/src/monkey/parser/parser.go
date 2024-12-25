package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
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
)

// precedences 연산자 우선순위 맵
// 연산 토큰과 연산자 우선순위를 매핑한다.
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
}

// 각각의 토큰 타입은 토큰이 전위 연산자로 쓰였는지 혹은 중위 연산자로 쓰였는지에 따라 다르게 처리된다.
// 이 실습에서는 후위 연산자는 생략한다.
type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	// 토큰 타입에 따라 어떤 파싱 함수를 호출할지 결정하는 맵
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

// New 파서를 생성한다.
// 이 파서는 프랫 파서로, 특정 파싱 함수를 특정 토큰과 연관짓는다.
// 예를 들어 A 라는 토큰을 만나면 A 를 파싱하는 함수를 호출하고, ast 노드를 반환한다.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)       // 식별자
	p.registerPrefix(token.INT, p.parseIntegerLiteral)     // 정수 리터럴
	p.registerPrefix(token.BANG, p.parsePrefixExpression)  // 전위 연산자
	p.registerPrefix(token.MINUS, p.parsePrefixExpression) // 전위 연산자
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)

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

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

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

// expectPeek 다음 토큰이 t 타입인지 확인하고, 맞다면 토큰을 진행시킨다.
// 만약 기대하는 타입이 아니라면 에러를 기록하고 false 를 반환한다.
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// noPrefixParseFnError 전위 연산자 파싱 함수가 없을 때 에러를 기록한다.
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// ParseProgram 파서 entrypoint 가 되며 프로그램의 모든 문장을 파싱한다. (AST 생성)
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

// parseStatement 토큰 타입에 따라 관련 파서 메서드를 호출한다.
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
		// 1 + 2 + 3; 같은 표현식문을 파싱한다면 가정하면, AST 는 ((1 + 2) + 3) 이 된다.
		// 이 표현식을 파싱하기 위해 parseExpressionStatement 를 호출한다.
	default:
		return p.parseExpressionStatement()
	}
}

// parseLetStatement let 문을 파싱한다.
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	// 1. 내부적으로 nextToken 을 호출하여 토큰을 진행시키고, token.IDENT 를 기대한다.
	// 즉 여기선 'let' 이 올 것을 기대함
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	// 2. 실제 변수명을 파싱하여 Identifier 노드를 생성한다.
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	// 3. 내부적으로 nextToken 을 호출하여 토큰을 진행시키고, token.ASSIGN 을 기대한다.
	// 즉, = 이 올 것을 기대함.
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	// 4. 내부적으로 nextToken 을 호출하여 토큰을 진행시키고, 변수의 값을 파싱한다. (우항 부분)
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST) // 표현식 파싱의 시작점이므로 우선순위를 가장 낮게 설정한다. README 참고

	// 다음 토큰이 세미콜론이면 다음 토큰으로 진행한다.
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseReturnStatement return 문을 파싱한다.
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	// 1. 토큰 타입을 읽어 statment 노드를 생성한 뒤,
	stmt := &ast.ReturnStatement{Token: p.curToken}

	// 2. 다음 토큰으로 진행한다.
	p.nextToken()

	// 3. 우항을 읽는다.
	stmt.ReturnValue = p.parseExpression(LOWEST) // 표현식 파싱의 시작점이므로 우선순위를 가장 낮게 설정한다. README 참고

	// 4. 마지막 토큰이 세미콜론임을 확인하고, 세미콜론이라면 다음 토큰으로 진행한다.
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseExpressionStatement 표현식문을 파싱하기 위한 Entrypoint 가 되는 함수
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	// 1 + 2 + 3; 같은 표현식문을 파싱한다면 가정하면, curToken 은 1 이고 peekToken 은 + 이다.
	// 이 상태에서 parseExpression 을 호출하면, ->
	stmt.Expression = p.parseExpression(LOWEST) // 표현식 파싱의 시작점이므로 우선순위를 가장 낮게 설정한다. README 참고

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseExpression 표현식을 실제 파싱한다.
// precedence 는 연산자 우선순위를 나타낸다. (함수를 호출한 쪽에서만 알고 있는 우선순위를 전달해주는 것임)
func (p *Parser) parseExpression(precedence int) ast.Expression {
	// 1 + 2 + 3; 에서 curToken 이 1 이므로 parseIntegerLiteral 이 호출된다.
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		// 연관된 전위 연산자 파싱 함수가 없음 -> 에러 기록 후 return nil
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	// parseIntegerLiteral 호출의 결과로 *ast.IntegerLiteral 노드가 반환된다.
	leftExp := prefix()

	// for 문 조건: 다음 토큰이 세미콜론이 아니고 (=아직 표현식이 끝나지 않았고), 다음 토큰의 우선순위가 더 높다면!
	// 1 + 2 + 3; 에서 다음 토큰인 + 연산자의 우선순위는 SUM 이고 현재 precedence 는 LOWEST 이므로 for문 조건에 해당된다.
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		// 1 + 2 + 3; 에서 peekToken 의 타입은 + 이므로 parseInfixExpression 이 infix 에 할당된다.
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		// leftExp 에 저장하기 전에, 다음 토큰으로 진행한다.
		// (1번째 반복) 진행 후 curToken 은 + 이고 peekToken 은 2 이다.
		// (2번째 반복) 진행 후 curToken 은 + 이고 peekToken 은 3 이다.
		p.nextToken()

		// (1번째 반복) leftExp = 1 (ast.IntegerLiteral) / curToken = + / peekToken = 2
		// parseInfixExpression 을 호출하면,
		// 1. curPrecedence 는 SUM 이다.
		// 2. 다음 토큰으로 진행한다. 진행 후 curToken 은 2 이고 peekToken 은 + 이다.
		// 3. curPrecedence 인 SUM 을 넘겨 parseExpression 을 호출한다.
		// 4. 이때 다음 토큰인 + 연산자의 우선순위는 SUM 이고 현재 precedence 도 SUM 이므로 for문 조건에 해당되지 않는다.
		// 5. 최종적으로 leftExp 에는 1 + 2 를 나타내는 *ast.InfixExpression 노드가 저장된다.
		// 즉, infix 함수에 1 을 넣고 curToken 인 + 와 peekToken 인 2 를 조합하여 1 + 2 를 나타내는 노드가 된다.
		leftExp = infix(leftExp)
		// (2번째 반복) leftExp = 1 + 2 (ast.InfixExpression) / curToken = + / peekToken = 3
		// parseInfixExpression 을 호출하면,
		// 1. curPrecedence 는 SUM 이다.
		// 2. 다음 토큰으로 진행한다. 진행 후 curToken 은 3 이고 peekToken 은 세미콜론이다.
		// 3. curPrecedence 인 SUM 을 넘겨 parseExpression 을 호출한다.
		// 4. 이때 다음 토큰이 세미콜론이므로 for문 조건에 해당되지 않는다.
		// 5. 최종적으로 leftExp 에는 1 + 2(ast.InfixExpression) + 3(ast.IntegerLiteral) 을 나타내는 *ast.InfixExpression 노드가 저장된다.
		// 즉, infix 함수에 1 + 2(ast.InfixExpression) 를 넣고 curToken 인 + 와 peekToken 인 3 을 조합하여 1 + 2 + 3 을 나타내는 노드가 된다.
	}
	// peekToken 이 세미콜론이므로 가장 외부에서 진행되었던 for 문을 빠져나오게 된다.

	return leftExp
}

// peekPrecedence 다음 토큰 (peekToken) 의 우선순위를 반환한다.
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

// curPrecedence 현재 토큰 (curToken) 의 우선순위를 반환한다.
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

// parseIdentifier 식별자를 파싱한다.
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseIntegerLiteral 정수 리터럴을 파싱한다.
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

// parsePrefixExpression 전위 연산자를 파싱한다.
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	// 다음 토큰으로 진행한다.
	p.nextToken()

	// -5 라면 nextToken 을 호출한 이후의 curToken 은 5 이다.
	// parseExpression 은 5를 파싱하게 되고 IntegerLiteral ast 노드를 반환한다. -> 재귀적 접근
	// 이 값이 PrefixExpression 의 Right 필드에 들어간다.
	expression.Right = p.parseExpression(PREFIX) // TODO: 왜 PREFIX 인지 이해 필요

	return expression
}

// parseInfixExpression 중위 표현식을 파싱한다.
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	// leftExp = 1 (ast.IntegerLiteral) / curToken = + / peekToken = 2
	// 위와 같은 상황이라고 가정하면,
	// 현재 precedence 는 SUM 이다.
	precedence := p.curPrecedence()
	// 이 상태에서 다음 토큰으로 진행하면,
	// curToken = 2 / peekToken = + 이다.
	p.nextToken()
	// precedence 가 SUM 인 상태에서 다시 parseExpression 을 호출하고 그 값을 Right 필드에 저장한다. (현재까지 2번 호출됨)
	// 2번째 호출의 결과로 2를 표현하는 *ast.IntegerLiteral 노드가 반환된다.
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
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

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
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
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
