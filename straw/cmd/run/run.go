package main

import (
	"io/ioutil"
	"os"

	"github.com/yjp20/turtle/straw"
	"github.com/yjp20/turtle/straw/ast"
	"github.com/yjp20/turtle/straw/interpreter"
	"github.com/yjp20/turtle/straw/parser"
	"github.com/yjp20/turtle/straw/token"
)

func main() {
	bytes, _ := ioutil.ReadAll(os.Stdin)

	errors := token.NewErrorList()
	file := parser.NewFile(bytes)
	ps := parser.NewParser(file, &errors)
	at := ps.ParseProgram()

	if len(errors) != 0 {
		errors.Print()
		ast.Print(at)
		return
	}

	env := interpreter.NewGlobalFrame(&errors)
	frame := interpreter.NewFunctionFrame(env)
	eval := interpreter.Eval(at, frame)
	println(eval.Inspect())
}
