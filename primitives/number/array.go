// Polynomial fields with coefficients in GF(2^8)
package number

type ArrayFieldElem []ByteFieldElem

var arrayModulus ArrayFieldElem = ArrayFieldElem{
	ByteFieldElem(1), ByteFieldElem(0), ByteFieldElem(0), ByteFieldElem(0), ByteFieldElem(1),
}

func (e ArrayFieldElem) trim() ArrayFieldElem { // Trim preceeding zeros from polynomial.
	for i := len(e) - 1; i >= 0; i-- {
		if !e[i].IsZero() {
			return e[:i+1]
		}
	}

	return ArrayFieldElem{}
}

func (e ArrayFieldElem) Add(f ArrayFieldElem) ArrayFieldElem {
	out := ArrayFieldElem{}

	for i := 0; i < len(e) || i < len(f); i++ {
		out = append(out, ByteFieldElem(0))

		if i < len(e) {
			out[i] = out[i].Add(e[i])
		}

		if i < len(f) {
			out[i] = out[i].Add(f[i])
		}
	}

	return out.trim()
}

func (e ArrayFieldElem) longMul(f ArrayFieldElem) (out ArrayFieldElem) {
	for i, e_i := range e { // Foreach byte e_i in e:
		if !e_i.IsZero() { // with non-zero coefficient:
			// Add f * e_i * x^i to the output
			temp := f.ScalarMul(e_i)

			for j := 0; j < i; j++ {
				temp = append(ArrayFieldElem{0}, temp...)
			}

			out = out.Add(temp)
		}
	}

	return
}

// Euclidean division of two polynomials with coefficients in GF(2^8)
func (numer ArrayFieldElem) longDiv(denom ArrayFieldElem) (ArrayFieldElem, ArrayFieldElem) {
	if denom.IsZero() {
		return ArrayFieldElem{}, numer
	}

	quotient := ArrayFieldElem{}

	if len(numer) >= len(denom) { // degree(numer) >= degree(denom)
		for i := len(numer) - 1; len(numer) > i && i >= len(denom)-1; i-- { // Foreach byte numer_i of numer, descending:
			if !numer[i].IsZero() { // with non-zero coefficient, use f to cancel this coefficient
				r := ArrayFieldElem{numer[i]}
				for j := len(denom); j <= i; j++ {
					r = append(ArrayFieldElem{0}, r...)
				}

				quotient = quotient.Add(r)          // Add c * x^(i - n) to the quotient
				numer = numer.Add(denom.longMul(r)) // New remainder is numer - (denom * c * x^(i - n))
			}
		}

		return quotient, numer
	} else { // degree(numer) < degree(denom), so we can return zero quotient
		return ArrayFieldElem{}, numer
	}
}

func (e ArrayFieldElem) ScalarMul(g ByteFieldElem) (out ArrayFieldElem) {
	for _, e_i := range e {
		out = append(out, e_i.Mul(g))
	}

	return out.trim()
}

func (e ArrayFieldElem) Mul(f ArrayFieldElem) (out ArrayFieldElem) {
	for i, e_i := range e { // Foreach byte e_i in e:
		if !e_i.IsZero() { // with non-zero coefficient:
			temp := f.ScalarMul(e_i) // Multiply f * e_i * x^i mod M(x):

			for j := 0; j < i; j++ { // Multiply (f * e_i) by x mod M(x), i times.
				temp = append(ArrayFieldElem{0}, temp...)

				if len(temp) == len(arrayModulus) {
					temp = temp.Add(arrayModulus.ScalarMul(temp[len(temp)-1]))
				}
			}

			out = out.Add(temp) // Add f * e_i * x^i to the output
		}
	}

	return
}

// Invert() function for ArrayFieldElems has been omitted because I can't figure out why it doesn't work.

func (e ArrayFieldElem) IsZero() bool { return len(e) == 0 }
func (e ArrayFieldElem) IsOne() bool  { return len(e) == 1 && e[0] == 1 }

func (e ArrayFieldElem) Dup() ArrayFieldElem { return e.Add(ArrayFieldElem{}) }
