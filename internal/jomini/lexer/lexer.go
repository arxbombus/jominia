package lexer

import (
	"strconv"

	"github.com/arxbombus/jominia/internal/jomini/syntax"
	"github.com/arxbombus/jominia/internal/text"
)

type Lexer struct {
	source     string
	position   text.TextSize
	sourceSize text.TextSize
}

func NewLexer(source string) *Lexer {
	return &Lexer{
		source:     source,
		sourceSize: text.SizeOf(source),
	}
}

func (l *Lexer) Next() Token {
	start := l.position

	if l.position >= l.sourceSize {
		return l.token(syntax.EOF, start)
	}

	switch l.source[l.position] {
	case ' ', '\t':
		l.scanWhitespace()
		return l.token(syntax.Whitespace, start)
	case '\n', '\r':
		l.scanNewline()
		return l.token(syntax.Newline, start)
	case '#':
		l.scanComment()
		return l.token(syntax.Comment, start)
	case '{':
		l.position++
		return l.token(syntax.LCurly, start)
	case '}':
		l.position++
		return l.token(syntax.RCurly, start)
	case '=':
		l.position++
		if l.eat('=') {
			return l.token(syntax.EqualsEquals, start)
		}
		return l.token(syntax.Equals, start)
	case '!':
		l.position++
		if l.eat('=') {
			return l.token(syntax.BangEquals, start)
		}
		return l.token(syntax.ErrorToken, start)
	case '<':
		l.position++
		if l.eat('=') {
			return l.token(syntax.LessEquals, start)
		}
		return l.token(syntax.Less, start)
	case '>':
		l.position++
		if l.eat('=') {
			return l.token(syntax.GreaterEquals, start)
		}
		return l.token(syntax.Greater, start)
	case '?':
		l.position++
		if l.eat('=') {
			return l.token(syntax.QuestionEquals, start)
		}
		return l.token(syntax.Question, start)
	case '"':
		kind := l.scanString()
		return l.token(kind, start)
	default:
		l.scanAtom()
		kind := syntax.Identifier

		if _, err := strconv.ParseFloat(l.source[start:l.position], 64); err == nil {
			kind = syntax.Number
		}
		return l.token(kind, start)
	}
}

func Lex(source string) []Token {
	l := NewLexer(source)

	var tokens []Token
	for {
		token := l.Next()
		tokens = append(tokens, token)

		if token.Kind == syntax.EOF {
			return tokens
		}
	}
}

func (l *Lexer) token(kind syntax.SyntaxKind, start text.TextSize) Token {
	return Token{
		Kind:  kind,
		Range: text.NewTextRange(start, l.position),
	}
}

func (l *Lexer) eat(expected byte) bool {
	if l.position >= l.sourceSize {
		return false
	}

	if l.source[l.position] != expected {
		return false
	}

	l.position++
	return true
}

func (l *Lexer) scanWhitespace() {
	for l.position < l.sourceSize {
		switch l.source[l.position] {
		case ' ', '\t':
			l.position++
		default:
			return
		}
	}
}

func (l *Lexer) scanNewline() {
	if l.source[l.position] == '\r' {
		l.position++

		if l.position < l.sourceSize && l.source[l.position] == '\n' {
			l.position++
		}

		return
	}

	l.position++
}

func (l *Lexer) scanComment() {
	for l.position < l.sourceSize {
		switch l.source[l.position] {
		case '\n', '\r':
			return
		default:
			l.position++
		}
	}
}

func (l *Lexer) scanString() syntax.SyntaxKind {
	l.position++

	for l.position < l.sourceSize {
		switch l.source[l.position] {
		case '"':
			l.position++
			return syntax.String
		// We don't handle escapes.
		case '\\':
			l.position++

			if l.position < l.sourceSize {
				l.position++
			}
		default:
			l.position++
		}
	}

	return syntax.ErrorToken
}

func (l *Lexer) scanAtom() {
	for l.position < l.sourceSize {
		if isAtomBoundary(l.source[l.position]) {
			return
		}

		l.position++
	}
}

func isAtomBoundary(char byte) bool {
	switch char {
	case ' ', '\t',
		'\n', '\r',
		'#',
		'{', '}',
		'=',
		'!',
		'<', '>',
		'?',
		'"':
		return true
	default:
		return false
	}
}
