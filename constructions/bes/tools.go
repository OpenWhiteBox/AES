package bes

import (
	"github.com/OpenWhiteBox/primitives/gfmatrix"
	"github.com/OpenWhiteBox/primitives/number"
)

// Expand takes a value from AES's space and embeds it in BES's space.
func Expand(in []byte) gfmatrix.Row {
	out := gfmatrix.NewRow(8 * len(in))

	for k, v := range in {
		num := number.ByteFieldElem(v)

		for pos := 0; pos < 8; pos++ {
			out[8*k+pos] = num
			num = num.Mul(num)
		}
	}

	return out
}

// Contract takes a value from BES's space and compacts it into AES's space.
func Contract(in gfmatrix.Row) []byte {
	out := make([]byte, len(in)/8)

	for pos := 0; pos < len(in); pos += 8 {
		out[pos/8] = byte(in[pos])
	}

	return out
}
