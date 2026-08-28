package syntax

import "github.com/arxbombus/jominia/internal/tree"

type SyntaxKind uint16

const (
	Tombstone SyntaxKind = iota
	EOF
	ErrorToken

	// Trivia.
	Whitespace
	Newline
	Comment

	// Literals.
	Identifier
	Number
	String
	SingleQuotedString

	// Punctuation and operators.
	LCurly
	RCurly
	LBracket
	RBracket
	LParen
	RParen
	Equals
	EqualsEquals
	Bang
	BangEquals
	Less
	LessEquals
	Greater
	GreaterEquals
	Question
	QuestionEquals
	Semicolon

	// Syntax nodes.
	Root
	Entry
	Block
	BracketGroup
	ParenGroup

	// Error-recovery nodes.
	Bogus
)

func (k SyntaxKind) Raw() tree.RawSyntaxKind {
	return tree.RawSyntaxKind(k)
}

func FromRaw(k tree.RawSyntaxKind) SyntaxKind {
	return SyntaxKind(k)
}

func (k SyntaxKind) IsTrivia() bool {
	return k == Whitespace || k == Newline || k == Comment
}

func (k SyntaxKind) IsBogus() bool {
	return k == Bogus
}

// IsOperator reports whether k is an operator.
func (k SyntaxKind) IsOperator() bool {
	return k == Equals ||
		k == EqualsEquals ||
		k == BangEquals ||
		k == Less ||
		k == LessEquals ||
		k == Greater ||
		k == GreaterEquals ||
		k == QuestionEquals
}

// IsScalar reports whether k is a scalar accepted by the normal grammar.
// Single-quoted strings are currently only recognized inside opaque groups.
// valid in gui / loc (inside bracket and paren group): [MakeLineIf( IsZero(State.GetTradeCapacity), 'NO_WORLD_MARKET_ACCESS_DUE_TO_NO_TRADE_CAPACITY')]
func (k SyntaxKind) IsScalar() bool {
	return k == Identifier ||
		k == Number ||
		k == String
}

func (k SyntaxKind) String() string {
	switch k {
	case Tombstone:
		return "Tombstone"
	case EOF:
		return "EOF"
	case ErrorToken:
		return "ErrorToken"
	case Whitespace:
		return "Whitespace"
	case Newline:
		return "Newline"
	case Comment:
		return "Comment"
	case Identifier:
		return "Identifier"
	case Number:
		return "Number"
	case String:
		return "String"
	case SingleQuotedString:
		return "SingleQuotedString"
	case LCurly:
		return "LCurly"
	case RCurly:
		return "RCurly"
	case LBracket:
		return "LBracket"
	case RBracket:
		return "RBracket"
	case LParen:
		return "LParen"
	case RParen:
		return "RParen"
	case Equals:
		return "Equals"
	case EqualsEquals:
		return "EqualsEquals"
	case Bang:
		return "Bang"
	case BangEquals:
		return "BangEquals"
	case Less:
		return "Less"
	case LessEquals:
		return "LessEquals"
	case Greater:
		return "Greater"
	case GreaterEquals:
		return "GreaterEquals"
	case Question:
		return "Question"
	case QuestionEquals:
		return "QuestionEquals"
	case Semicolon:
		return "Semicolon"
	case Root:
		return "Root"
	case Entry:
		return "Entry"
	case Block:
		return "Block"
	case BracketGroup:
		return "BracketGroup"
	case ParenGroup:
		return "ParenGroup"
	case Bogus:
		return "Bogus"
	default:
		return "Unknown"
	}
}
