package lexer

import (
	"fmt"
	"testing"

	"github.com/arxbombus/jominia/internal/script/syntax"
)

type expectedToken struct {
	kind syntax.SyntaxKind
	text string
}

func TestLex(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   []expectedToken
	}{
		{
			name:   "empty",
			source: "",
			want: []expectedToken{
				{syntax.EOF, ""},
			},
		},
		{
			name:   "simple assignment",
			source: "country = GER",
			want: []expectedToken{
				{syntax.Identifier, "country"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.Identifier, "GER"},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "block and comment",
			source: "country = {\n\ttag = GER # hello\n}",
			want: []expectedToken{
				{syntax.Identifier, "country"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.LCurly, "{"},
				{syntax.Newline, "\n"},
				{syntax.Whitespace, "\t"},
				{syntax.Identifier, "tag"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.Identifier, "GER"},
				{syntax.Whitespace, " "},
				{syntax.Comment, "# hello"},
				{syntax.Newline, "\n"},
				{syntax.RCurly, "}"},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "operators",
			source: "= == != < <= > >= ? ?= !",
			want: []expectedToken{
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.EqualsEquals, "=="},
				{syntax.Whitespace, " "},
				{syntax.BangEquals, "!="},
				{syntax.Whitespace, " "},
				{syntax.Less, "<"},
				{syntax.Whitespace, " "},
				{syntax.LessEquals, "<="},
				{syntax.Whitespace, " "},
				{syntax.Greater, ">"},
				{syntax.Whitespace, " "},
				{syntax.GreaterEquals, ">="},
				{syntax.Whitespace, " "},
				{syntax.Question, "?"},
				{syntax.Whitespace, " "},
				{syntax.QuestionEquals, "?="},
				{syntax.Whitespace, " "},
				{syntax.Bang, "!"},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "numbers and identifiers",
			source: "123 -42 3.14 GER country 1936.1.1",
			want: []expectedToken{
				{syntax.Number, "123"},
				{syntax.Whitespace, " "},
				{syntax.Number, "-42"},
				{syntax.Whitespace, " "},
				{syntax.Number, "3.14"},
				{syntax.Whitespace, " "},
				{syntax.Identifier, "GER"},
				{syntax.Whitespace, " "},
				{syntax.Identifier, "country"},
				{syntax.Whitespace, " "},
				{syntax.Identifier, "1936.1.1"},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "double quoted string",
			source: `"hello world"`,
			want: []expectedToken{
				{syntax.String, `"hello world"`},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "escaped quote in double quoted string",
			source: `"hello \"world\""`,
			want: []expectedToken{
				{syntax.String, `"hello \"world\""`},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "unterminated double quoted string",
			source: `"hello`,
			want: []expectedToken{
				{syntax.ErrorToken, `"hello`},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "single quoted token",
			source: `'NO_WORLD_MARKET_ACCESS'`,
			want: []expectedToken{
				{syntax.SingleQuotedString, `'NO_WORLD_MARKET_ACCESS'`},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "escaped quote in single quoted token",
			source: `'can\'t'`,
			want: []expectedToken{
				{syntax.SingleQuotedString, `'can\'t'`},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "unterminated single quoted token",
			source: `'hello`,
			want: []expectedToken{
				{syntax.ErrorToken, `'hello`},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "unix newline",
			source: "foo\nbar",
			want: []expectedToken{
				{syntax.Identifier, "foo"},
				{syntax.Newline, "\n"},
				{syntax.Identifier, "bar"},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "windows newline",
			source: "foo\r\nbar",
			want: []expectedToken{
				{syntax.Identifier, "foo"},
				{syntax.Newline, "\r\n"},
				{syntax.Identifier, "bar"},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "carriage return newline",
			source: "foo\rbar",
			want: []expectedToken{
				{syntax.Identifier, "foo"},
				{syntax.Newline, "\r"},
				{syntax.Identifier, "bar"},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "unicode identifier",
			source: "café = 日本",
			want: []expectedToken{
				{syntax.Identifier, "café"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.Identifier, "日本"},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "semicolon",
			source: `foo=bar;`,
			want: []expectedToken{
				{syntax.Identifier, "foo"},
				{syntax.Equals, "="},
				{syntax.Identifier, "bar"},
				{syntax.Semicolon, ";"},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "boolean",
			source: "do_something = yes\nenabled = no",
			want: []expectedToken{
				{syntax.Identifier, "do_something"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.Boolean, "yes"},
				{syntax.Newline, "\n"},
				{syntax.Identifier, "enabled"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.Boolean, "no"},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "brackets",
			source: `@[1-leo_x]`,
			want: []expectedToken{
				{syntax.InlineMathStart, "@["},
				{syntax.Identifier, "1-leo_x"},
				{syntax.RBracket, "]"},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "parentheses",
			source: `Localize('NEWLINE')`,
			want: []expectedToken{
				{syntax.Identifier, "Localize"},
				{syntax.LParen, "("},
				{syntax.SingleQuotedString, `'NEWLINE'`},
				{syntax.RParen, ")"},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "brackets adjacent to identifier",
			source: `foo[bar]`,
			want: []expectedToken{
				{syntax.Identifier, "foo"},
				{syntax.LBracket, "["},
				{syntax.Identifier, "bar"},
				{syntax.RBracket, "]"},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "parentheses are atom boundaries",
			source: `foo(bar)baz`,
			want: []expectedToken{
				{syntax.Identifier, "foo"},
				{syntax.LParen, "("},
				{syntax.Identifier, "bar"},
				{syntax.RParen, ")"},
				{syntax.Identifier, "baz"},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "hyphens and variables remain atoms",
			source: `-42 dashed-identifier @my_var $VALUE$ $VALUE|100$ prefix_$VALUE$ $VALUE$_suffix`,
			want: []expectedToken{
				{syntax.Number, "-42"},
				{syntax.Whitespace, " "},
				{syntax.Identifier, "dashed-identifier"},
				{syntax.Whitespace, " "},
				{syntax.Identifier, "@my_var"},
				{syntax.Whitespace, " "},
				{syntax.Identifier, "$VALUE$"},
				{syntax.Whitespace, " "},
				{syntax.Identifier, "$VALUE|100$"},
				{syntax.Whitespace, " "},
				{syntax.Identifier, "prefix_$VALUE$"},
				{syntax.Whitespace, " "},
				{syntax.Identifier, "$VALUE$_suffix"},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "compact boundary characters",
			source: `a={b="1"c=d}foo=bar#good`,
			want: []expectedToken{
				{syntax.Identifier, "a"},
				{syntax.Equals, "="},
				{syntax.LCurly, "{"},
				{syntax.Identifier, "b"},
				{syntax.Equals, "="},
				{syntax.String, `"1"`},
				{syntax.Identifier, "c"},
				{syntax.Equals, "="},
				{syntax.Identifier, "d"},
				{syntax.RCurly, "}"},
				{syntax.Identifier, "foo"},
				{syntax.Equals, "="},
				{syntax.Identifier, "bar"},
				{syntax.Comment, "#good"},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "vic 3 gui expression tokens",
			source: `[ConcatIfNeitherEmpty(AddLocalizationIf(Condition, LocKey), Localize( 'NEWLINE' ))]`,
			want: []expectedToken{
				{syntax.LBracket, "["},
				{syntax.Identifier, "ConcatIfNeitherEmpty"},
				{syntax.LParen, "("},
				{syntax.Identifier, "AddLocalizationIf"},
				{syntax.LParen, "("},
				{syntax.Identifier, "Condition"},
				{syntax.Comma, ","},
				{syntax.Whitespace, " "},
				{syntax.Identifier, "LocKey"},
				{syntax.RParen, ")"},
				{syntax.Comma, ","},
				{syntax.Whitespace, " "},
				{syntax.Identifier, "Localize"},
				{syntax.LParen, "("},
				{syntax.Whitespace, " "},
				{syntax.SingleQuotedString, `'NEWLINE'`},
				{syntax.Whitespace, " "},
				{syntax.RParen, ")"},
				{syntax.RParen, ")"},
				{syntax.RBracket, "]"},
				{syntax.EOF, ""},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertTokens(t, test.source, test.want)
		})
	}
}

func TestReLexInlineMath(t *testing.T) {
	const source = `@[1-leo_x*($FACTOR|2$+@foo)%3] next-value`
	l := NewLexer(source)
	opener := l.Next()
	if opener.Kind != syntax.InlineMathStart {
		t.Fatalf("opener kind = %s, want InlineMathStart", opener.Kind)
	}
	current := l.Next()
	if current.Kind != syntax.Identifier {
		t.Fatalf("normal interior kind = %s, want Identifier", current.Kind)
	}
	current = l.ReLex(current, ReLexInlineMath)
	got := []Token{opener}
	for {
		got = append(got, current)
		if current.Kind == syntax.EOF {
			break
		}
		current = l.Next()
	}
	want := []expectedToken{
		{syntax.InlineMathStart, "@["},
		{syntax.Number, "1"},
		{syntax.Minus, "-"},
		{syntax.Identifier, "leo_x"},
		{syntax.Star, "*"},
		{syntax.LParen, "("},
		{syntax.Dollar, "$"},
		{syntax.ParameterName, "FACTOR"},
		{syntax.Pipe, "|"},
		{syntax.ParameterArgument, "2"},
		{syntax.Dollar, "$"},
		{syntax.Plus, "+"},
		{syntax.At, "@"},
		{syntax.Identifier, "foo"},
		{syntax.RParen, ")"},
		{syntax.Percent, "%"},
		{syntax.Number, "3"},
		{syntax.RBracket, "]"},
		{syntax.Whitespace, " "},
		{syntax.Identifier, "next-value"},
		{syntax.EOF, ""},
	}
	if len(got) != len(want) {
		t.Fatalf("token count = %d, want %d\n%s", len(got), len(want), formatTokens(source, got))
	}
	for index, token := range got {
		expected := want[index]
		if token.Kind != expected.kind {
			t.Fatalf("token %d kind = %s, want %s\n%s", index, token.Kind, expected.kind, formatTokens(source, got))
		}
		if text := source[int(token.Range.Start()):int(token.Range.End())]; text != expected.text {
			t.Fatalf("token %d text = %q, want %q", index, text, expected.text)
		}
	}
}

func TestNextWithBracketExpressionContext(t *testing.T) {
	const source = `[Root?.GetValue(3.14, yes, 'KEY', $DEFAULT$)|0] next.value`
	l := NewLexer(source)
	got := []Token{l.Next()}
	current := l.NextWithContext(LexBracketExpression)
	for {
		got = append(got, current)
		if current.Kind == syntax.EOF {
			break
		}
		if current.Kind == syntax.RBracket {
			current = l.NextWithContext(LexNormal)
		} else {
			current = l.Next()
		}
	}
	want := []expectedToken{
		{syntax.LBracket, "["},
		{syntax.Identifier, "Root"},
		{syntax.Question, "?"},
		{syntax.Dot, "."},
		{syntax.Identifier, "GetValue"},
		{syntax.LParen, "("},
		{syntax.Number, "3.14"},
		{syntax.Comma, ","},
		{syntax.Whitespace, " "},
		{syntax.Boolean, "yes"},
		{syntax.Comma, ","},
		{syntax.Whitespace, " "},
		{syntax.SingleQuotedString, "'KEY'"},
		{syntax.Comma, ","},
		{syntax.Whitespace, " "},
		{syntax.Dollar, "$"},
		{syntax.ParameterName, "DEFAULT"},
		{syntax.Dollar, "$"},
		{syntax.RParen, ")"},
		{syntax.Pipe, "|"},
		{syntax.Number, "0"},
		{syntax.RBracket, "]"},
		{syntax.Whitespace, " "},
		{syntax.Identifier, "next.value"},
		{syntax.EOF, ""},
	}
	assertTokenSequence(t, source, got, want)
}

func TestBracketContextInsideInterpolatedString(t *testing.T) {
	const source = `"before [Root.Get($VALUE$)|0] after" next`
	l := NewLexer(source)
	current := l.Next()
	current = l.ReLex(current, ReLexInterpolatedString)
	var got []Token
	for {
		got = append(got, current)
		if current.Kind == syntax.EOF {
			break
		}
		switch current.Kind { //nolint:exhaustive // Only delimiters change the next-token context in this test driver.
		case syntax.LBracket:
			current = l.NextWithContext(LexBracketExpression)
		case syntax.RBracket:
			current = l.NextWithContext(LexInterpolatedString)
		default:
			current = l.Next()
		}
	}
	want := []expectedToken{
		{syntax.StringQuote, `"`},
		{syntax.StringFragment, "before "},
		{syntax.LBracket, "["},
		{syntax.Identifier, "Root"},
		{syntax.Dot, "."},
		{syntax.Identifier, "Get"},
		{syntax.LParen, "("},
		{syntax.Dollar, "$"},
		{syntax.ParameterName, "VALUE"},
		{syntax.Dollar, "$"},
		{syntax.RParen, ")"},
		{syntax.Pipe, "|"},
		{syntax.Number, "0"},
		{syntax.RBracket, "]"},
		{syntax.StringFragment, " after"},
		{syntax.StringQuote, `"`},
		{syntax.Whitespace, " "},
		{syntax.Identifier, "next"},
		{syntax.EOF, ""},
	}
	assertTokenSequence(t, source, got, want)
}

func TestReLexParameterAndRestoreNormalMode(t *testing.T) {
	const source = `$VALUE|100$ next-value = yes`
	l := NewLexer(source)
	current := l.Next()
	if current.Kind != syntax.Identifier {
		t.Fatalf("normal parameter kind = %s, want Identifier", current.Kind)
	}
	current = l.ReLex(current, ReLexParameter)
	got := []Token{}
	for {
		got = append(got, current)
		if current.Kind == syntax.EOF {
			break
		}
		current = l.Next()
	}
	want := []expectedToken{
		{syntax.Dollar, "$"},
		{syntax.ParameterName, "VALUE"},
		{syntax.Pipe, "|"},
		{syntax.ParameterArgument, "100"},
		{syntax.Dollar, "$"},
		{syntax.Whitespace, " "},
		{syntax.Identifier, "next-value"},
		{syntax.Whitespace, " "},
		{syntax.Equals, "="},
		{syntax.Whitespace, " "},
		{syntax.Boolean, "yes"},
		{syntax.EOF, ""},
	}
	assertTokenSequence(t, source, got, want)
}

func TestReLexMalformedParameterRestoresNormalModeAtLineBreak(t *testing.T) {
	const source = "$VALUE\nnext = yes"
	l := NewLexer(source)
	current := l.Next()
	current = l.ReLex(current, ReLexParameter)
	got := []Token{}
	for {
		got = append(got, current)
		if current.Kind == syntax.EOF {
			break
		}
		current = l.Next()
	}
	want := []expectedToken{
		{syntax.Dollar, "$"},
		{syntax.ParameterName, "VALUE"},
		{syntax.Newline, "\n"},
		{syntax.Identifier, "next"},
		{syntax.Whitespace, " "},
		{syntax.Equals, "="},
		{syntax.Whitespace, " "},
		{syntax.Boolean, "yes"},
		{syntax.EOF, ""},
	}
	assertTokenSequence(t, source, got, want)
}

func TestReLexInterpolatedIdentifierAndRestoreNormalMode(t *testing.T) {
	const source = `ethic_$ETHIC$_$TIER|1$ next-value`
	l := NewLexer(source)
	current := l.Next()
	if current.Kind != syntax.Identifier {
		t.Fatalf("normal interpolated identifier = %s, want Identifier", current.Kind)
	}
	current = l.ReLex(current, ReLexInterpolatedIdentifier)
	got := []Token{}
	for {
		got = append(got, current)
		if current.Kind == syntax.EOF {
			break
		}
		current = l.Next()
	}
	want := []expectedToken{
		{syntax.IdentifierFragment, "ethic_"},
		{syntax.Dollar, "$"},
		{syntax.ParameterName, "ETHIC"},
		{syntax.Dollar, "$"},
		{syntax.IdentifierFragment, "_"},
		{syntax.Dollar, "$"},
		{syntax.ParameterName, "TIER"},
		{syntax.Pipe, "|"},
		{syntax.ParameterArgument, "1"},
		{syntax.Dollar, "$"},
		{syntax.Whitespace, " "},
		{syntax.Identifier, "next-value"},
		{syntax.EOF, ""},
	}
	assertTokenSequence(t, source, got, want)
}

func TestReLexInterpolatedStringKeepsInlineMathAsText(t *testing.T) {
	const source = `"$KEY$_a @\[ $COUNT$ * 500]" next`
	l := NewLexer(source)
	current := l.Next()
	if current.Kind != syntax.String {
		t.Fatalf("normal interpolated string = %s, want String", current.Kind)
	}
	current = l.ReLex(current, ReLexInterpolatedString)
	got := []Token{}
	for {
		got = append(got, current)
		if current.Kind == syntax.EOF {
			break
		}
		current = l.Next()
	}
	want := []expectedToken{
		{syntax.StringQuote, `"`},
		{syntax.Dollar, "$"},
		{syntax.ParameterName, "KEY"},
		{syntax.Dollar, "$"},
		{syntax.StringFragment, `_a @\[ `},
		{syntax.Dollar, "$"},
		{syntax.ParameterName, "COUNT"},
		{syntax.Dollar, "$"},
		{syntax.StringFragment, ` * 500]`},
		{syntax.StringQuote, `"`},
		{syntax.Whitespace, " "},
		{syntax.Identifier, "next"},
		{syntax.EOF, ""},
	}
	assertTokenSequence(t, source, got, want)
}

func TestReLexMalformedInterpolationStopsAtOriginalAtom(t *testing.T) {
	const source = "prefix_$NAME\nnext = yes"
	l := NewLexer(source)
	current := l.Next()
	current = l.ReLex(current, ReLexInterpolatedIdentifier)
	got := []Token{}
	for {
		got = append(got, current)
		if current.Kind == syntax.EOF {
			break
		}
		current = l.Next()
	}
	want := []expectedToken{
		{syntax.IdentifierFragment, "prefix_"},
		{syntax.Dollar, "$"},
		{syntax.ParameterName, "NAME"},
		{syntax.Newline, "\n"},
		{syntax.Identifier, "next"},
		{syntax.Whitespace, " "},
		{syntax.Equals, "="},
		{syntax.Whitespace, " "},
		{syntax.Boolean, "yes"},
		{syntax.EOF, ""},
	}
	assertTokenSequence(t, source, got, want)
}

func TestReLexMissingInterpolationArgumentStopsAtOriginalAtom(t *testing.T) {
	const source = `0$|#comment`
	l := NewLexer(source)
	current := l.Next()
	current = l.ReLex(current, ReLexInterpolatedIdentifier)
	got := []Token{}
	for {
		got = append(got, current)
		if current.Kind == syntax.EOF {
			break
		}
		current = l.Next()
	}
	want := []expectedToken{
		{syntax.IdentifierFragment, "0"},
		{syntax.Dollar, "$"},
		{syntax.Pipe, "|"},
		{syntax.Comment, "#comment"},
		{syntax.EOF, ""},
	}
	assertTokenSequence(t, source, got, want)
}

func TestReLexVariableReferenceAndRestoreNormalMode(t *testing.T) {
	const source = `@example = 2`
	l := NewLexer(source)
	current := l.Next()
	if current.Kind != syntax.Identifier {
		t.Fatalf("normal variable = %s, want Identifier", current.Kind)
	}
	current = l.ReLex(current, ReLexVariableReference)
	got := []Token{}
	for {
		got = append(got, current)
		if current.Kind == syntax.EOF {
			break
		}
		current = l.Next()
	}
	want := []expectedToken{
		{syntax.At, "@"},
		{syntax.Identifier, "example"},
		{syntax.Whitespace, " "},
		{syntax.Equals, "="},
		{syntax.Whitespace, " "},
		{syntax.Number, "2"},
		{syntax.EOF, ""},
	}
	assertTokenSequence(t, source, got, want)
}

func TestLexRecognizesEscapedInlineMathStart(t *testing.T) {
	assertTokens(t, `@\[1]`, []expectedToken{
		{syntax.InlineMathStart, `@\[`},
		{syntax.Number, "1"},
		{syntax.RBracket, "]"},
		{syntax.EOF, ""},
	})
}

func assertTokens(t *testing.T, source string, want []expectedToken) {
	t.Helper()
	got := Lex(source)
	assertTokenSequence(t, source, got, want)
}

func assertTokenSequence(t *testing.T, source string, got []Token, want []expectedToken) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf(
			"token count = %d, want %d\n%s",
			len(got),
			len(want),
			formatTokens(source, got),
		)
	}
	for i, token := range got {
		expected := want[i]

		if token.Kind != expected.kind {
			t.Fatalf(
				"token %d kind = %v, want %v\n%s",
				i,
				token.Kind,
				expected.kind,
				formatTokens(source, got),
			)
		}
		start := int(token.Range.Start())
		end := int(token.Range.End())
		gotText := source[start:end]
		if gotText != expected.text {
			t.Fatalf(
				"token %d text = %q, want %q\n%s",
				i,
				gotText,
				expected.text,
				formatTokens(source, got),
			)
		}
	}
}

func formatTokens(source string, tokens []Token) string {
	var result string
	for _, token := range tokens {
		start := int(token.Range.Start())
		end := int(token.Range.End())
		result += fmt.Sprintf(
			"%v [%d,%d) %q\n",
			token.Kind,
			start,
			end,
			source[start:end],
		)
	}
	return result
}
