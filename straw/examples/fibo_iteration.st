fibo: λ (n i64) → {
	n = 0 ⇒ return 0
	(a,b): (0,1)
	∀ i ∈ range[1‥n) → {
		(a,b): (b,a+b)
	}
	return b
}

.fibo 40
