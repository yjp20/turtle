fibo: λ (n i32) → match n {
	0 ⇒ 0
	1 ⇒ 1
	_ ⇒ .fibo {n-1} + .fibo {n-2}
}

.fibo 20