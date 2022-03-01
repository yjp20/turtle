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
	p := parser.NewParser(straw.Filter(bytes))
	pg := p.ParseProgram()
	ast.Print(pg)
	env := interpreter.NewGlobalFrame()
	eval := interpreter.Eval(pg, env)
	println(eval.Inspect())
}
