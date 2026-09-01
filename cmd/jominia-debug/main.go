package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/arxbombus/jominia/internal/script/parser"
	"github.com/arxbombus/jominia/internal/script/syntax"
	"github.com/arxbombus/jominia/internal/tree"
)

func main() {
	outputPath := flag.String("o", "", "write output to file")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: jominia-debug [-o output] <file>")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	inputPath := flag.Arg(0)

	// #nosec G703
	source, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", inputPath, err)
		os.Exit(1)
	}
	root := parser.Parse(string(source))

	output := os.Stdout
	if *outputPath != "" {
		// #nosec G703
		output, err = os.Create(*outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create %s: %v\n", *outputPath, err)
			os.Exit(1)
		}
	}

	dumpGreenTree(output, root, 0)

	if output != os.Stdout {
		if err := output.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "close %s: %v\n", *outputPath, err)
			os.Exit(1)
		}
	}
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
