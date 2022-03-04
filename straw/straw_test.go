package straw

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/yjp20/turtle/straw/ast"
	"github.com/yjp20/turtle/straw/interpreter"
	"github.com/yjp20/turtle/straw/parser"
)

type Test struct {
	name string
	in   string
	out  interpreter.Object
}

var tests = []Test{
	{
		"order of operations",
		`1 + 2 * 3 + 4`,
		&interpreter.I64{Value: 11},
	},
	{
		"match",
		`k: 3
		match k {
			3 => 5
			2 => 5
			_ => 7
		}`,
		&interpreter.I64{Value: 5},
	},
	{
		"match default",
		`k: 10
		match k {
			3 => 5
			2 => 5
			_ => 7
		}`,
		&interpreter.I64{Value: 7},
	},
	{
		"if true",
		`k: 3
		k = 3 => 1 ~ 2`,
		&interpreter.I64{Value: 1},
	},
	{
		"if false",
		`k: 1
		k = 3 => 1 ~ 2`,
		&interpreter.I64{Value: 2},
	},
	{
		"if chain true false",
		`j: 1, k: 3
		j = 1 => 3 ~ k = 2 => 4 ~ 5`,
		&interpreter.I64{Value: 3},
	},
	{
		"if chain false true",
		`j: 0, k: 2
		j = 1 => 3 ~ k = 2 => 4 ~ 5`,
		&interpreter.I64{Value: 4},
	},
	{
		"if chain false false",
		`j: 0, k: 3
		j = 1 => 3 ~ k = 2 => 4 ~ 5`,
		&interpreter.I64{Value: 5},
	},
	{
		"tuple",
		`f: func (a i64, b i64) -> (a+b, a-b)
		(x, y): .f 8 2
		(y, x)`,
		&interpreter.Tuple{
			Fields: []interpreter.Field{
				{Name: "0", Type: &interpreter.Type{Kind: interpreter.TypeI64}, Value: &interpreter.I64{Value: 6}},
				{Name: "1", Type: &interpreter.Type{Kind: interpreter.TypeI64}, Value: &interpreter.I64{Value: 10}},
			},
		},
	},
	{
		"return 1",
		`f: λ (i i64) → {
			i = 10 ⇒ return 100
			return i
		}
		.f 10`,
		&interpreter.I64{Value: 100},
	},
	{
		"return 2",
		`f: λ (i i64) → {
			i = 10 ⇒ return 100
			return i
		}
		.f 1`,
		&interpreter.I64{Value: 1},
	},
	{
		"fibo recursive 20",
		`fibo: λ (n i64) → {
			match n {
				0 ⇒ 0
				1 ⇒ 1
				_ ⇒ .fibo { n - 1 } + .fibo { n - 2 }
			}
		}
		.fibo 20`,
		&interpreter.I64{Value: 6765},
	},
	{
		"fibo array 40",
		`fibo: λ (n i64) → {
			a: .make array[i64] { n+1 }
			a[0]: 0
			a[1]: 1
			∀ i ∈ range[2‥n] → {
				a[i]: a[i-2] + a[i-1]
			}
			a[n]
		}
		.fibo 40`,
		&interpreter.I64{Value: 102334155},
	},
}

func TestStraw(t *testing.T) {
	for _, test := range tests {
	e := []error{}
		c := Filter([]byte(test.in))
		p := parser.NewParser(c, e)
		t := p.ParseProgram()
		g := interpreter.NewGlobalFrame()
		f := interpreter.NewFunctionFrame(g)
		o := interpreter.Eval(t, f)
		if reflect.DeepEqual(o, test.out) {
			fmt.Printf("PASS %s\n", test.name)
		} else {
			fmt.Printf("FAIL %s\n", test.name)
			fmt.Println(strings.Replace(string(c), "\n\t\t", "\n", -1))
			ast.Print(t)
			if len(e) != 0 {
				fmt.Println(e)
			}
			fmt.Printf(" expected: %s\n got: %s\n", test.out.Inspect(), o.Inspect())
		}
	}
}
