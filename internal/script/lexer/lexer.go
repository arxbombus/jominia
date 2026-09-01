package lexer

import (
	"strconv"

	"github.com/arxbombus/jominia/internal/script/syntax"
	"github.com/arxbombus/jominia/internal/text"
)

// Lexer tokenizes Jomini source text.
type Lexer struct {
	source     string
	position   text.TextSize
	sourceSize text.TextSize
}

// NewLexer returns a Lexer positioned at the start of source.
func NewLexer(source string) *Lexer {
	return &Lexer{
		source:     source,
		sourceSize: text.SizeOf(source),
	}
}

// Next scans and returns the next token.
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
	case '[':
		l.position++
		return l.token(syntax.LBracket, start)
	case ']':
		l.position++
		return l.token(syntax.RBracket, start)
	case '(':
		l.position++
		return l.token(syntax.LParen, start)
	case ')':
		l.position++
		return l.token(syntax.RParen, start)
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
		return l.token(syntax.Bang, start)
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
	case ';':
		l.position++
		return l.token(syntax.Semicolon, start)
	case ',':
		l.position++
		return l.token(syntax.Comma, start)
	case '"', '\'':
		kind := l.scanString(l.source[l.position])
		return l.token(kind, start)
	default:
		l.scanAtom()
		kind := syntax.Identifier
		value := l.source[start:l.position]
		if value == "yes" || value == "no" {
			kind = syntax.Boolean
		}
		if _, err := strconv.ParseFloat(value, 64); err == nil {
			kind = syntax.Number
		}
		return l.token(kind, start)
	}
}

// Lex tokenizes source, including trivia and the final EOF token.
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

// token returns a token spanning start through the current lexer position.
func (l *Lexer) token(kind syntax.SyntaxKind, start text.TextSize) Token {
	return Token{
		Kind:  kind,
		Range: text.NewTextRange(start, l.position),
	}
}

// eat consumes expected if it is the current byte.
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

// scanWhitespace consumes consecutive spaces and tabs.
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

// scanNewline consumes one newline, treating CRLF as a single newline.
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

// scanComment consumes a comment up to, but not including, the next newline.
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

// scanString consumes a quoted string and returns ErrorToken if it is unterminated.
func (l *Lexer) scanString(quoteType byte) syntax.SyntaxKind {
	kind := syntax.String
	if quoteType == '\'' {
		kind = syntax.SingleQuotedString
	}

	l.position++

	for l.position < l.sourceSize {
		switch l.source[l.position] {
		case '\\':
			l.position++
			if l.position < l.sourceSize {
				l.position++
			}
		default:
			if l.source[l.position] == quoteType {
				l.position++
				return kind
			}
			l.position++
		}
	}

	return syntax.ErrorToken
}

// scanAtom consumes bytes until the next atom boundary.
func (l *Lexer) scanAtom() {
	for l.position < l.sourceSize {
		if isAtomBoundary(l.source[l.position]) {
			return
		}

		l.position++
	}
}

// isAtomBoundary reports whether char terminates an unquoted atom.
func isAtomBoundary(char byte) bool {
	switch char {
	case ' ', '\t',
		'\n', '\r',
		'#',
		'{', '}',
		'[', ']',
		'(', ')',
		'=',
		'!',
		'<', '>',
		'?',
		';',
		',',
		'"', '\'':
		return true
	default:
		return false
	}
}
