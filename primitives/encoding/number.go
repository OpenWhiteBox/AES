package encoding

import (
	"github.com/OpenWhiteBox/AES/primitives/number"
)

// ByteMultiplication implements the Byte interface over multiplication by an element of GF(2^8).
type ByteMultiplication struct {
	// Forwards is the number to multiply by in the forwards (encoding) direction.
	Forwards number.ByteFieldElem
	// Backwards is the number to multiply by in the backwards (decoding) direction. It should be the inverse of Forwards.
	Backwards number.ByteFieldElem
}

func (bm ByteMultiplication) Encode(i byte) byte {
	x, j := number.ByteFieldElem(bm.Forwards), number.ByteFieldElem(i)
	return byte(x.Mul(j))
}

func (bm ByteMultiplication) Decode(i byte) byte {
	x, j := number.ByteFieldElem(bm.Backwards), number.ByteFieldElem(i)
	return byte(x.Mul(j))
}
