# fibo: λ (n i64) → {
# 	a: .make array[i64] {n+1}
# 	a[0]: 0
# 	a[1]: 1
# 	∀ i ∈ range[2‥n] → {
# 		a[i]: a[i-2] + a[i-1]
# 	}
# 	return a[n]
# }
# .fibo 20
# <i64 6765>
