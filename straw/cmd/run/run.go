package main

import (
	"io/ioutil"
	"os"

	"github.com/yjp20/turtle/straw"
	"github.com/yjp20/turtle/straw/ast"
	"github.com/yjp20/turtle/straw/interpreter"
	"github.com/yjp20/turtle/straw/parser"
)

func main() {
	bytes, _ := ioutil.ReadAll(os.Stdin)
	filtered := straw.Filter(bytes)

	errors := make([]error, 0)
	p := parser.NewParser(filtered, errors)
	pg := p.ParseProgram()

	if len(errors) != 0 {
		for _, err := range errors {
			println(err.Error())
		}
		ast.Print(pg)
		return
	}

	env := interpreter.NewGlobalFrame()
	eval := interpreter.Eval(pg, env)
	println(eval.Inspect())
}
