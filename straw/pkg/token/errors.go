package token

type ErrorList []error

func NewErrorList() ErrorList {
	return make([]error, 0)
}

func (el *ErrorList) Print() {
	for _, err := range *el {
		println(err.Error())
	}
}
