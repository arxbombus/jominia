package lexer

import (
	"strconv"

	"github.com/arxbombus/jominia/internal/script/syntax"
	"github.com/arxbombus/jominia/internal/text"
)

// Lexer tokenizes source text.
type Lexer struct {
	source              string
	position            text.TextSize
	sourceSize          text.TextSize
	mode                lexMode
	parameterReturnMode lexMode
	reLexEnd            text.TextSize
}

// ReLexContext selects a contextual grammar for the current token.
//
// The normal lexer deliberately keeps punctuation such as hyphens, slashes, and dollar signs inside atoms. Nested grammars assign those same bytes more specific meanings, so the parser requests contextual re-lexing only after it has recognized the surrounding syntax.
type ReLexContext uint8

const (
	ReLexNormal ReLexContext = iota
	ReLexInlineMath
	ReLexParameter
	ReLexInterpolatedIdentifier
	ReLexInterpolatedString
	ReLexVariableReference
)

// LexContext selects the grammar used to scan the token after the current token. Unlike ReLexContext, it never changes a token that has already been scanned.
type LexContext uint8

const (
	LexNormal LexContext = iota
	LexInlineMath
	LexBracketExpression
	LexInterpolatedString
)

type lexMode uint8

const (
	lexNormal lexMode = iota
	lexInlineMath
	lexBracketExpression
	lexParameterName
	lexParameterArgument
	lexParameterArgumentEnd
	lexInterpolatedIdentifier
	lexInterpolatedString
	lexVariableReference
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
	case lexBracketExpression:
		return l.nextBracketExpression()
	case lexParameterName:
		return l.nextParameterName()
	case lexParameterArgument:
		return l.nextParameterArgument()
	case lexParameterArgumentEnd:
		return l.nextParameterArgumentEnd()
	case lexInterpolatedIdentifier:
		return l.nextInterpolatedIdentifier()
	case lexInterpolatedString:
		return l.nextInterpolatedString()
	case lexVariableReference:
		return l.nextVariableReference()
	default:
		panic("lexer: unknown lex mode")
	}
}

// NextWithContext scans the next token using context. Parameter and interpolation sub-grammars remain active until their own bounded region ends, then return to the requested enclosing grammar.
func (l *Lexer) NextWithContext(context LexContext) Token {
	switch l.mode {
	case lexNormal, lexInlineMath, lexBracketExpression, lexInterpolatedString:
		l.mode = modeForContext(context)
	case lexParameterName,
		lexParameterArgument,
		lexParameterArgumentEnd,
		lexInterpolatedIdentifier,
		lexVariableReference:
		// Bounded sub-grammars decide when to return to their enclosing mode.
	}
	return l.Next()
}

func modeForContext(context LexContext) lexMode {
	switch context {
	case LexNormal:
		return lexNormal
	case LexInlineMath:
		return lexInlineMath
	case LexBracketExpression:
		return lexBracketExpression
	case LexInterpolatedString:
		return lexInterpolatedString
	default:
		panic("lexer: unknown lex context")
	}
}

