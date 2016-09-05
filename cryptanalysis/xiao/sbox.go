package xiao

import (
	"github.com/OpenWhiteBox/primitives/encoding"
	"github.com/OpenWhiteBox/primitives/equivalence"
	"github.com/OpenWhiteBox/primitives/matrix"

	"github.com/OpenWhiteBox/AES/constructions/saes"
)

// sbox is a Byte encoding of AES's "standard" S-box.
type sbox struct{}

func (sbox sbox) Encode(in byte) byte {
	constr := saes.Construction{}
	return constr.SubByte(in)
}

func (sbox sbox) Decode(in byte) byte {
	constr := saes.Construction{}
	return constr.UnSubByte(in)
}

// sboxLayer implements methods for disambiguating an S-box layer of the SPN.
type sboxLayer encoding.ConcatenatedBlock

func (sbl sboxLayer) Encode(in [16]byte) [16]byte {
	return encoding.ConcatenatedBlock(sbl).Encode(in)
}

func (sbl sboxLayer) Decode(in [16]byte) [16]byte {
	return encoding.ConcatenatedBlock(sbl).Decode(in)
}

// permuteBy permutes the S-boxes according to the 128-by-128 permutation matrix perm. If forwards = true, the forwards
// permutation is used; else, the backwards permutation is used.
func (sbl *sboxLayer) permuteBy(perm matrix.Matrix, forwards bool) {
	temp := encoding.ConcatenatedBlock{}
	for row := 0; row < 16; row++ {
		col := perm[8*row].Height() / 8

		if forwards {
			temp[row] = (*sbl)[col]
		} else {
			temp[col] = (*sbl)[row]
		}
	}

	*sbl = sboxLayer(temp)
}

// whiten puts an xor-mask on the input to each S-box so that S(0x00) = 0x00.
func (sbl *sboxLayer) whiten() (mask [16]byte) {
	for pos := 0; pos < 16; pos++ {
		m := (*sbl)[pos].Decode(0x00)

		mask[pos] = m
		(*sbl)[pos] = encoding.ComposedBytes{
			encoding.ByteAdditive(m),
			(*sbl)[pos],
		}
	}

	return
}

// cleanLinear finds the linear error on the input and output of each middle S-box. It removes it from the S-box and
// returns it. After this function is applied, all S-boxes will be equal to the whitened standard S-box, Sbar.
func (sbl *sboxLayer) cleanLinear() (in, out encoding.ConcatenatedBlock) {
	Sbar := encoding.ComposedBytes{encoding.ByteAdditive(0x52), sbox{}}

	for pos := 0; pos < 16; pos++ {
		eqs := equivalence.FindLinear((*sbl)[pos], Sbar, 1)

		(*sbl)[pos] = encoding.ComposedBytes{eqs[0].A, (*sbl)[pos], encoding.InverseByte{eqs[0].B}}
		in[pos], out[pos] = encoding.InverseByte{eqs[0].A}, eqs[0].B
	}

	return
}
