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
	Boolean

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
	Comma

	// Syntax nodes.
	Root
	StatementList
	ValueStatement
	BinaryStatement
	BlockStatement
	BlockHeader
	ScalarList
	ValueList
	Block
	BracketGroup
	ParenGroup

	// Error-recovery nodes.
	BogusStatement
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
	return k == BogusStatement
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
func (k SyntaxKind) IsScalar() bool {
	return k == Identifier ||
		k == Number ||
		k == String ||
		k == Boolean
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
	case Boolean:
		return "Boolean"
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
	case Comma:
		return "Comma"
	case Root:
		return "Root"
	case StatementList:
		return "StatementList"
	case ValueStatement:
		return "ValueStatement"
	case BinaryStatement:
		return "BinaryStatement"
	case BlockStatement:
		return "BlockStatement"
	case BlockHeader:
		return "BlockHeader"
	case ScalarList:
		return "ScalarList"
	case ValueList:
		return "ValueList"
	case Block:
		return "Block"
	case BracketGroup:
		return "BracketGroup"
	case ParenGroup:
		return "ParenGroup"
	case BogusStatement:
		return "BogusStatement"
	default:
		return "Unknown"
	}
}
