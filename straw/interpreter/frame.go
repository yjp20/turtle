package interpreter

type Frame interface {
	Get(selector string) Object
	Set(selector string, value Object)
}

type FunctionFrame struct {
	Parent Frame
	Values map[string]Object
	Return Object
}

func NewFunctionFrame(parent Frame) *FunctionFrame {
	return &FunctionFrame{
		Parent: parent,
		Values: make(map[string]Object),
	}
}

func (f *FunctionFrame) Get(selector string) Object {
	if _, ok := f.Values[selector]; ok {
		return f.Values[selector]
	}
	if f.Parent != nil {
		return f.Parent.Get(selector)
	}
	return NULL
}
func (f *FunctionFrame) Set(selector string, obj Object) {
	f.Values[selector] = obj
}

func NewGlobalFrame() *GlobalFrame {
	return &GlobalFrame{}
}

type GlobalFrame struct {
	Modules []string
}

func (f *GlobalFrame) Get(selector string) Object {
	switch selector {
	case "print":
		return &BuiltinFunction{Kind: "print"}
	case "debug":
		return &BuiltinFunction{Kind: "debug"}
	case "make":
		return &BuiltinFunction{Kind: "make"}
	case "import":
		return &BuiltinFunction{Kind: "import"}
	case "i32":
		return &Type{Kind: TypeI32}
	case "i64":
		return &Type{Kind: TypeI64}
	case "bool":
		return &Type{Kind: TypeBool}
	case "f64":
		return &Type{Kind: TypeF64}
	case "any":
		return &Type{Kind: TypeAny}
	case "array":
		return &Factory{
			Params: []Field{{Name: "T", Type: &Type{Kind: TypeType}}},
			Kind:   TypeArray,
		}
	case "slice":
		return &Factory{
			Params: []Field{{Name: "T", Type: &Type{Kind: TypeType}}},
			Kind:   TypeSlice,
		}
	}
	return NULL
}
func (f *GlobalFrame) Set(selector string, obj Object) {}
