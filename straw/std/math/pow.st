Pow: λ (a i64, b i64) → {
	product: 1
	∀ _ ∈ range[0‥b) → {
		product: product * a
	}
	return product
}
