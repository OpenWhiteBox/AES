package number

// ByteFieldElem is an element of Rijndael's field, GF(2^8)--i.e., a byte with field operations.
//
// The two operations implemented are addition and multiplication. The additive identity is 0x00 and all elements have
// themselves as additive inverses: x.Add(x) = 0x00 always. The multiplicative identity is 0x01 and all non-zero
// elements have a multiplicative inverse such that x.Mul(x.Invert()) = 0x01.
type ByteFieldElem uint16

var byteModulus ByteFieldElem = 0x11b

// Add returns e + f.
func (e ByteFieldElem) Add(f ByteFieldElem) ByteFieldElem {
	return e ^ f
}

// longMul returns e * f unreduced.
func (e ByteFieldElem) longMul(f ByteFieldElem) (out ByteFieldElem) {
	for i := uint(0); i < 16; i++ { // Foreach bit e_i in e:
		if e&(1<<i) != 0 { // where e_i equals 1:
			out = out ^ (f << i) // Add f * x^i to the output
		}
	}

	return
}

// longDiv returns the quotient and remainder of e / f (by Euclidean division), where e and f are considered elements of
// the ring of polynomials and not an element of the field.
func (numer ByteFieldElem) longDiv(denom ByteFieldElem) (ByteFieldElem, ByteFieldElem) {
	var (
		i, j     uint          = 15, 15
		quotient ByteFieldElem = 0
	)

	for ; i > 0; i-- { // i is the index of the higest bit in the denominator
		if denom&(1<<i) > 0 {
			break
		}
	}

	if numer > denom || numer&(1<<i) > 0 { // degree(numer) >= degree(denom)
		for ; 15 >= j && j >= i; j-- { // Foreach bit numer_j of numer, descending:
			if numer&(1<<j) != 0 { // where numer_j equals 1
				quotient += 1 << (j - i)           // Add x^(j - i) to the quotient
				numer = numer ^ (denom << (j - i)) // New remainder is numer - (denom * x^(j - i))
			}
		}

		return quotient, numer
	} else { // degree(numer) < degree(denom), so we can return zero quotient
		return 0, numer
	}
}

// Mul returns e * f.
func (e ByteFieldElem) Mul(f ByteFieldElem) (out ByteFieldElem) {
	var i, j uint // uints because they are used for bit shifts

	for i = 0; i < 8; i++ { // Foreach bit e_i in e:
		if e&(1<<i) != 0 { // where e_i equals 1:
			temp := f // Multiply f * x^i mod M(x):

			for j = 0; j < i; j++ { // Multiply f by x mod M(x), i times.
				temp = temp << 1

				if temp >= 0x100 {
					temp = temp ^ byteModulus
				}
			}

			out = out ^ temp // Add f * x^i to the output
		}
	}

	return
}

// Invert returns the multiplicative inverse of e, or 0x00 if e = 0x00.
func (e ByteFieldElem) Invert() ByteFieldElem { // Euclid's Extended Algorithm.  Computes inverse elements in GF(2^8).
	var r0, r1 ByteFieldElem = e.Dup(), byteModulus.Dup()
	var s0, s1 ByteFieldElem = 1, 0

	for r1 != 0 {
		q, _ := r0.longDiv(r1)

		r0, r1 = r1, r0.Add(q.longMul(r1))
		s0, s1 = s1, s0.Add(q.longMul(s1))
	}

	_, rem := s0.longDiv(byteModulus)
	return rem
}

// IsZero returns whether or not e is zero.
func (e ByteFieldElem) IsZero() bool { return e == 0 }

// IsOne returns whether ot not e is one.
func (e ByteFieldElem) IsOne() bool { return e == 1 }

// Dup returns a duplicate of e.
func (e ByteFieldElem) Dup() ByteFieldElem { return e.Add(0) }
