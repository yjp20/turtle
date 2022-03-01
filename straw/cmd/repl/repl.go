package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/yjp20/turtle/straw"
	"github.com/yjp20/turtle/straw/ast"
	"github.com/yjp20/turtle/straw/interpreter"
	"github.com/yjp20/turtle/straw/parser"
)

var PROMPT = ">>> "

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	env := interpreter.NewGlobalFrame()

	for {
		fmt.Fprintf(os.Stdout, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Bytes()
		p := parser.NewParser(straw.Filter(line))
		pg := p.ParseProgram()
		ast.Print(pg)
		eval := interpreter.Eval(pg, env)
		println(eval.Inspect())
	}
}
