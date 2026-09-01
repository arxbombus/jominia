package parser

import (
	"github.com/arxbombus/jominia/internal/jomini/syntax"
	"github.com/arxbombus/jominia/internal/text"
	"github.com/arxbombus/jominia/internal/tree"
)

// TreeSink receives parser events and builds a syntax tree.
type TreeSink interface {
	Token(kind syntax.SyntaxKind, end text.TextSize)
	StartNode(kind syntax.SyntaxKind)
	FinishNode()
}

// LosslessTreeSink builds a green tree while preserving the complete source text.
type LosslessTreeSink struct {
	source string
	trivia []Trivia

	textPosition   text.TextSize
	triviaPosition int
	openNodes      int
	needsEOF       bool
	triviaPieces   []tree.TriviaPiece

	builder *tree.TreeBuilder
}

// NewLosslessTreeSink returns a lossless tree sink with its own node cache.
func NewLosslessTreeSink(source string, trivia []Trivia) *LosslessTreeSink {
	return newLosslessTreeSink(
		source,
		trivia,
		tree.NewTreeBuilder(),
	)
}

// NewLosslessTreeSinkWithCache returns a lossless tree sink that reuses cache.
func NewLosslessTreeSinkWithCache(source string, trivia []Trivia, cache *tree.NodeCache) *LosslessTreeSink {
	return newLosslessTreeSink(
		source,
		trivia,
		tree.NewTreeBuilderWithCache(cache),
	)
}

func newLosslessTreeSink(source string, trivia []Trivia, builder *tree.TreeBuilder) *LosslessTreeSink {
	return &LosslessTreeSink{
		source:       source,
		trivia:       trivia,
		needsEOF:     true,
		triviaPieces: make([]tree.TriviaPiece, 0, 128),
		builder:      builder,
	}
}

// Token adds a parser token and its surrounding trivia to the tree.
func (s *LosslessTreeSink) Token(kind syntax.SyntaxKind, end text.TextSize) {
	s.doToken(kind, end)
}

// StartNode begins a new syntax node of the given kind.
func (s *LosslessTreeSink) StartNode(kind syntax.SyntaxKind) {
	s.builder.StartNode(kind.Raw())
	s.openNodes++
}

// FinishNode finishes the current syntax node.
func (s *LosslessTreeSink) FinishNode() {
	if s.openNodes == 0 {
		panic("parser: tree sink has no open node to finish")
	}
	s.openNodes--
	if s.openNodes == 0 && s.needsEOF {
		s.doToken(syntax.EOF, text.SizeOf(s.source))
	}
	s.builder.FinishNode()
}

// Finish completes tree construction and returns the root green node.
func (s *LosslessTreeSink) Finish() *tree.GreenNode {
	return s.builder.Finish()
}

func (s *LosslessTreeSink) doToken(kind syntax.SyntaxKind, tokenEnd text.TextSize) {
	if kind == syntax.EOF {
		s.needsEOF = false
	}
	tokenStart := s.textPosition
	s.eatTrivia(false, tokenEnd)
	trailingStart := len(s.triviaPieces)
	s.textPosition = tokenEnd
	s.eatTrivia(true, tokenEnd)
	value := s.source[int(tokenStart):int(s.textPosition)]
	leading := s.triviaPieces[:trailingStart]
	trailing := s.triviaPieces[trailingStart:]
	s.builder.TokenWithTrivia(kind.Raw(), value, leading, trailing)
	s.triviaPieces = s.triviaPieces[:0]
}

func (s *LosslessTreeSink) eatTrivia(trailing bool, tokenEnd text.TextSize) {
	for s.triviaPosition < len(s.trivia) {
		trivia := s.trivia[s.triviaPosition]
		if trivia.IsTrailing != trailing ||
			trivia.Range.Start() != s.textPosition ||
			(!trailing && trivia.Range.End() > tokenEnd) {
			return
		}
		s.triviaPieces = append(s.triviaPieces, triviaPiece(trivia))
		s.textPosition = trivia.Range.End()
		s.triviaPosition++
	}
}

func triviaPiece(trivia Trivia) tree.TriviaPiece {
	if trivia.Kind == syntax.Whitespace {
		return tree.NewWhitespaceTriviaPiece(trivia.Range.Len())
	}
	if trivia.Kind == syntax.Newline {
		return tree.NewNewlineTriviaPiece(trivia.Range.Len())
	}
	if trivia.Kind == syntax.Comment {
		return tree.NewCommentTriviaPiece(trivia.Range.Len())
	}
	panic("parser: unsupported trivia kind")
}
