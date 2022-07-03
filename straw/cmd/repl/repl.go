package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/yjp20/turtle/straw/pkg/ast"
	"github.com/yjp20/turtle/straw/pkg/generator"
	"github.com/yjp20/turtle/straw/pkg/parser"
	"github.com/yjp20/turtle/straw/pkg/token"
	"github.com/yjp20/turtle/straw/pkg/vm"
)

var PROMPT = ">>> "

func main() {
	errors := token.NewErrorList()

	scn := bufio.NewScanner(os.Stdin)
	env := vm.NewVM(&errors)
	frame := vm.NewFunctionFrame(nil)

	for {
		fmt.Fprintf(os.Stdout, PROMPT)
		scanned := scn.Scan()
		if !scanned {
			return
		}

		line := scn.Bytes()
		file := token.NewFile(line)
		lex := parser.NewLexer(file, &errors)
		par := parser.NewParser(lex, &errors)
		gen := generator.NewGenerator(&errors)

		node := par.ParseProgram()
		code := gen.Generate(node)

		if len(errors) != 0 {
			errors.Print()
			ast.Print(node)
			continue
		}

		res := env.Eval(code, frame)
		println(res.Inspect())
	}
}
