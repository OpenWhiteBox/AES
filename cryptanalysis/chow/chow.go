package chow

import (
	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/matrix"
	"github.com/OpenWhiteBox/AES/primitives/number"
	"github.com/OpenWhiteBox/AES/primitives/table"

	"github.com/OpenWhiteBox/AES/constructions/chow"
	"github.com/OpenWhiteBox/AES/constructions/common"
	"github.com/OpenWhiteBox/AES/constructions/saes"
)

// Element Characteristic -> Elements with that characteristic.
var CharToBeta = map[byte]byte{
	0x74: 0xf5,
	0xdb: 0x8d,
	0x34: 0xf6,
	0xef: 0x8f,
}

// A new lookup table mapping an input position to an output position with other values in the column held constant.
type F struct {
	Constr         *chow.Construction
	Round, In, Out int
	Base           byte
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

// RecoverKey returns the AES key used to generate the given white-box construction.
func RecoverKey(constr *chow.Construction) []byte {
	temp := make([]encoding.Byte, 16)

	// Recover all output affine encodings for round 1.
	for pos := 0; pos < 16; pos++ {
		temp[pos], _ = RecoverEncodings(constr, 1, pos)
	}

	// Recover all input affine encodings up to a key byte for round 2.
	// Compose it with the above's output.
	for col := 0; col < 4; col++ {
		_, in := RecoverEncodings(constr, 2, 4*col)

		for row := 0; row < 4; row++ {
			backPos := common.UnShiftRows(4*col + row)
			temp[backPos] = encoding.ComposedBytes{temp[backPos], in[row]}
		}
	}

	// Recover round key for round 2.
	// The output encoding of round 1 composed with the approximate input encoding of round 2 should be an affine
	// transformation with the identity matrix as the linear part and a key byte as the constant part.
	roundKey := make([]byte, 16)
	for pos := 0; pos < 16; pos++ {
		roundKey[pos] = temp[pos].Encode(0)
	}

	// Recover the master key from the round key and return.
	return BackOneRound(BackOneRound(roundKey, 2), 1)
}

// RecoverEncodings returns the full affine output encoding of affine-encoded f at the given position, as well as the
// input affine encodings for all neighboring bytes up to a key byte.  Returns (out, []in)
func RecoverEncodings(constr *chow.Construction, round, pos int) (encoding.ByteAffine, []encoding.ByteAffine) {
	Ps := make([]encoding.ByteAffine, 4)  // Approximate input encodings.
	Ds := make([]number.ByteFieldElem, 4) // Array of gamma * MC coefficient
	q := byte(0x00)                       // The constant part of the output encoding.

	L := RecoverL(constr, round, pos)
	Atilde := FindAtilde(constr, L)
	AtildeInv, _ := Atilde.Invert()

	for i := 0; i < 4; i++ {
		j := pos/4*4 + i

		inEnc, _ := RecoverAffineEncoded(
			constr, encoding.IdentityByte{}, round-1, common.UnShiftRows(j), common.UnShiftRows(j),
		)
		_, f := RecoverAffineEncoded(constr, inEnc, round, j, pos)

		var c byte
		Ds[i], c, Ps[i] = FindPartialEncoding(constr, f, L, AtildeInv)
		q ^= c

		if i == 0 {
			q ^= f.Get(0x00)
		}
	}

	DInv, _ := DecomposeAffineEncoding(encoding.ByteMultiplication(FindDuplicate(Ds).Invert()))
	A := Atilde.Compose(DInv)
	AInv, _ := A.Invert()

	return encoding.ByteAffine{encoding.ByteLinear{A, AInv}, q}, Ps
}

// FindPartialEncoding takes an affine encoded F and finds the values that strip its output encoding.  It returns the
// parameters it finds and the input encoding of f up to a key byte.
func FindPartialEncoding(constr *chow.Construction, f table.Byte, L, AtildeInv matrix.Matrix) (number.ByteFieldElem, byte, encoding.ByteAffine) {
	fInv := table.Invert(f)
	id := encoding.ByteLinear{matrix.GenerateIdentity(8), nil}

	SInv := table.InvertibleTable(common.InvTBox{saes.Construction{}, 0x00, 0x00})
	S := table.Invert(SInv)

	// Brute force the constant part of the output encoding and the beta in Atilde = A_i <- D(beta)
	for c := 0; c < 256; c++ {
		for d := 1; d < 256; d++ {
			cand := encoding.ComposedBytes{
				TableAsEncoding{f, fInv},
				encoding.ByteAffine{id, byte(c)},
				encoding.ByteLinear{AtildeInv, nil},
				encoding.ByteMultiplication(byte(d)), // D below
				TableAsEncoding{SInv, S},
			}

			if isAffine(cand) {
				a, b := DecomposeAffineEncoding(cand)
				return number.ByteFieldElem(d), byte(c), encoding.ByteAffine{encoding.ByteLinear{a, nil}, byte(b)}
			}
		}
	}

	panic("Failed to strip output encodings!")
}

// FindAtilde calculates a non-trivial matrix Atilde s.t. L <- Atilde = Atilde <- D(beta), where
// L = A_i <- D(beta) <- A_i^(-1)
func FindAtilde(constr *chow.Construction, L matrix.Matrix) matrix.Matrix {
	beta := CharToBeta[FindCharacteristic(L)]
	D, _ := DecomposeAffineEncoding(encoding.ByteMultiplication(beta))

	x := L.RightStretch().Add(D.LeftStretch()).NullSpace()

	m := matrix.Matrix(make([]matrix.Row, len(x)))
	for i, e := range x {
		m[i] = matrix.Row{e}
	}

	return m
}

// RecoverL recovers the matrix L = A_i <- D(beta) <- A_i^(-1) where A_i is the affine output mask at position i and
// D(beta) is the matrix of multiplication by beta in GF(2^8).
func RecoverL(constr *chow.Construction, round, pos int) matrix.Matrix {
	inPos, outPos := pos/4*4, pos/4*4+(pos+1)%4

	A := RecoverAffineRel(constr, round, inPos+0, outPos, pos)
	B := RecoverAffineRel(constr, round, inPos+1, pos, outPos)

	LEnc := encoding.ComposedBytes{A, B}
	L, _ := DecomposeAffineEncoding(LEnc)

	return L
}

// RecoverAffineRel returns the affine relationship that maps y_i to y_j (instances of F with affine output encodings),
// both taking input in the (inPos)th position and outputting in the (outPos1)th and (outPos2)th position, respectively.
func RecoverAffineRel(constr *chow.Construction, round, inPos, outPos1, outPos2 int) encoding.ByteAffine {
	_, y_i := RecoverAffineEncoded(constr, encoding.IdentityByte{}, round, inPos, outPos1)
	_, y_j := RecoverAffineEncoded(constr, encoding.IdentityByte{}, round, inPos, outPos2)

	RelEnc := encoding.ComposedBytes{
		TableAsEncoding{table.Invert(y_i), nil},
		TableAsEncoding{y_j, nil},
	}

	L, c := DecomposeAffineEncoding(RelEnc)
	return encoding.ByteAffine{encoding.ByteLinear{L, nil}, c}
}

// RecoverAffineEncoded reduces the output encodings of a function to affine transformations.
func RecoverAffineEncoded(constr *chow.Construction, inputEnc encoding.Byte, round, inPos, outPos int) (encoding.Byte, table.InvertibleTable) {
	S := GenerateS(constr, round, inPos/4*4, outPos)
	_ = FindBasisAndSort(S)

	qtilde := Qtilde{S}

	outEnc := qtilde
	outTable := encoding.ByteTable{
		encoding.InverseByte{inputEnc},
		encoding.InverseByte{qtilde},
		F{constr, round, inPos, outPos, 0x00},
	}

	return outEnc, table.InvertibleTable(outTable)
}

// GenerateS creates the set of elements S, of the form fXX(f00^(-1)(x)) = Q(Q^(-1)(x) + b) for indeterminate x is
// isomorphic to the additive group (GF(2)^8, xor) under composition.
func GenerateS(constr *chow.Construction, round, inPos, outPos int) [][256]byte {
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
