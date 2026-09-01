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
	InlineMathStart
	At
	Dollar
	Pipe
	Plus
	Minus
	Star
	Slash
	Percent
	ParameterName
	ParameterArgument

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
	ConditionalBlock
	ConditionalHeader
	InlineMath
	NumberExpression
	NameExpression
	ParameterExpression
	UnaryExpression
	BinaryExpression
	ParenthesizedExpression
	AbsoluteExpression

	// Error-recovery nodes.
	BogusStatement
	BogusExpression
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
	return k == BogusStatement || k == BogusExpression
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
	case InlineMathStart:
		return "InlineMathStart"
	case At:
		return "At"
	case Dollar:
		return "Dollar"
	case Pipe:
		return "Pipe"
	case Plus:
		return "Plus"
	case Minus:
		return "Minus"
	case Star:
		return "Star"
	case Slash:
		return "Slash"
	case Percent:
		return "Percent"
	case ParameterName:
		return "ParameterName"
	case ParameterArgument:
		return "ParameterArgument"
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
	case ConditionalBlock:
		return "ConditionalBlock"
	case ConditionalHeader:
		return "ConditionalHeader"
	case InlineMath:
		return "InlineMath"
	case NumberExpression:
		return "NumberExpression"
	case NameExpression:
		return "NameExpression"
	case ParameterExpression:
		return "ParameterExpression"
	case UnaryExpression:
		return "UnaryExpression"
	case BinaryExpression:
		return "BinaryExpression"
	case ParenthesizedExpression:
		return "ParenthesizedExpression"
	case AbsoluteExpression:
		return "AbsoluteExpression"
	case BogusStatement:
		return "BogusStatement"
	case BogusExpression:
		return "BogusExpression"
	default:
		return "Unknown"
	}
}
