package main

import (
	"io/ioutil"
	"os"

	"github.com/yjp20/turtle/straw/ast"
	"github.com/yjp20/turtle/straw/compiler"
	"github.com/yjp20/turtle/straw/parser"
)

func main() {
	b, _ := ioutil.ReadAll(os.Stdin)
	file := parser.NewFile(b)
	errors := make([]error, 0)
	p := parser.NewParser(file, &errors)
	g := parser.NewGenerator()
	a := p.ParseProgram()
	ast.Print(a)
	g.Generate(a)
	g.Print()

	println(compiler.Compile(g.Instructions))
}
