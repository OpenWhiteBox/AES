// A number encoding is one specified by an element of GF(2^8).
package encoding

import (
	"github.com/OpenWhiteBox/AES/primitives/number"
)

type ByteMultiplication number.ByteFieldElem

func (bm ByteMultiplication) Encode(i byte) byte {
	x, j := number.ByteFieldElem(bm), number.ByteFieldElem(i)
	return byte(x.Mul(j))
}

func (bm ByteMultiplication) Decode(i byte) byte {
	x, j := number.ByteFieldElem(bm), number.ByteFieldElem(i)
	return byte(x.Invert().Mul(j))
}
