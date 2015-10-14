package chow

import (
	"github.com/OpenWhiteBox/AES/constructions/chow"
	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/table"
)

// A new lookup table mapping an input position to an output position with other values in the column held constant.
type F struct {
	Constr          chow.Construction
	Round, In, Out int
	Base            byte
}

func (f F) Get(i byte) byte {
	pos := f.Out / 4 * 4

	block := make([]byte, 4)
	block[f.In%4] = i
	block[(f.In+1)%4] = f.Base

	stretched := f.Constr.ExpandWord(f.Constr.TBoxTyiTable[f.Round][pos:pos+4], block)
	f.Constr.SquashWords(f.Constr.HighXORTable[f.Round][2*pos:2*pos+8], stretched, block)

	stretched = f.Constr.ExpandWord(f.Constr.MBInverseTable[f.Round][pos:pos+4], block)
	f.Constr.SquashWords(f.Constr.LowXORTable[f.Round][2*pos:2*pos+8], stretched, block)

	return block[f.Out%4]
}

// Qtilde is the approximation of the RoundEncoding between two rounds.
type Qtilde struct {
	S [][256]byte
}

func (q Qtilde) Encode(i byte) byte {
	return q.S[i][0]
}

func (q Qtilde) Decode(i byte) byte {
	for j, perm := range q.S {
		if perm[0] == i {
			return byte(j)
		}
	}

	return byte(0)
}

// MakeAffineRound reduces the output encodings of a round to affine transformations.
func MakeAffineRound(constr chow.Construction, inputEnc [16]encoding.Byte, round int) (outEnc [16]encoding.Byte, out [16]table.Byte) {
	for i, _ := range out {
		outEnc[shiftRows(i)], out[shiftRows(i)] = RecoverAffineEncoded(constr, inputEnc[i/4*4], round, i/4*4, i)
	}

	return
}

// RecoverAffineEncoded reduces the output encodings of a function to affine transformations.
func RecoverAffineEncoded(constr chow.Construction, inputEnc encoding.Byte, round, inPos, outPos int) (encoding.Byte, table.Byte) {
	S := GenerateS(constr, round, inPos, outPos)
	_ = FindBasisAndSort(S)

	qtilde := Qtilde{S}

	outEnc := qtilde
	outTable := encoding.ByteTable{
		encoding.InverseByte{inputEnc},
		encoding.InverseByte{qtilde},
		F{constr, round, inPos, outPos, 0x00},
	}

	return outEnc, outTable
}

// GenerateS creates the set of elements S, of the form fXX(f00^(-1)(x)) = Q(Q^(-1)(x) + b) for indeterminate x is
// isomorphic to the additive group (GF(2)^8, xor) under composition.
func GenerateS(constr chow.Construction, round, inPos, outPos int) [][256]byte {
	f00 := table.InvertibleTable(F{constr, round, inPos, outPos, 0x00})
	f00Inv := table.Invert(f00)

	S := make([][256]byte, 256)
	for x := 0; x < 256; x++ {
		copy(S[x][:], table.SerializeByte(table.ComposedBytes{
			f00Inv,
			F{constr, round, inPos, outPos, byte(x)},
		}))
	}

	return S
}

// FindBasisAndSort finds 8 elements of S that act as a basis for S and build isomorphism psi.
func FindBasisAndSort(S [][256]byte) (basis []table.Byte) {
	for len(basis) < 8 { // Until we have a full basis.
		basis = append(basis, table.ParsedByte(S[1<<uint(len(basis))][:])) // Add the first independent vector to the basis.

		// Move all (now) dependent vectors from S into their correct position.
		for i := 1 << uint(len(basis)-1); i < 1<<uint(len(basis)); i++ {
			vect := [256]byte{}
			copy(vect[:], table.SerializeByte(FunctionFromBasis(i, basis)))

			// Move it to the correct position in S.
			for j := i; j < len(S); j++ {
				if vect == S[j] {
					S[i], S[j] = S[j], S[i]
					break
				}
			}
		}
	}

	return
}
