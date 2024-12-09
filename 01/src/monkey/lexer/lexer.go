package lexer

import "monkey/token"

// Lexer 는 입력을 토큰으로 변환하는 구조체이다.
// 이 렉서는 유니코드를 지원하지 않으며 ASCII 문자만 지원한다.
type Lexer struct {
	input        string
	position     int  // 입력에서 현재 위치 (현재 문자를 가리킴)
	readPosition int  // 입력에서 현재 읽는 위치 (현재 문자의 다음을 가리킴) -> 현재 문자를 보존하면서 다음 문자를 볼 수 있어야 하므로 두 개의 포인터가 필요함
	ch           byte // 현재 조사하고 있는 문자 (ASCII 만 지원하므로 rune 이 아닌 byte)
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar() // 첫 번째 문자를 읽어 ch 에 설정한 뒤 Position 과 readPosition 을 설정 (초기화)
	return l
}

// readChar 는 현재 조사하고 있는 문자를 읽고 다음 문자로 넘어감
func (l *Lexer) readChar() {
	// 문자열 Input 의 끝에 도달했는지 확인하고 그에 따라 l.ch 를 설정
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			// == 연산자인 경우
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			// = 대입 연산자인 경우
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '!':
		if l.peekChar() == '=' {
			// != 연산자인 경우
			ch := l.ch   // 현재 문자
			l.readChar() // 다음 문자로 이동
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			// ! 부정 연산자인 경우
			tok = newToken(token.BANG, l.ch)
		}
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '<':
		tok = newToken(token.LT, l.ch)
	case '>':
		tok = newToken(token.GT, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			// 문자열이라면,
			tok.Literal = l.readIdentifier()          // 식별자 읽기
			tok.Type = token.LookupIdent(tok.Literal) // 식별자의 literal 을 통해 식별자인지 혹은 키워드인지 판단하고 예약어라면 IDENT 타입으로 처리
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	// 현재 position 과 ReadPosition 증가
	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// peekChar 는 position 을 변경하지 않고 다음 문자열만 읽는다.
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

// readIdentifier 는 식별자를 읽고 식별자의 끝까지 읽는다.
func (l *Lexer) readIdentifier() string {
	position := l.position // 시작점
	for isLetter(l.ch) {
		// 문자가 유효한 식별자 문자인 경우 계속 읽음
		l.readChar()
	}
	return l.input[position:l.position] // 시작점부터 현재 위치까지의 문자열을 반환 = 식별자
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// isLetter 는 주어진 바이트가 유효한 식별자 문자인지 여부를 반환한다.
// _ 도 식별자 문자로 간주한다.
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// isDigit 는 주어진 바이트가 '정수'인지 여부를 반환한다.
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// newToken 은 새로운 Token 을 생성하고 반환한다.
func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}
