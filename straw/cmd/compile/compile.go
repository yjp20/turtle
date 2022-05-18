package main

import (
	"io/ioutil"
	"os"

	"github.com/yjp20/turtle/straw/ast"
	"github.com/yjp20/turtle/straw/compiler"
	"github.com/yjp20/turtle/straw/parser"
	"github.com/yjp20/turtle/straw/token"
)

func main() {
	b, _ := ioutil.ReadAll(os.Stdin)

	errors := token.NewErrorList()
	file := parser.NewFile(b)
	ps := parser.NewParser(file, &errors)
	gn := parser.NewGenerator()
	at := ps.ParseProgram()
	ast.Print(at)
	gn.Generate(at)
	gn.Print()

	println(compiler.Compile(gn.Instructions))
}
