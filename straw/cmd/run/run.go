package main

import (
	"io/ioutil"
	"os"

	"github.com/yjp20/turtle/straw/pkg/ast"
	"github.com/yjp20/turtle/straw/pkg/generator"
	"github.com/yjp20/turtle/straw/pkg/parser"
	"github.com/yjp20/turtle/straw/pkg/token"
	"github.com/yjp20/turtle/straw/pkg/vm"
)

func main() {
	bytes, _ := ioutil.ReadAll(os.Stdin)

	errors := token.NewErrorList()

	file := token.NewFile(bytes)
	lex := parser.NewLexer(file, &errors)
	par := parser.NewParser(lex, &errors)
	gen := generator.NewGenerator(&errors)

	node := par.ParseProgram()
	code := gen.Generate(node)
	println(code.String())

	if len(errors) != 0 {
		errors.Print()
		ast.Print(node)
		return
	}

	eval := vm.Eval(code, &errors, nil)
	println(eval.String())
}