// ReLex rewinds current and scans it using context.
func (l *Lexer) ReLex(current Token, context ReLexContext) Token {
	if current.Range.End() != l.position {
		panic("lexer: can only re-lex the current token")
	}
	start := current.Range.Start()
	l.position = start
	switch context {
	case ReLexNormal:
		l.mode = lexNormal
	case ReLexInlineMath:
		l.mode = lexInlineMath
	case ReLexParameter:
		if l.position >= l.sourceSize || l.source[l.position] != '$' {
			panic("lexer: parameter re-lex must start at a dollar sign")
		}
		return l.startParameter(start, lexNormal)
	case ReLexInterpolatedIdentifier:
		if current.Kind != syntax.Identifier {
			panic("lexer: interpolated identifier re-lex requires an identifier")
		}
		l.reLexEnd = current.Range.End()
		l.mode = lexInterpolatedIdentifier
	case ReLexInterpolatedString:
		if current.Kind != syntax.String || l.position >= l.sourceSize || l.source[l.position] != '"' {
			panic("lexer: interpolated string re-lex requires a double-quoted string")
		}
		l.reLexEnd = current.Range.End()
		l.mode = lexInterpolatedString
	case ReLexVariableReference:
		if current.Kind != syntax.Identifier || l.position >= l.sourceSize || l.source[l.position] != '@' {
			panic("lexer: variable re-lex requires an at-prefixed identifier")
		}
		l.reLexEnd = current.Range.End()
		l.position++
		l.mode = lexVariableReference
		return l.token(syntax.At, start)
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
		return l.startParameter(start, lexInlineMath)
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

func (l *Lexer) nextBracketExpression() Token {
	start := l.position
	if l.position >= l.sourceSize {
		return l.token(syntax.EOF, start)
	}
	// A bracket expression embedded in a re-lexed double-quoted string must never consume the host string's closing quote as an argument string.
	if l.position+1 == l.reLexEnd && l.source[l.position] == '"' {
		l.position++
		l.mode = lexNormal
		return l.token(syntax.StringQuote, start)
	}
	switch l.source[l.position] {
	case ' ', '\t':
		l.scanWhitespace()
		return l.token(syntax.Whitespace, start)
	case '\n', '\r':
		l.scanNewline()
		return l.token(syntax.Newline, start)
	case '#':
		if l.reLexEnd > l.position {
			l.scanBracketAtom()
			return l.classifyBracketAtom(start)
		}
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
		l.mode = lexNormal
		return l.token(syntax.RBracket, start)
	case '(':
		l.position++
		return l.token(syntax.LParen, start)
	case ')':
		l.position++
		return l.token(syntax.RParen, start)
	case ',':
		l.position++
		return l.token(syntax.Comma, start)
	case '.':
		l.position++
		return l.token(syntax.Dot, start)
	case '?':
		l.position++
		if l.eat('=') {
			return l.token(syntax.QuestionEquals, start)
		}
		return l.token(syntax.Question, start)
	case '!':
		l.position++
		if l.eat('=') {
			return l.token(syntax.BangEquals, start)
		}
		return l.token(syntax.Bang, start)
	case '=':
		l.position++
		if l.eat('=') {
			return l.token(syntax.EqualsEquals, start)
		}
		return l.token(syntax.Equals, start)
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
	case ';':
		l.position++
		return l.token(syntax.Semicolon, start)
	case '|':
		l.position++
		return l.token(syntax.Pipe, start)
	case '$':
		return l.startParameter(start, lexBracketExpression)
	case '@':
		l.position++
		return l.token(syntax.At, start)
	case '"', '\'':
		kind := l.scanBracketString(l.source[l.position])
		return l.token(kind, start)
	default:
		if l.scanBracketNumber() {
			return l.token(syntax.Number, start)
		}
		l.scanBracketAtom()
		return l.classifyBracketAtom(start)
	}
}

func (l *Lexer) nextParameterName() Token {
	start := l.position
	if l.atParameterContextEnd() {
		return l.nextAfterParameter()
	}
	if l.position >= l.sourceSize {
		return l.token(syntax.EOF, start)
	}
	switch l.source[l.position] {
	case '$':
		return l.finishParameter(start)
	case '|':
		l.position++
		l.mode = lexParameterArgument
		return l.token(syntax.Pipe, start)
	default:
		if isParameterNameByte(l.source[l.position]) {
			l.scanParameterName()
			return l.token(syntax.ParameterName, start)
		}
		// Return to the enclosing grammar for malformed parameters so recovery still sees operators, delimiters, whitespace, and line breaks with their normal meaning.
		return l.nextAfterParameter()
	}
}

func (l *Lexer) nextParameterArgument() Token {
	start := l.position
	if l.atParameterContextEnd() {
		return l.nextAfterParameter()
	}
	if l.position >= l.sourceSize {
		return l.token(syntax.EOF, start)
	}
	switch l.source[l.position] {
	case '$':
		return l.finishParameter(start)
	default:
		if l.isParameterArgumentBoundary(l.source[l.position]) {
			return l.nextAfterParameter()
		}
		l.scanParameterArgument()
		l.mode = lexParameterArgumentEnd
		return l.token(syntax.ParameterArgument, start)
	}
}

func (l *Lexer) nextParameterArgumentEnd() Token {
	start := l.position
	if l.atParameterContextEnd() {
		return l.nextAfterParameter()
	}
	if l.position >= l.sourceSize {
		return l.token(syntax.EOF, start)
	}
	switch l.source[l.position] {
	case '$':
		return l.finishParameter(start)
	default:
		return l.nextAfterParameter()
	}
}

func (l *Lexer) nextInterpolatedIdentifier() Token {
	start := l.position
	if l.position >= l.reLexEnd {
		l.mode = lexNormal
		return l.Next()
	}
	if l.source[l.position] == '$' && l.isPlausibleParameterStart(l.reLexEnd) {
		return l.startParameter(start, lexInterpolatedIdentifier)
	}
	for l.position < l.reLexEnd {
		if l.source[l.position] == '$' && l.isPlausibleParameterStart(l.reLexEnd) {
			break
		}
		l.position++
	}
	return l.token(syntax.IdentifierFragment, start)
}

func (l *Lexer) nextInterpolatedString() Token {
	start := l.position
	if l.position >= l.reLexEnd {
		l.mode = lexNormal
		return l.Next()
	}
	if l.source[l.position] == '"' {
		l.position++
		if l.position == l.reLexEnd {
			l.mode = lexNormal
		}
		return l.token(syntax.StringQuote, start)
	}
	if l.source[l.position] == '$' && l.isPlausibleParameterStart(l.reLexEnd-1) {
		return l.startParameter(start, lexInterpolatedString)
	}
	if l.source[l.position] == '[' {
		l.position++
		return l.token(syntax.LBracket, start)
	}
	contentEnd := l.reLexEnd - 1
	for l.position < contentEnd {
		if l.source[l.position] == '\\' && l.position+1 < contentEnd {
			l.position += 2
			continue
		}
		if l.source[l.position] == '$' && l.isPlausibleParameterStart(contentEnd) {
			break
		}
		if l.source[l.position] == '[' {
			break
		}
		l.position++
	}
	return l.token(syntax.StringFragment, start)
}

func (l *Lexer) nextVariableReference() Token {
	start := l.position
	if l.position >= l.reLexEnd {
		l.mode = lexNormal
		return l.Next()
	}
	l.position = l.reLexEnd
	l.mode = lexNormal
	return l.token(syntax.Identifier, start)
}

func (l *Lexer) startParameter(start text.TextSize, returnMode lexMode) Token {
	if l.source[l.position] != '$' {
		panic("lexer: expected parameter opener")
	}
	l.position++
	l.parameterReturnMode = returnMode
	l.mode = lexParameterName
	return l.token(syntax.Dollar, start)
}

func (l *Lexer) finishParameter(start text.TextSize) Token {
	l.position++
	l.mode = l.parameterReturnMode
	return l.token(syntax.Dollar, start)
}

func (l *Lexer) nextAfterParameter() Token {
	l.mode = l.parameterReturnMode
	return l.Next()
}

func (l *Lexer) atParameterContextEnd() bool {
	switch l.parameterReturnMode {
	case lexInterpolatedIdentifier:
		return l.position >= l.reLexEnd
	case lexInterpolatedString:
		return l.position >= l.reLexEnd-1
	case lexNormal,
		lexInlineMath,
		lexBracketExpression,
		lexParameterName,
		lexParameterArgument,
		lexParameterArgumentEnd,
		lexVariableReference:
		return false
	}
	panic("lexer: unknown parameter return mode")
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

func (l *Lexer) classifyBracketAtom(start text.TextSize) Token {
	value := l.source[start:l.position]
	if value == "yes" || value == "no" {
		return l.token(syntax.Boolean, start)
	}
	return l.token(syntax.Identifier, start)
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
	return l.scanStringUntil(quoteType, l.sourceSize)
}

func (l *Lexer) scanBracketString(quoteType byte) syntax.SyntaxKind {
	end := l.sourceSize
	if l.reLexEnd > l.position {
		end = l.reLexEnd - 1
	}
	return l.scanStringUntil(quoteType, end)
}

func (l *Lexer) scanStringUntil(quoteType byte, end text.TextSize) syntax.SyntaxKind {
	kind := syntax.String
	if quoteType == '\'' {
		kind = syntax.SingleQuotedString
	}
	l.position++
	for l.position < end {
		switch l.source[l.position] {
		case '\\':
			l.position++
			if l.position < end {
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

func (l *Lexer) scanBracketAtom() {
	for l.position < l.sourceSize {
		if l.position+1 == l.reLexEnd && l.source[l.position] == '"' {
			return
		}
		switch l.source[l.position] {
		case ' ', '\t', '\n', '\r',
			'{', '}', '[', ']', '(', ')',
			'=', '<', '>', '?', '!', ';', ',', '.', '"', '\'', '@', '$', '|':
			return
		default:
			l.position++
		}
	}
}

// scanBracketNumber consumes a decimal number only when the entire candidate ends at a bracket-expression boundary. This keeps dotted member names split without breaking decimal arguments.
func (l *Lexer) scanBracketNumber() bool {
	start := l.position
	if l.position < l.sourceSize && (l.source[l.position] == '+' || l.source[l.position] == '-') {
		l.position++
	}
	digitStart := l.position
	for l.position < l.sourceSize && l.source[l.position] >= '0' && l.source[l.position] <= '9' {
		l.position++
	}
	if l.position == digitStart {
		l.position = start
		return false
	}
	if l.position < l.sourceSize && l.source[l.position] == '.' {
		dot := l.position
		l.position++
		fractionStart := l.position
		for l.position < l.sourceSize && l.source[l.position] >= '0' && l.source[l.position] <= '9' {
			l.position++
		}
		if l.position == fractionStart {
			l.position = dot
		}
	}
	if l.position < l.sourceSize && !isBracketAtomBoundary(l.source[l.position]) {
		l.position = start
		return false
	}
	return true
}

func isBracketAtomBoundary(char byte) bool {
	switch char {
	case ' ', '\t', '\n', '\r',
		'{', '}', '[', ']', '(', ')',
		'=', '<', '>', '?', '!', ';', ',', '.', '"', '\'', '@', '$', '|':
		return true
	default:
		return false
	}
}

func (l *Lexer) scanParameterName() {
	for l.position < l.sourceSize && isParameterNameByte(l.source[l.position]) {
		l.position++
	}
}

func (l *Lexer) scanParameterArgument() {
	for l.position < l.sourceSize {
		if l.parameterReturnMode == lexInterpolatedIdentifier && l.position >= l.reLexEnd {
			return
		}
		if l.parameterReturnMode == lexInterpolatedString {
			if l.position >= l.reLexEnd-1 {
				return
			}
			if l.source[l.position] == '\\' && l.position+1 < l.reLexEnd-1 {
				l.position += 2
				continue
			}
		}
		if l.isParameterArgumentBoundary(l.source[l.position]) {
			return
		}
		l.position++
	}
}

func (l *Lexer) isParameterArgumentBoundary(char byte) bool {
	if char == '$' || char == ' ' || char == '\t' || char == '\n' || char == '\r' {
		return true
	}
	if l.parameterReturnMode == lexNormal {
		return isAtomBoundary(char)
	}
	if l.parameterReturnMode == lexInterpolatedIdentifier {
		return false
	}
	if l.parameterReturnMode == lexInterpolatedString {
		return char == '"'
	}
	switch char {
	case '#', '{', '}', '[', ']', '(', ')', ';', ',', '"', '\'':
		return true
	default:
		return false
	}
}

func (l *Lexer) isPlausibleParameterStart(end text.TextSize) bool {
	next := l.position + 1
	if next >= end {
		return false
	}
	char := l.source[next]
	return isParameterNameByte(char) || char == '|' || char == '$'
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
