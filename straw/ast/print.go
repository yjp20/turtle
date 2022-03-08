package ast

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

func Print(n Node) {
	printer(n, "", os.Stdout)
}

func Sprint(n Node) string {
	b := &strings.Builder{}
	printer(n, "", b)
	return b.String()
}

var NODE = reflect.TypeOf((*Node)(nil)).Elem()

func printer(n Node, indent string, w io.Writer) {
	if n == nil || reflect.ValueOf(n).IsNil() {
		fmt.Fprint(w, "nil\n")
		return
	}

	v := reflect.ValueOf(n).Elem()
	t := reflect.TypeOf(n).Elem()

	fmt.Fprintf(w, "!%s\n", t.Name())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)
		if f.Type.Kind() == reflect.Slice {
			for j := 0; j < fv.Len(); j++ {
				fmt.Fprintf(w, "%s| %s[%d]: ", indent, f.Name, j)
				printer(fv.Index(j).Interface().(Node), indent+"| ", w)
			}
		} else if fv.Type().Implements(NODE) {
			fmt.Fprintf(w, "%s| %s: ", indent, f.Name)
			if n, ok := fv.Interface().(Node); ok {
				printer(n, indent+"| ", w)
			} else {
				printer(nil, indent+"| ", w)
			}
		} else {
			switch e := fv.Interface().(type) {
			case string:
				fmt.Fprintf(w, "%s| %s: %s\n", indent, f.Name, e)
			case int64:
				fmt.Fprintf(w, "%s| %s: %d\n", indent, f.Name, e)
			case int32:
				fmt.Fprintf(w, "%s| %s: %d\n", indent, f.Name, e)
			case bool:
				fmt.Fprintf(w, "%s| %s: %t\n", indent, f.Name, e)
			}
		}
	}
}
