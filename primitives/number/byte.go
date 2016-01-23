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
func (e ByteFieldElem) Invert() ByteFieldElem {
	out, temp := e.Dup(), e.Dup()

	for i := 0; i < 6; i++ {
		temp = temp.Mul(temp)
		out = out.Mul(temp)
	}

	return out.Mul(out)
}

// IsZero returns whether or not e is zero.
func (e ByteFieldElem) IsZero() bool { return e == 0 }

// IsOne returns whether ot not e is one.
func (e ByteFieldElem) IsOne() bool { return e == 1 }

// Dup returns a duplicate of e.
func (e ByteFieldElem) Dup() ByteFieldElem { return e.Add(0) }
