// Package sas implements a cryptanalysis of generic SAS block ciphers. See constructions/sas for more information on
// the construction itself.
//
// It is based on Biryukov's multiset calculus.
//
// "Structural Cryptanalysis of SASAS" by Alex Biryukov and Adi Shamir,
// https://www.iacr.org/archive/eurocrypt2001/20450392.pdf
package sas

import (
	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/gfmatrix"

	"github.com/OpenWhiteBox/AES/constructions/sas"
)

// Construction represents a construction of an SAS block cipher. The implementation doesn't assume that this is a
// constructions/sas.Construction for generality, and the cryptanalysis doesn't assume that you have access to Encrypt
// AND Decrypt--access to either allows you to break it.
type Construction interface {
	Encrypt([]byte, []byte)
}

// Encoding implements encoding.Block over a Construction to make some code simpler. Decode can not be called.
type Encoding struct{ Construction }

func (e Encoding) Encode(in [16]byte) (out [16]byte) {
	e.Construction.Encrypt(out[:], in[:])
	return
}

func (e Encoding) Decode(in [16]byte) (out [16]byte) {
	panic("cryptanalysis/sas.Encoding.Decode should never be called!")
}

// DecomposeSAS takes a Construction as input and outputs a functionally identical constructions/sas.Construction, with
// which you can Encrypt, Decrypt, inspect internal constants, etc.
func DecomposeSAS(constr Construction) (out sas.Construction) {
	cipher := Encoding{constr}

	out.Last = RecoverLastSBoxes(cipher)
	out.First = RecoverFirstSBoxes(encoding.ComposedBlocks{cipher, encoding.InverseBlock{out.Last}})
	out.Affine = encoding.DecomposeBlockAffine(encoding.ComposedBlocks{
		encoding.InverseBlock{out.First},
		cipher,
		encoding.InverseBlock{out.Last},
	})

	return
}

// RecoverLastSBoxes takes an SAS block cipher as input and returns the trailing S-boxes.
func RecoverLastSBoxes(cipher encoding.Block) (out encoding.ConcatenatedBlock) {
	// It's advantageous to parallelize this step, so we get all the constraints at once.
	ms := LastSBoxConstraints(cipher)

	for i, m := range ms {
		out[i] = NewSBox(FindPermutation(m.NullSpace()), true)
	}

	return
}

// RecoverFirstSBoxes recovers the input S-Boxes for the construction given the trailing S-Boxes.
func RecoverFirstSBoxes(cipher encoding.Block) (out encoding.ConcatenatedBlock) {
	// There's no good way to parallelize this step, so we get the constraints and solve them one at a time.
	for i, _ := range out {
		out[i] = RecoverFirstSBox(cipher, i)
	}

	return
}

// RecoverFirstSBox takes as input an SAS construction with the trailing S-boxes stripped and returns the input S-Box at
// position pos.
func RecoverFirstSBox(cipher encoding.Block, pos int) encoding.Byte {
	im := gfmatrix.NewIncrementalMatrix(256)

	x := cipher.Encode(XatY(0x00, pos))
	target := [16]byte{}

	for c := 1; c < 256 && !SufficientlyDefined(im); c++ {
		y := cipher.Encode(XatY(byte(c), pos))
		encoding.XOR(target[:], x[:], y[:])

		rows := FirstSBoxConstraints(cipher, pos, target) // Finds pairs of inputs s.t. S(x) ^ S(y) = S(0) ^ S(c).

		for _, row := range rows {
			row[0], row[c] = row[0].Add(1), row[c].Add(1)
			im.Add(row)
		}
	}

	return NewSBox(FindPermutation(im.Matrix().NullSpace()), false)
}
