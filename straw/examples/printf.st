Printf: λ [format string] (args slice[interface{}]) → {
	∀ c ∈ format → match c (
		'%': .print_number %
		_: .print c
	)
}
# WIP
