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
	file := parser.NewFile([]byte(filtered))
	ps := parser.NewParser(file, &errors)
	at := ps.ParseProgram()

	if len(errors) != 0 {
		for _, err := range errors {
			println(err.Error())
		}
		ast.Print(at)
		return
	}

	env := interpreter.NewGlobalFrame(&errors)
	frame := interpreter.NewFunctionFrame(env)
	eval := interpreter.Eval(at, frame)
	println(eval.Inspect())
}
