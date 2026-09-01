package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/arxbombus/jominia/internal/jomini/parser"
	"github.com/arxbombus/jominia/internal/jomini/syntax"
	"github.com/arxbombus/jominia/internal/tree"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: jominia-debug <file>")
		os.Exit(2)
	}
	path := os.Args[1]
	// #nosec G703
	source, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err)
		os.Exit(1)
	}
	root := parser.Parse(string(source))
	dumpGreenTree(os.Stdout, root, 0)
}

func dumpGreenTree(output *os.File, element tree.GreenElement, depth int) {
	indent := strings.Repeat("  ", depth)
	switch element := element.(type) {
	case *tree.GreenNode:
		fmt.Fprintf(
			output,
			"%s%s\n",
			indent,
			syntax.FromRaw(element.Kind()),
		)
		for i := 0; i < element.ChildCount(); i++ {
			dumpGreenTree(output, element.Child(i), depth+1)
		}
	case *tree.GreenToken:
		fmt.Fprintf(
			output,
			"%s%s %q\n",
			indent,
			syntax.FromRaw(element.Kind()),
			element.TextTrimmed(),
		)
	}
}
