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
	frame := interpreter.NewFunctionFrame(env)

	for {
		fmt.Fprintf(os.Stdout, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		errors := make([]error, 0)
		line := scanner.Bytes()
		file := parser.NewFile(straw.Filter(line))
		sp := parser.NewParser(file, &errors)
		tree := sp.ParseProgram()

		if len(errors) != 0 {
			for _, err := range errors {
				println(err.Error())
			}
			ast.Print(tree)
			continue
		}

		eval := interpreter.Eval(tree, frame)
		println(eval.Inspect())
	}
}
