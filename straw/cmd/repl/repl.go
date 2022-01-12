package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/yjp20/turtle/straw"
)

var PROMPT = ">>> "

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	env := straw.NewFrame(nil)

	for {
		fmt.Fprintf(os.Stdout, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Bytes()
		parser := straw.NewParser(line)
		program := parser.ParseProgram()
		eval := straw.Eval(program, env)
		debug(program)
		println(eval.Inspect())
	}
}

func debug(something interface{}) {
	res, _ := json.MarshalIndent(something, "", "| ")
	println(string(res))
}
