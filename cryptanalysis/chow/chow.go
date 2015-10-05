package chow

import (
	"github.com/OpenWhiteBox/AES/constructions/chow"
	"github.com/OpenWhiteBox/AES/primitives/table"
)

type InvertibleTable table.Byte

func Invert(it InvertibleTable) InvertibleTable {
	out := make([]byte, 256)

	for i := 0; i < 256; i++ {
		out[it.Get(byte(i))] = byte(i)
	}

	return InvertibleTable(table.ParsedByte(out))
}

// A new lookup table mapping an input position to an output position with other values in the column held constant.
type F struct {
	Constr          chow.Construction
	Round, Position int
	Base            byte
}

func (f F) Get(i byte) byte {
	pos := f.Position / 4 * 4

	block := make([]byte, 4)
	block[0], block[1] = i, f.Base

	stretched := f.Constr.ExpandWord(f.Constr.TBoxTyiTable[f.Round][pos:pos+4], block)
	copy(block, f.Constr.SquashWords(f.Constr.HighXORTable[f.Round][2*pos:2*pos+8], stretched))

	stretched = f.Constr.ExpandWord(f.Constr.MBInverseTable[f.Round][pos:pos+4], block)
	copy(block, f.Constr.SquashWords(f.Constr.LowXORTable[f.Round][2*pos:2*pos+8], stretched))

	return block[f.Position%4]
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

func RecoverKey(constr chow.Construction) (key [16]byte) {
	// The set of elements S, of the form fXX(f00^(-1)(x)) = Q(Q^(-1)(x) + b) for indeterminate x is isomorphic to the
	// additive group (GF(2)^8, xor) under composition.

	// Find 8 elements of S that act as a basis for S and build isomorphism psi.

	// Build Qtilde, with the identity g(00) = Qtilde(psi(g)), foreach g in S.

	return
}

// The set of elements S, of the form fXX(f00^(-1)(x)) = Q(Q^(-1)(x) + b) for indeterminate x is isomorphic to the
// additive group (GF(2)^8, xor) under composition.
func GenerateS(constr chow.Construction) [][256]byte {
	f00 := InvertibleTable(F{constr, 0, 0, 0x00})
	f00Inv := Invert(f00)

	S := make([][256]byte, 256)
	for x := 0; x < 256; x++ {
		copy(S[x][:], table.SerializeByte(table.ComposedBytes{
			F{constr, 0, 0, byte(x)},
			f00Inv,
		}))
	}

	return S
}

// Find 8 elements of S that act as a basis for S and build isomorphism psi.
func FindBasisAndSort(S [][256]byte) (basis []table.Byte) {
	for len(basis) < 8 { // Until we have a full basis.
		basis = append(basis, table.ParsedByte(S[1<<uint(len(basis))][:])) // Add the first independent vector to the basis.

		// Move all (now) dependent vectors from S into their correct position.
		for i := 1 << uint(len(basis)-1); i < 1<<uint(len(basis)); i++ {
			vect := [256]byte{}
			copy(vect[:], table.SerializeByte(FunctionFromBasis(i, basis)))

			// Move it to the correct position in S.
			for j := i + 1; j < len(S); j++ {
				if vect == S[j] {
					S[i], S[j] = S[j], S[i]
					break
				}
			}
		}
	}

	return
}
