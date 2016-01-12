// Cryptanalysis of SAS block ciphers.
package sas

import (
	"github.com/OpenWhiteBox/AES/primitives/encoding"
	binmatrix "github.com/OpenWhiteBox/AES/primitives/matrix"
	matrix "github.com/OpenWhiteBox/AES/primitives/gfmatrix"
	"github.com/OpenWhiteBox/AES/primitives/number"
)

type Construction interface {
	Encrypt([]byte, []byte)
}

func DecomposeSAS(constr Construction) (encoding.Block, binmatrix.Matrix, [16]byte, encoding.Block) {
	outer := encoding.ConcatenatedBlock(RecoverLastSBoxes(constr))
	inner := encoding.ConcatenatedBlock(RecoverFirstSBoxes(constr, outer))

	linear, constant := RecoverAffine(constr, inner, outer)

	return inner, linear, constant, outer
}

func RecoverAffine(constr Construction, inner, outer encoding.Block) (binmatrix.Matrix, [16]byte) {
	// Find constant part of affine transformation.
	constant := [16]byte{}

	constant = inner.Decode(constant)
	constr.Encrypt(constant[:], constant[:])
	constant = outer.Decode(constant)

	// Find the linear part one column at a time.
	linear := binmatrix.Matrix([]binmatrix.Row{})

	for pos := 0; pos < 16; pos++ {
		for bitPos := 0; bitPos < 8; bitPos++ {
			col := [16]byte{}
			col[pos] = 0x01 << uint(bitPos)

			col = inner.Decode(col)
			constr.Encrypt(col[:], col[:])
			col = outer.Decode(col)

			col = xorArray(col, constant)

			linear = append(linear, col[:])
		}
	}

	return linear.Transpose(), constant
}

// RecoverFirstSBoxes recovers the input S-Boxes for the construction given the trailing S-Boxes.
func RecoverFirstSBoxes(constr Construction, outer encoding.Block) (out [16]encoding.Byte) {
	for i, _ := range out {
		out[i] = RecoverFirstSBox(constr, outer, i)
	}

	return
}

// RecoverFirstSBox recovers the input S-Box at position pos.
//
// constr: An SAS construction.
// outer:  The outer s-box of the construction.
// pos:    The position to recover the s-box from.
//
// Returns the S-Box as a byte encoding.
func RecoverFirstSBox(constr Construction, outer encoding.Block, pos int) encoding.Byte {
	balance := matrix.Matrix{}
	basis := []matrix.Row(matrix.GenerateIdentity(256))

	x := outer.Decode(EncryptAtPosition(constr, pos, 0x00))

	for c := 1; len(basis) > 9; c++ { // The size of the basis seems to asymptote around 9 vectors.
		for i, _ := range balance { // Add a new position for the new constant to each row above us.
			balance[i] = append(balance[i], 0x00)
		}

		y := outer.Decode(EncryptAtPosition(constr, pos, byte(c))) // Our constant will be S(0x00) ^ S(c).
		target := xorArray(x, y)

		rows := GenerateInnerBalance(constr, outer, pos, target) // Finds pairs of inputs s.t. S(x) ^ S(y) = S(0) ^ S(c).
		for _, row := range rows {
			// Because we need to collect equations for several constants to get a small enough nullspace, the index variable
			// is 0x01 at position i if the preceeding relation applies to constant i.
			index := make([]number.ByteFieldElem, c)
			index[c-1] = 0x01

			balance = append(balance, append(row, index...))
		}

		basis = balance.NullSpace() // Calculate the new nullspace to see if we've met our mark.
	}

	return NewSBox(FindPermutation(basis), false)
}

// GenerateInnerBalance finds pairs of inputs x,y such that E(x) + E(y) = target by toggling the (pos)th position.
//
// constr: An SAS construction.
// outer:  The outer s-box of the construction.
// pos:    The position in the plaintexts to toggle.
// target: The target ciphertext.
//
// Returns an array of rows where the ith and jth positions are one iff:
//   x[pos] = i, y[pos] = j   =>   E(x) + E(y) = target
func GenerateInnerBalance(constr Construction, outer encoding.Block, pos int, target [16]byte) (out []matrix.Row) {
	for i := 0; i < 255; i++ { // 255 rather than 256 because the for loop below will be degenerate if i = 255.
		x := outer.Decode(EncryptAtPosition(constr, pos, byte(i)))

		// Skip this if we've already found it.
		found := false
		for _, row := range out {
			if !row[i].IsZero() {
				found = true
				break
			}
		}

		if found {
			continue
		}

		for j := i + 1; j < 256; j++ {
			y := outer.Decode(EncryptAtPosition(constr, pos, byte(j)))

			if xorArray(x, y) == target {
				row := matrix.Row(make([]number.ByteFieldElem, 256))
				row[i], row[j] = 0x01, 0x01

				out = append(out, row)
				break
			}
		}
	}

	return
}

// RecoverLastSBoxes takes an SAS block cipher as input and returns the trailing S-boxes of each position.
func RecoverLastSBoxes(constr Construction) (out [16]encoding.Byte) {
	ms := GenerateOuterBalance(constr)

	for i, m := range ms {
		out[i] = NewSBox(FindPermutation(m.NullSpace()), true)
	}

	return
}

func unfinished(pointers [16]int) bool {
	for _, x := range pointers {
		if x != 256 {
			return true
		}
	}

	return false
}

// GenerateOuterBalance takes an SAS block cipher as input and finds a balance matrix for the trailing S-Box of each
// position.
//
// A balance matrix is a row of balances specifying an S-Box.
// The dot product of each row of the balance matrix with [ S'(0) S'(1) ... S'(255) ] equals zero, where S' is the
// inverse of the S-Box.
func GenerateOuterBalance(constr Construction) (out [16]matrix.Matrix) {
	// Set defaults for out.
	for i, _ := range out {
		out[i] = matrix.GenerateEmpty(256)
	}

	pointers := [16]int{} // Represents how far we've gotten filling the balance matrix for each S-Box.
	for unfinished(pointers) {
		pts := GenerateRandomPlaintexts(0)
		ct := make([]byte, 16)

		for _, pt := range pts {
			constr.Encrypt(ct, pt) // C[..]PC[..] -(S)-> C[..]PC[..] -(A)->   D[..]   -(S)-> D[..] (See: Biryukov's Calculus)
			constr.Encrypt(ct, ct) //    D[..]    -(S)->    D[..]    -(A)->   B[..]   -(S)-> x[..]

			// Accumulate the linear relationships of all the ciphertexts in the next empty row of the matrix.
			for i, pointer := range pointers {
				if pointer < 256 {
					out[i][pointer][ct[i]] = out[i][pointer][ct[i]].Add(0x01)
				}
			}
		}

		for i, pointer := range pointers { // Advance the matrix's pointer if the new row isn't a duplicate.
			duplicate := false

			for j := 0; j < pointer; j++ {
				if out[i][pointer].Equal(out[i][j]) {
					duplicate = true
					break
				}
			}

			if duplicate { // If the row we built was a duplicate, clear it out.
				out[i][pointer] = matrix.Row(make([]number.ByteFieldElem, 256))
			} else { // If it wasn't, advance the position.
				pointers[i]++
			}
		}
	}

	return
}
