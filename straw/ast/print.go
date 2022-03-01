package ast

import "fmt"
import "reflect"

func Print(n Node) {
	printer(n, "")
}

var NODE = reflect.TypeOf((*Node)(nil)).Elem()

func printer(n Node, indent string) {
	if n == nil || reflect.ValueOf(n).IsNil() {
		println("nil")
		return
	}

	v := reflect.ValueOf(n).Elem()
	t := reflect.TypeOf(n).Elem()

	fmt.Printf("!%s\n", t.Name())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)
		if f.Type.Kind() == reflect.Slice {
			for j := 0; j < fv.Len(); j++ {
				fmt.Printf("%s| %s[%d]: ", indent, f.Name, j)
				printer(fv.Index(j).Interface().(Node), indent+"| ")
			}
		} else if fv.Type().Implements(NODE) {
			fmt.Printf("%s| %s: ", indent, f.Name)
			if n, ok := fv.Interface().(Node); ok {
				printer(n, indent+"| ")
			} else {
				printer(nil, indent+"| ")
			}
		} else {
			switch e := fv.Interface().(type) {
			case string:
				fmt.Printf("%s| %s: %s\n", indent, f.Name, e)
			case int64:
				fmt.Printf("%s| %s: %d\n", indent, f.Name, e)
			}
		}
	}
}
