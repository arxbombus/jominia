package lexer

import (
	"strconv"

	"github.com/arxbombus/jominia/internal/script/syntax"
	"github.com/arxbombus/jominia/internal/text"
)

// Lexer tokenizes source text.
type Lexer struct {
	source     string
	position   text.TextSize
	sourceSize text.TextSize
	mode       lexMode
}

// ReLexContext selects a contextual grammar for the current token.
//
// The normal lexer deliberately keeps punctuation such as hyphens and slashes inside atoms. Inline math assigns those same bytes expression meanings, so the parser requests a contextual re-lex after seeing `@[` (or its escaped spelling, `@\[`).
type ReLexContext uint8

const (
	ReLexNormal ReLexContext = iota
	ReLexInlineMath
)

type lexMode uint8

const (
	lexNormal lexMode = iota
	lexInlineMath
	lexInlineMathParameterName
	lexInlineMathParameterArgument
	lexInlineMathParameterArgumentEnd
)

// NewLexer returns a Lexer positioned at the start of source.
func NewLexer(source string) *Lexer {
	return &Lexer{
		source:     source,
		sourceSize: text.SizeOf(source),
	}
}

// Next scans and returns the next token.
func (l *Lexer) Next() Token {
	switch l.mode {
	case lexNormal:
		return l.nextNormal()
	case lexInlineMath:
		return l.nextInlineMath()
	case lexInlineMathParameterName:
		return l.nextInlineMathParameterName()
	case lexInlineMathParameterArgument:
		return l.nextInlineMathParameterArgument()
	case lexInlineMathParameterArgumentEnd:
		return l.nextInlineMathParameterArgumentEnd()
	default:
		panic("lexer: unknown lex mode")
	}
}

// ReLex rewinds current and scans it using context.
func (l *Lexer) ReLex(current Token, context ReLexContext) Token {
	if current.Range.End() != l.position {
		panic("lexer: can only re-lex the current token")
	}

	l.position = current.Range.Start()
	switch context {
	case ReLexNormal:
		l.mode = lexNormal
	case ReLexInlineMath:
		l.mode = lexInlineMath
	default:
		panic("lexer: unknown re-lex context")
	}

	return l.Next()
}

func (l *Lexer) nextNormal() Token {
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
	case '@':
		if l.scanInlineMathStart() {
			return l.token(syntax.InlineMathStart, start)
		}
		l.scanAtom()
		return l.classifyAtom(start)
	case '"', '\'':
		kind := l.scanString(l.source[l.position])
		return l.token(kind, start)
	default:
		l.scanAtom()
		return l.classifyAtom(start)
	}
}

func (l *Lexer) nextInlineMath() Token {
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
	case ']':
		l.position++
		l.mode = lexNormal
		return l.token(syntax.RBracket, start)
	case '(':
		l.position++
		return l.token(syntax.LParen, start)
	case ')':
		l.position++
		return l.token(syntax.RParen, start)
	case '+':
		l.position++
		return l.token(syntax.Plus, start)
	case '-':
		l.position++
		return l.token(syntax.Minus, start)
	case '*':
		l.position++
		return l.token(syntax.Star, start)
	case '/':
		l.position++
		return l.token(syntax.Slash, start)
	case '%':
		l.position++
		return l.token(syntax.Percent, start)
	case '|':
		l.position++
		return l.token(syntax.Pipe, start)
	case '$':
		l.position++
		l.mode = lexInlineMathParameterName
		return l.token(syntax.Dollar, start)
	case '@':
		l.position++
		return l.token(syntax.At, start)
	case '{':
		l.position++
		return l.token(syntax.LCurly, start)
	case '}':
		l.position++
		return l.token(syntax.RCurly, start)
	case '[':
		l.position++
		return l.token(syntax.LBracket, start)
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
		l.scanInlineMathAtom()
		return l.classifyInlineMathAtom(start)
	}
}

