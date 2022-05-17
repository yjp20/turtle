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
	errors := make([]error, 0)

	scanner := bufio.NewScanner(os.Stdin)
	env := interpreter.NewGlobalFrame(&errors)
	frame := interpreter.NewFunctionFrame(env)

	for {
		fmt.Fprintf(os.Stdout, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Bytes()
		pf := parser.NewFile(straw.Filter(line))
		ps := parser.NewParser(pf, &errors)
		at := ps.ParseProgram()

		if len(errors) != 0 {
			for _, err := range errors {
				println(err.Error())
			}
			ast.Print(at)
			continue
		}

		eval := interpreter.Eval(at, frame)
		println(eval.Inspect())
	}
}
