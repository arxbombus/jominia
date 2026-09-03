package parser

import (
	"os"
	"testing"
)

func FuzzParseLossless(f *testing.F) {
	seeds := []string{
		"",
		"foo = bar",
		`a={b="1"c=d}foo=bar#good`,
		`type add_wargoal_panel = default_block_window {}`,
		`type bad_widget = default_block_window = {}`,
		`[BrokenParen(foo))]`,
		`@third = @[1/3]`,
		`value = @[-(@base + |$DELTA|2$|)]`,
		`value = $VALUE|100$`,
		`$STATE$ = { owner = $OWNER$ }`,
		"value = $NAME\ngood = yes",
		`value = $NAME|$`,
		`foo_$TYPE$ = "$PREFIX$_name"`,
		`has_global_flag = crisis_stage_$STAGE|1$`,
		`quoted_math = "@\[ $COUNT$ * 500]"`,
		`@example = @other`,
		`@@>`,
		`0$|#`,
		"value = prefix_$NAME\ngood = yes",
		"value = \"prefix_$NAME\"\ngood = yes",
		`$%<0`,
		`[[enabled] value = yes]`,
		`[[outer] [[inner] value] ]`,
		`foo =`,
		`broken = "unterminated`,
	}
	for _, seed := range seeds {
		f.Add(seed)
	}
	corpus, err := os.ReadFile("testdata/syntax-exploration.txt")
	if err != nil {
		f.Fatal(err)
	}
	f.Add(string(corpus))
	f.Fuzz(func(t *testing.T, source string) {
		root := Parse(source)
		if got := root.Text(); got != source {
			t.Fatalf("root text:\n%q\nwant:\n%q", got, source)
		}
	})
}
