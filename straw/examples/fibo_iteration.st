fibo: λ (n i64) → {
	n = 0 ⇒ return 0 ~ 0
	a: 0
	b: 1
	∀ i ∈ range[0‥n) → {
		t: a+b
		a: b
		b: t
	}
	return a
}
.fibo 20
# <i64 6765>
