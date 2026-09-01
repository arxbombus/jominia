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
