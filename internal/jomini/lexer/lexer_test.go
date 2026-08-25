package lexer

import (
	"fmt"
	"testing"

	"github.com/arxbombus/jominia/internal/jomini/syntax"
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
				{syntax.ErrorToken, "!"},
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
			name:   "string",
			source: `"hello world"`,
			want: []expectedToken{
				{syntax.String, `"hello world"`},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "escaped quote in string",
			source: `"hello \"world\""`,
			want: []expectedToken{
				{syntax.String, `"hello \"world\""`},
				{syntax.EOF, ""},
			},
		},
		{
			name:   "unterminated string",
			source: `"hello`,
			want: []expectedToken{
				{syntax.ErrorToken, `"hello`},
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
			name: "vic 3 state",
			source: `STATE_LOMBARDY = {
    id = 76
    subsistence_building = "building_subsistence_farm"
    provinces = { "x3F1E38" "x4713EE" "x50C060" "x70B8A9" "x867A90" "x9AC196" "xA40CE9" "xD04060" }
    traits = { "state_trait_po_river" }
    city = "xD04060"
    farm = "x867A90"
    mine = "x3F1E38"
    wood = "x4713EE"
    arable_land = 110
    arable_resources = { "building_wheat_farm" "building_livestock_ranch" "building_cotton_plantation" "building_silk_plantation" "building_vineyard" }
    capped_resources = {
        building_iron_mine = 27
        building_lead_mine = 19
        building_logging_camp = 7
    }
    resource = {
        type = "building_oil_rig"
        undiscovered_amount = 20
    }
}`,
			want: []expectedToken{
				{syntax.Identifier, "STATE_LOMBARDY"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.LCurly, "{"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "    "},
				{syntax.Identifier, "id"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.Number, "76"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "    "},
				{syntax.Identifier, "subsistence_building"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.String, `"building_subsistence_farm"`},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "    "},
				{syntax.Identifier, "provinces"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.LCurly, "{"},
				{syntax.Whitespace, " "},
				{syntax.String, `"x3F1E38"`},
				{syntax.Whitespace, " "},
				{syntax.String, `"x4713EE"`},
				{syntax.Whitespace, " "},
				{syntax.String, `"x50C060"`},
				{syntax.Whitespace, " "},
				{syntax.String, `"x70B8A9"`},
				{syntax.Whitespace, " "},
				{syntax.String, `"x867A90"`},
				{syntax.Whitespace, " "},
				{syntax.String, `"x9AC196"`},
				{syntax.Whitespace, " "},
				{syntax.String, `"xA40CE9"`},
				{syntax.Whitespace, " "},
				{syntax.String, `"xD04060"`},
				{syntax.Whitespace, " "},
				{syntax.RCurly, "}"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "    "},
				{syntax.Identifier, "traits"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.LCurly, "{"},
				{syntax.Whitespace, " "},
				{syntax.String, `"state_trait_po_river"`},
				{syntax.Whitespace, " "},
				{syntax.RCurly, "}"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "    "},
				{syntax.Identifier, "city"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.String, `"xD04060"`},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "    "},
				{syntax.Identifier, "farm"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.String, `"x867A90"`},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "    "},
				{syntax.Identifier, "mine"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.String, `"x3F1E38"`},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "    "},
				{syntax.Identifier, "wood"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.String, `"x4713EE"`},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "    "},
				{syntax.Identifier, "arable_land"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.Number, "110"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "    "},
				{syntax.Identifier, "arable_resources"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.LCurly, "{"},
				{syntax.Whitespace, " "},
				{syntax.String, `"building_wheat_farm"`},
				{syntax.Whitespace, " "},
				{syntax.String, `"building_livestock_ranch"`},
				{syntax.Whitespace, " "},
				{syntax.String, `"building_cotton_plantation"`},
				{syntax.Whitespace, " "},
				{syntax.String, `"building_silk_plantation"`},
				{syntax.Whitespace, " "},
				{syntax.String, `"building_vineyard"`},
				{syntax.Whitespace, " "},
				{syntax.RCurly, "}"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "    "},
				{syntax.Identifier, "capped_resources"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.LCurly, "{"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "        "},
				{syntax.Identifier, "building_iron_mine"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.Number, "27"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "        "},
				{syntax.Identifier, "building_lead_mine"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.Number, "19"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "        "},
				{syntax.Identifier, "building_logging_camp"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.Number, "7"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "    "},
				{syntax.RCurly, "}"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "    "},
				{syntax.Identifier, "resource"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.LCurly, "{"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "        "},
				{syntax.Identifier, "type"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.String, `"building_oil_rig"`},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "        "},
				{syntax.Identifier, "undiscovered_amount"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.Number, "20"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "    "},
				{syntax.RCurly, "}"},
				{syntax.Newline, "\n"},

				{syntax.RCurly, "}"},
				{syntax.EOF, ""},
			},
		},
		{
			name: "vic 3 gui",
			source: `types wargoal_types
{
	type add_wargoal_panel = default_block_window  {
		name = "add_wargoal_panel"
		datacontext = "[AddWarGoalPanel.AccessDiplomaticPlay]"

		blockoverride "window_header_name" {
			text = "ADD_WARGOAL_HEADER"
		}

		blockoverride "entire_back_button" {
			back_button_large = {
				position = { 8 30 }
				onclick = "[AddWarGoalPanel.ClearSelectedWarGoalType]"
				input_action = "back"
				visible = "[AddWarGoalPanel.HasSelectedWarGoalType]"
			}

			back_button_large = {
				position = { 8 30 }
				onclick = "[InformationPanelBar.OpenPreviousPanel]"
				input_action = "back"
				visible = "[Not(AddWarGoalPanel.HasSelectedWarGoalType)]"
			}
		}
	}
}`,
			want: []expectedToken{
				{syntax.Identifier, "types"},
				{syntax.Whitespace, " "},
				{syntax.Identifier, "wargoal_types"},
				{syntax.Newline, "\n"},

				{syntax.LCurly, "{"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t"},
				{syntax.Identifier, "type"},
				{syntax.Whitespace, " "},
				{syntax.Identifier, "add_wargoal_panel"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.Identifier, "default_block_window"},
				{syntax.Whitespace, "  "},
				{syntax.LCurly, "{"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t"},
				{syntax.Identifier, "name"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.String, `"add_wargoal_panel"`},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t"},
				{syntax.Identifier, "datacontext"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.String, `"[AddWarGoalPanel.AccessDiplomaticPlay]"`},
				{syntax.Newline, "\n"},

				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t"},
				{syntax.Identifier, "blockoverride"},
				{syntax.Whitespace, " "},
				{syntax.String, `"window_header_name"`},
				{syntax.Whitespace, " "},
				{syntax.LCurly, "{"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t\t"},
				{syntax.Identifier, "text"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.String, `"ADD_WARGOAL_HEADER"`},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t"},
				{syntax.RCurly, "}"},
				{syntax.Newline, "\n"},

				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t"},
				{syntax.Identifier, "blockoverride"},
				{syntax.Whitespace, " "},
				{syntax.String, `"entire_back_button"`},
				{syntax.Whitespace, " "},
				{syntax.LCurly, "{"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t\t"},
				{syntax.Identifier, "back_button_large"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.LCurly, "{"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t\t\t"},
				{syntax.Identifier, "position"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.LCurly, "{"},
				{syntax.Whitespace, " "},
				{syntax.Number, "8"},
				{syntax.Whitespace, " "},
				{syntax.Number, "30"},
				{syntax.Whitespace, " "},
				{syntax.RCurly, "}"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t\t\t"},
				{syntax.Identifier, "onclick"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.String, `"[AddWarGoalPanel.ClearSelectedWarGoalType]"`},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t\t\t"},
				{syntax.Identifier, "input_action"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.String, `"back"`},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t\t\t"},
				{syntax.Identifier, "visible"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.String, `"[AddWarGoalPanel.HasSelectedWarGoalType]"`},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t\t"},
				{syntax.RCurly, "}"},
				{syntax.Newline, "\n"},

				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t\t"},
				{syntax.Identifier, "back_button_large"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.LCurly, "{"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t\t\t"},
				{syntax.Identifier, "position"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.LCurly, "{"},
				{syntax.Whitespace, " "},
				{syntax.Number, "8"},
				{syntax.Whitespace, " "},
				{syntax.Number, "30"},
				{syntax.Whitespace, " "},
				{syntax.RCurly, "}"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t\t\t"},
				{syntax.Identifier, "onclick"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.String, `"[InformationPanelBar.OpenPreviousPanel]"`},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t\t\t"},
				{syntax.Identifier, "input_action"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.String, `"back"`},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t\t\t"},
				{syntax.Identifier, "visible"},
				{syntax.Whitespace, " "},
				{syntax.Equals, "="},
				{syntax.Whitespace, " "},
				{syntax.String, `"[Not(AddWarGoalPanel.HasSelectedWarGoalType)]"`},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t\t"},
				{syntax.RCurly, "}"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t\t"},
				{syntax.RCurly, "}"},
				{syntax.Newline, "\n"},

				{syntax.Whitespace, "\t"},
				{syntax.RCurly, "}"},
				{syntax.Newline, "\n"},

				{syntax.RCurly, "}"},
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

func assertTokens(t *testing.T, source string, want []expectedToken) {
	t.Helper()

	got := Lex(source)

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
