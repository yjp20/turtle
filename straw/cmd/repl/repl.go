package main

import (
	"bufio"
	"os"
	"fmt"
	"encoding/json"

	"github.com/yjp20/straw"
)

var PROMPT = ">>> "

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Fprintf(os.Stdout, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Bytes()
		parser := straw.NewParser(line)
		debug(parser.ParseProgram())
	}
}

func debug(something interface{}) {
	res, _ := json.MarshalIndent(something, "", "| ")
	println(string(res))
}
