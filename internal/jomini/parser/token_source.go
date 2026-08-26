package parser

import (
	"github.com/arxbombus/jominia/internal/jomini/lexer"
	"github.com/arxbombus/jominia/internal/jomini/syntax"
	"github.com/arxbombus/jominia/internal/text"
)

/*
 * In the future we can add stuff like `Nth()`, `Peek()`, `Checkpoint()`, `Rewind()`, `SkipAsTrivia()` etc.
 */

type Trivia struct {
	Kind       syntax.SyntaxKind
	Range      text.TextRange
	IsTrailing bool
}

type TokenSource struct {
	source string
	lexer  *lexer.Lexer

	current               lexer.Token
	trivia                []Trivia
	hasPrecedingLineBreak bool
}

func NewTokenSource(source string) *TokenSource {
	ts := &TokenSource{
		source: source,
		lexer:  lexer.NewLexer(source),
	}
	ts.nextNonTriviaToken(true)
	return ts
}

func (ts *TokenSource) Current() syntax.SyntaxKind {
	return ts.current.Kind
}

func (ts *TokenSource) CurrentRange() text.TextRange {
	return ts.current.Range
}

func (ts *TokenSource) Text() string {
	return ts.source
}

func (ts *TokenSource) HasPrecedingLineBreak() bool {
	return ts.hasPrecedingLineBreak
}

func (ts *TokenSource) Bump() {
	if ts.current.Kind == syntax.EOF {
		return
	}
	ts.nextNonTriviaToken(false)
}

// Should return trivia AND lexer diagnostics in the future
func (ts *TokenSource) Finish() []Trivia {
	return ts.trivia
}

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
