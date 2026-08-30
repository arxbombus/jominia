package tree

import "github.com/arxbombus/jominia/internal/text"

// TriviaPieceKind identifies the kind of trivia represented by a TriviaPiece.
type TriviaPieceKind uint8

const (
	TriviaNewline TriviaPieceKind = iota
	TriviaWhitespace
	TriviaComment
)

// IsNewline reports whether the trivia is a line break.
func (k TriviaPieceKind) IsNewline() bool {
	return k == TriviaNewline
}

// IsWhitespace reports whether the trivia is non-line-breaking whitespace.
func (k TriviaPieceKind) IsWhitespace() bool {
	return k == TriviaWhitespace
}

// IsComment reports whether the trivia is a comment.
func (k TriviaPieceKind) IsComment() bool {
	return k == TriviaComment
}

// TriviaPiece describes one contiguous piece of trivia.
//
// A trivia piece stores only its kind and text length. Source text remains part of the GreenToken containing the piece.
type TriviaPiece struct {
	kind   TriviaPieceKind
	length text.TextSize
}

// NewTriviaPiece returns a trivia piece with the given kind and text length.
func NewTriviaPiece(kind TriviaPieceKind, length text.TextSize) TriviaPiece {
	validateTriviaPieceKind(kind)
	return TriviaPiece{
		kind:   kind,
		length: length,
	}
}

// NewNewlineTriviaPiece returns a newline trivia piece.
func NewNewlineTriviaPiece(length text.TextSize) TriviaPiece {
	return NewTriviaPiece(TriviaNewline, length)
}

// NewWhitespaceTriviaPiece returns a whitespace trivia piece.
func NewWhitespaceTriviaPiece(length text.TextSize) TriviaPiece {
	return NewTriviaPiece(TriviaWhitespace, length)
}

// NewCommentTriviaPiece returns a comment trivia piece.
func NewCommentTriviaPiece(length text.TextSize) TriviaPiece {
	return NewTriviaPiece(TriviaComment, length)
}

// Kind returns the trivia piece's kind.
func (p TriviaPiece) Kind() TriviaPieceKind {
	return p.kind
}

// TextLen returns the trivia piece's text length.
func (p TriviaPiece) TextLen() text.TextSize {
	return p.length
}

// greenTrivia is an immutable handle to a list of trivia pieces.
type greenTrivia struct {
	data *greenTriviaData
}

// greenTriviaData stores the trivia pieces referenced by greenTrivia.
type greenTriviaData struct {
	pieces []TriviaPiece
}

// newGreenTrivia returns green trivia containing pieces.
func newGreenTrivia(pieces []TriviaPiece) greenTrivia {
	if len(pieces) == 0 {
		return greenTrivia{}
	}
	ownedPieces := make([]TriviaPiece, len(pieces))
	copy(ownedPieces, pieces)
	var textLen text.TextSize
	for _, piece := range ownedPieces {
		validateTriviaPieceKind(piece.Kind())
		pieceLen := piece.TextLen()
		if pieceLen > ^text.TextSize(0)-textLen {
			panic("tree: trivia text exceeds maximum TextSize")
		}
		textLen += pieceLen
	}
	return greenTrivia{
		data: &greenTriviaData{
			pieces: ownedPieces,
		},
	}
}

// textLen returns the total text length of the trivia.
func (t greenTrivia) textLen() text.TextSize {
	if t.data == nil {
		return 0
	}
	var textLen text.TextSize
	// potential optimization path by keeping `textLen` in `greenTriviaData`
	for _, piece := range t.data.pieces {
		textLen += piece.TextLen()
	}
	return textLen
}

// len returns the number of trivia pieces.
func (t greenTrivia) len() int {
	if t.data == nil {
		return 0
	}
	return len(t.data.pieces)
}

// piece returns the trivia piece at index.
func (t greenTrivia) piece(index int) TriviaPiece {
	if t.data == nil {
		panic("tree: trivia piece index out of range")
	}
	return t.data.pieces[index]
}

func validateTriviaPieceKind(kind TriviaPieceKind) {
	switch kind {
	case TriviaNewline,
		TriviaWhitespace,
		TriviaComment:
		return
	default:
		panic("tree: invalid trivia piece kind")
	}
}
