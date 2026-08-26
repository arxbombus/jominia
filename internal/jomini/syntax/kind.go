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
	// Punctuation and operators.
	LCurly
	RCurly
	LBracket
	RBracket
	Equals
	EqualsEquals
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
	case LCurly:
		return "LCurly"
	case RCurly:
		return "RCurly"
	case LBracket:
		return "LBracket"
	case RBracket:
		return "RBracket"
	case Equals:
		return "Equals"
	case EqualsEquals:
		return "EqualsEquals"
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
	case Bogus:
		return "Bogus"
	default:
		return "Unknown"
	}
}
