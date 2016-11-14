// Package full implements the full white-box AES construction, with decomposed S-boxes. An attack on this construction
// is not implemented.
//
// http://dl.acm.org/citation.cfm?id=2995314
package full

import (
	"github.com/OpenWhiteBox/primitives/encoding"
	"github.com/OpenWhiteBox/primitives/matrix"
)

// blockAffine is a modification of encoding.BlockAffine that allows non-bijective transformations.
type blockAffine struct {
	linear   matrix.Matrix
	constant matrix.Row
}

func parseBlockAffine(in []byte) (*blockAffine, []byte) {
	out := &blockAffine{linear: matrix.Matrix{}}

	h, w := 8*int(in[0]), int(in[1])
	in = in[2:]
	for i := 0; i < h; i++ {
		out.linear = append(out.linear, matrix.Row(in[:w]))
		in = in[w:]
	}
	out.constant = matrix.Row(in[:(h / 8)])
	in = in[(h / 8):]

	return out, in
}

func (ba *blockAffine) compose(in *blockAffine) *blockAffine {
	return &blockAffine{
		linear:   ba.linear.Compose(in.linear),
		constant: ba.linear.Mul(in.constant).Add(ba.constant),
	}
}

func (ba *blockAffine) transform(in []byte) []byte {
	temp := ba.linear.Mul(matrix.Row(in))
	temp = temp.Add(ba.constant)

	return []byte(temp)
}

func (ba *blockAffine) BlockAffine() encoding.BlockAffine {
	out := encoding.BlockAffine{
		BlockLinear: encoding.NewBlockLinear(ba.linear),
	}
	copy(out.BlockAdditive[:], ba.constant)

	return out
}

func (ba *blockAffine) serialize(out *[]byte) {
	h, w := ba.linear.Size()
	*out = append(*out, byte(h/8), byte(w/8))

	for _, row := range ba.linear {
		*out = append(*out, row...)
	}
	*out = append(*out, ba.constant...)
}

// compress compute the AND of neighboring bits in src and stores the result in dst.
func compress(dst, src []byte) {
	for i := 0; i < 8*len(dst); i++ {
		b1 := src[(2*i+0)/8] >> uint((2*i+0)%8)
		b2 := src[(2*i+1)/8] >> uint((2*i+1)%8)

		dst[i/8] += (b1 & b2 & 1) << uint(i%8)
	}
}

type Construction [41]*blockAffine

// BlockSize returns the block size of AES. (Necessary to implement cipher.Block.)
func (constr Construction) BlockSize() int { return 16 }

// Encrypt encrypts the first block in src into dst. Dst and src may point at the same memory.
func (constr Construction) Encrypt(dst, src []byte) {
	state := src[:16]

	for i, m := range constr[:len(constr)-1] {
		temp := m.transform(state)
		state = make([]byte, stateSize[i%4])

		cs := compressSize[i%4]
		compress(state[:cs], temp[:2*cs])
		copy(state[cs:], temp[2*cs:])
	}

	state = constr[40].transform(state)
	copy(dst[:16], state[:16])
}

// Decrypt is not implemented.
func (constr Construction) Decrypt(_, _ []byte) {}
