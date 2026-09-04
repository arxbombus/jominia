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
	IdentifierFragment
	StringFragment

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
	Dot
	InlineMathStart
	At
	Dollar
	Pipe
	Plus
	Minus
	Star
	Slash
	Percent
	StringQuote
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
	ConditionalBlock
	ConditionalHeader
	InlineMath
	BracketExpression
	CallExpression
	ArgumentList
	MemberExpression
	FormatSpecifier
	NumberExpression
	BooleanExpression
	NameExpression
	StringExpression
	ParameterExpression
	VariableReference
	InterpolatedIdentifier
	InterpolatedString
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

// IsScalar reports whether k is a single-token scalar accepted by the normal
// grammar. Structured logical scalars are parsed contextually. Single-quoted
// strings are recognized only by the bracket-expression grammar.
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
	case IdentifierFragment:
		return "IdentifierFragment"
	case StringFragment:
		return "StringFragment"
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
	case Dot:
		return "Dot"
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
	case StringQuote:
		return "StringQuote"
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
	case ConditionalBlock:
		return "ConditionalBlock"
	case ConditionalHeader:
		return "ConditionalHeader"
	case InlineMath:
		return "InlineMath"
	case BracketExpression:
		return "BracketExpression"
	case CallExpression:
		return "CallExpression"
	case ArgumentList:
		return "ArgumentList"
	case MemberExpression:
		return "MemberExpression"
	case FormatSpecifier:
		return "FormatSpecifier"
	case NumberExpression:
		return "NumberExpression"
	case BooleanExpression:
		return "BooleanExpression"
	case NameExpression:
		return "NameExpression"
	case StringExpression:
		return "StringExpression"
	case ParameterExpression:
		return "ParameterExpression"
	case VariableReference:
		return "VariableReference"
	case InterpolatedIdentifier:
		return "InterpolatedIdentifier"
	case InterpolatedString:
		return "InterpolatedString"
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
