package parser

import (
	"github.com/arxbombus/jominia/internal/script/lexer"
	"github.com/arxbombus/jominia/internal/script/syntax"
	"github.com/arxbombus/jominia/internal/text"
)

// Trivia records trivia skipped between parser tokens.
type Trivia struct {
	Kind       syntax.SyntaxKind
	Range      text.TextRange
	IsTrailing bool
}

type tokenLookahead struct {
	kind                  syntax.SyntaxKind
	hasPrecedingLineBreak bool
}

// TokenSource exposes non-trivia lexer tokens to the parser while retaining trivia.
type TokenSource struct {
	source string
	lexer  *lexer.Lexer

	current               lexer.Token
	trivia                []Trivia
	hasPrecedingLineBreak bool
}

// NewTokenSource returns a TokenSource positioned at the first non-trivia token.
func NewTokenSource(source string) *TokenSource {
	ts := &TokenSource{
		source: source,
		lexer:  lexer.NewLexer(source),
	}
	ts.nextNonTriviaToken(true)
	return ts
}

// Current returns the kind of the current non-trivia token.
func (ts *TokenSource) Current() syntax.SyntaxKind {
	return ts.current.Kind
}

// CurrentRange returns the source range of the current non-trivia token.
func (ts *TokenSource) CurrentRange() text.TextRange {
	return ts.current.Range
}

// Text returns the original source text.
func (ts *TokenSource) Text() string {
	return ts.source
}

// Nth returns the kind of the nth non-trivia token without consuming it. Nth(0) is equivalent to Current.
func (ts *TokenSource) Nth(n int) syntax.SyntaxKind {
	return ts.nth(n).kind
}

// NthHasPrecedingLineBreak reports whether the nth non-trivia lookahead token is preceded by a newline. NthHasPrecedingLineBreak(0) is equivalent to HasPrecedingLineBreak.
func (ts *TokenSource) NthHasPrecedingLineBreak(n int) bool {
	return ts.nth(n).hasPrecedingLineBreak
}

// HasPrecedingLineBreak reports whether the current token is preceded by a newline.
func (ts *TokenSource) HasPrecedingLineBreak() bool {
	return ts.hasPrecedingLineBreak
}

// Bump advances to the next non-trivia token.
func (ts *TokenSource) Bump() {
	if ts.current.Kind == syntax.EOF {
		return
	}
	ts.nextNonTriviaToken(false)
}

// Finish returns the trivia collected while advancing the token source.
//
// Lexer diagnostics may also be returned here in the future.
func (ts *TokenSource) Finish() []Trivia {
	return ts.trivia
}

func (ts *TokenSource) nth(n int) tokenLookahead {
	if n < 0 {
		panic("token source: lookahead index must be non-negative")
	}
	if n == 0 {
		return tokenLookahead{kind: ts.current.Kind, hasPrecedingLineBreak: ts.hasPrecedingLineBreak}
	}
	lookaheadLexer := *ts.lexer
	index := 1
	hasPrecedingLineBreak := false
	for {
		token := lookaheadLexer.Next()
		if token.Kind.IsTrivia() {
			if token.Kind == syntax.Newline {
				hasPrecedingLineBreak = true
			}
			continue
		}
		if index == n || token.Kind == syntax.EOF {
			return tokenLookahead{kind: token.Kind, hasPrecedingLineBreak: hasPrecedingLineBreak}
		}
		index++
		hasPrecedingLineBreak = false
	}
}

// nextNonTriviaToken advances until it finds a non-trivia token and records any trivia skipped along the way.
func (ts *TokenSource) nextNonTriviaToken(isFirstToken bool) {
	isTrailing := !isFirstToken
	ts.hasPrecedingLineBreak = false

	for {
		token := ts.lexer.Next()

		if !token.Kind.IsTrivia() {
			ts.current = token
			return
		}
		if token.Kind == syntax.Newline {
			isTrailing = false
			ts.hasPrecedingLineBreak = true
		}
		ts.trivia = append(ts.trivia, Trivia{
			Kind:       token.Kind,
			Range:      token.Range,
			IsTrailing: isTrailing,
		})
	}
}
