package syntax

import "testing"

func TestTombstoneIsZeroValue(t *testing.T) {
	var kind SyntaxKind

	if kind != Tombstone {
		t.Fatalf("zero value of SyntaxKind = %d, want Tombstone (%d)", kind, Tombstone)
	}
}

func TestSyntaxKindRawRoundTrip(t *testing.T) {
	kind := Identifier

	if got := FromRaw(kind.Raw()); got != kind {
		t.Fatalf("round trip = %d, want %d", got, kind)
	}
}

func TestTriviaKinds(t *testing.T) {
	if !Whitespace.IsTrivia() {
		t.Error("Whitespace should be trivia")
	}

	if !Newline.IsTrivia() {
		t.Error("Newline should be trivia")
	}

	if !Comment.IsTrivia() {
		t.Error("Comment should be trivia")
	}

	if Identifier.IsTrivia() {
		t.Error("Identifier should not be trivia")
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
	}

	for _, kind := range scalars {
		if !kind.IsScalar() {
			t.Errorf("%s should be a scalar", kind)
		}
	}

	// Single-quoted strings are only meaningful inside opaque bracket/paren
	// groups for now, so the normal grammar must not treat them as scalars.
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

func TestBogusKind(t *testing.T) {
	if !Bogus.IsBogus() {
		t.Error("Bogus should be bogus")
	}

	if Identifier.IsBogus() {
		t.Error("Identifier should not be bogus")
	}
}

func TestSyntaxKindString(t *testing.T) {
	tests := []struct {
		kind SyntaxKind
		want string
	}{
		{SingleQuotedString, "SingleQuotedString"},
		{LParen, "LParen"},
		{RParen, "RParen"},
		{Bang, "Bang"},
		{BracketGroup, "BracketGroup"},
		{ParenGroup, "ParenGroup"},
	}

	for _, test := range tests {
		if got := test.kind.String(); got != test.want {
			t.Errorf("%d.String() = %q, want %q", test.kind, got, test.want)
		}
	}
}