func (l *Lexer) nextInlineMathParameterName() Token {
	start := l.position
	if l.position >= l.sourceSize {
		return l.token(syntax.EOF, start)
	}

	switch l.source[l.position] {
	case '$':
		l.position++
		l.mode = lexInlineMath
		return l.token(syntax.Dollar, start)
	case '|':
		l.position++
		l.mode = lexInlineMathParameterArgument
		return l.token(syntax.Pipe, start)
	case ']':
		l.position++
		l.mode = lexNormal
		return l.token(syntax.RBracket, start)
	case ' ', '\t':
		l.scanWhitespace()
		return l.token(syntax.Whitespace, start)
	case '\n', '\r':
		l.scanNewline()
		return l.token(syntax.Newline, start)
	default:
		if isParameterNameByte(l.source[l.position]) {
			l.scanParameterName()
			return l.token(syntax.ParameterName, start)
		}
		// Return to expression lexing for malformed parameters so recovery
		// still sees operators and delimiters with their normal math meaning.
		l.mode = lexInlineMath
		return l.nextInlineMath()
	}
}

func (l *Lexer) nextInlineMathParameterArgument() Token {
	start := l.position
	if l.position >= l.sourceSize {
		return l.token(syntax.EOF, start)
	}

	switch l.source[l.position] {
	case '$':
		l.position++
		l.mode = lexInlineMath
		return l.token(syntax.Dollar, start)
	case ']':
		l.position++
		l.mode = lexNormal
		return l.token(syntax.RBracket, start)
	case ' ', '\t':
		l.scanWhitespace()
		return l.token(syntax.Whitespace, start)
	case '\n', '\r':
		l.scanNewline()
		return l.token(syntax.Newline, start)
	default:
		l.scanParameterArgument()
		l.mode = lexInlineMathParameterArgumentEnd
		return l.token(syntax.ParameterArgument, start)
	}
}

func (l *Lexer) nextInlineMathParameterArgumentEnd() Token {
	start := l.position
	if l.position >= l.sourceSize {
		return l.token(syntax.EOF, start)
	}

	switch l.source[l.position] {
	case '$':
		l.position++
		l.mode = lexInlineMath
		return l.token(syntax.Dollar, start)
	case ']':
		l.position++
		l.mode = lexNormal
		return l.token(syntax.RBracket, start)
	case ' ', '\t':
		l.scanWhitespace()
		return l.token(syntax.Whitespace, start)
	case '\n', '\r':
		l.scanNewline()
		return l.token(syntax.Newline, start)
	default:
		l.position++
		return l.token(syntax.ErrorToken, start)
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

func (l *Lexer) classifyAtom(start text.TextSize) Token {
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

func (l *Lexer) classifyInlineMathAtom(start text.TextSize) Token {
	kind := syntax.Identifier
	if _, err := strconv.ParseFloat(l.source[start:l.position], 64); err == nil {
		kind = syntax.Number
	}
	return l.token(kind, start)
}

// scanInlineMathStart consumes either @[ or the escaped opener @\[.
func (l *Lexer) scanInlineMathStart() bool {
	position := int(l.position)
	if position+1 < len(l.source) && l.source[position+1] == '[' {
		l.position += 2
		return true
	}
	if position+2 < len(l.source) &&
		l.source[position+1] == '\\' &&
		l.source[position+2] == '[' {
		l.position += 3
		return true
	}
	return false
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

func (l *Lexer) scanInlineMathAtom() {
	for l.position < l.sourceSize {
		switch l.source[l.position] {
		case ' ', '\t', '\n', '\r', '#',
			'{', '}', '[', ']', '(', ')',
			'=', '!', '<', '>', '?', ';', ',',
			'"', '\'', '@', '$', '|', '+', '-', '*', '/', '%':
			return
		default:
			l.position++
		}
	}
}

func (l *Lexer) scanParameterName() {
	for l.position < l.sourceSize && isParameterNameByte(l.source[l.position]) {
		l.position++
	}
}

func (l *Lexer) scanParameterArgument() {
	for l.position < l.sourceSize {
		switch l.source[l.position] {
		case '$', ']', ' ', '\t', '\n', '\r':
			return
		default:
			l.position++
		}
	}
}

func isParameterNameByte(char byte) bool {
	return char >= 'a' && char <= 'z' ||
		char >= 'A' && char <= 'Z' ||
		char >= '0' && char <= '9' ||
		char == '_'
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
