package syntax

import "testing"

func TestTombstoneIsZeroValue(t *testing.T) {
	var kind SyntaxKind
	if kind != Tombstone {
		t.Fatalf("zero value of SyntaxKind = %d, want Tombstone (%d)", kind, Tombstone)
	}
}

func TestSyntaxKindRawRoundTrip(t *testing.T) {
	kind := BinaryStatement

	if got := FromRaw(kind.Raw()); got != kind {
		t.Fatalf("round trip = %d, want %d", got, kind)
	}
}

func TestTriviaKinds(t *testing.T) {
	trivia := []SyntaxKind{
		Whitespace,
		Newline,
		Comment,
	}
	for _, kind := range trivia {
		if !kind.IsTrivia() {
			t.Errorf("%s should be trivia", kind)
		}
	}
	nonTrivia := []SyntaxKind{
		Identifier,
		Equals,
		StatementList,
	}
	for _, kind := range nonTrivia {
		if kind.IsTrivia() {
			t.Errorf("%s should not be trivia", kind)
		}
	}
}

func TestOperatorKinds(t *testing.T) {
	operators := []SyntaxKind{
		Equals,
		EqualsEquals,
		BangEquals,
		Less,
		LessEquals,
		Greater,
		GreaterEquals,
		QuestionEquals,
	}
	for _, kind := range operators {
		if !kind.IsOperator() {
			t.Errorf("%s should be an operator", kind)
		}
	}
	nonOperators := []SyntaxKind{
		Bang,
		Question,
		Identifier,
		LParen,
	}
	for _, kind := range nonOperators {
		if kind.IsOperator() {
			t.Errorf("%s should not be an operator", kind)
		}
	}
}

func TestScalarKinds(t *testing.T) {
	scalars := []SyntaxKind{
		Identifier,
		Number,
		String,
		Boolean,
	}
	for _, kind := range scalars {
		if !kind.IsScalar() {
			t.Errorf("%s should be a scalar", kind)
		}
	}
	nonScalars := []SyntaxKind{
		SingleQuotedString,
		Equals,
		LCurly,
		LBracket,
		LParen,
	}
	for _, kind := range nonScalars {
		if kind.IsScalar() {
			t.Errorf("%s should not be a scalar", kind)
		}
	}
}

func TestBogusKinds(t *testing.T) {
	if !BogusStatement.IsBogus() {
		t.Error("BogusStatement should be bogus")
	}
	if !BogusExpression.IsBogus() {
		t.Error("BogusExpression should be bogus")
	}
	if BinaryStatement.IsBogus() {
		t.Error("BinaryStatement should not be bogus")
	}
}

func TestSyntaxKindString(t *testing.T) {
	tests := []struct {
		kind SyntaxKind
		want string
	}{
		{SingleQuotedString, "SingleQuotedString"},
		{IdentifierFragment, "IdentifierFragment"},
		{StringFragment, "StringFragment"},
		{StringQuote, "StringQuote"},
		{LParen, "LParen"},
		{RParen, "RParen"},
		{Dot, "Dot"},
		{Bang, "Bang"},
		{InlineMathStart, "InlineMathStart"},
		{ParameterArgument, "ParameterArgument"},
		{Root, "Root"},
		{StatementList, "StatementList"},
		{ValueStatement, "ValueStatement"},
		{BinaryStatement, "BinaryStatement"},
		{BlockStatement, "BlockStatement"},
		{BlockHeader, "BlockHeader"},
		{ScalarList, "ScalarList"},
		{ValueList, "ValueList"},
		{Block, "Block"},
		{ConditionalBlock, "ConditionalBlock"},
		{ConditionalHeader, "ConditionalHeader"},
		{InlineMath, "InlineMath"},
		{BracketExpression, "BracketExpression"},
		{CallExpression, "CallExpression"},
		{ArgumentList, "ArgumentList"},
		{MemberExpression, "MemberExpression"},
		{FormatSpecifier, "FormatSpecifier"},
		{BooleanExpression, "BooleanExpression"},
		{StringExpression, "StringExpression"},
		{VariableReference, "VariableReference"},
		{InterpolatedIdentifier, "InterpolatedIdentifier"},
		{InterpolatedString, "InterpolatedString"},
		{BinaryExpression, "BinaryExpression"},
		{AbsoluteExpression, "AbsoluteExpression"},
		{BogusStatement, "BogusStatement"},
		{BogusExpression, "BogusExpression"},
	}
	for _, test := range tests {
		if got := test.kind.String(); got != test.want {
			t.Errorf("%d.String() = %q, want %q", test.kind, got, test.want)
		}
	}
}
