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
		p := parser.NewParser(straw.Filter(line), &errors)
		pg := p.ParseProgram()

		if len(errors) != 0 {
			for _, err := range errors {
				println(err.Error())
			}
			ast.Print(pg)
			continue
		}

		eval := interpreter.Eval(pg, frame)
		println(eval.Inspect())
	}
}
