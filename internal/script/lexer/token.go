package lexer

import (
	"github.com/arxbombus/jominia/internal/script/syntax"
	"github.com/arxbombus/jominia/internal/text"
)

type Token struct {
	Kind  syntax.SyntaxKind
	Range text.TextRange
}
