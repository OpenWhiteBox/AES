// Cryptanalysis of SAS block ciphers.
package sas

import (
	"github.com/OpenWhiteBox/AES/primitives/encoding"
	matrix "github.com/OpenWhiteBox/AES/primitives/matrix2"
	"github.com/OpenWhiteBox/AES/primitives/number"

	"github.com/OpenWhiteBox/AES/constructions/sas"
)

// RecoverFirstSBox recovers the input S-Box at position pos.
//
// constr: An SAS construction.
// outer:  The outer s-box of the construction.
// pos:    The position to recover the s-box from.
//
// Returns the S-Box as a byte encoding.
func RecoverFirstSBox(constr sas.Construction, outer encoding.Block, pos int) encoding.Byte {
	cap := 6
	// When the cap=1, the size of the basis of the nullspace is about 130 vectors. As cap increases, the basis asymptotes
	// around 10 vectors. cap=6 seems to be the most efficient tradeoff between time spent collecting vectors and time
	// spent searching the nullspace.

	balance := matrix.Matrix{}
	x := outer.Decode(EncryptAtPosition(constr, pos, 0x00))

	for c := 1; c <= cap; c++ {
		y := outer.Decode(EncryptAtPosition(constr, pos, byte(c)))
		target := xorArray(x, y)

		rows := FindTargetBalances(constr, outer, pos, target)
		for _, row := range rows {
			index := make([]number.ByteFieldElem, cap)
			index[c-1] = 0x01

			balance = append(balance, append(row, index...))
		}
	}

	return NewSBox(FindPermutation(balance.NullSpace()))
}

// FindTargetBalances finds pairs of inputs x,y such that E(x) + E(y) = target by toggling the (pos)th position.
//
// constr: An SAS construction.
// outer:  The outer s-box of the construction.
// pos:    The position in the plaintexts to toggle.
// target: The target ciphertext.
//
// Returns an array of rows where the ith and jth positions are one iff:
//   x[pos] = i, y[pos] = j   =>   E(x) + E(y) = target
func FindTargetBalances(constr sas.Construction, outer encoding.Block, pos int, target [16]byte) (out []matrix.Row) {
	for i := 1; i < 255; i++ { // 255 rather than 256 because the for loop below will be degenerate if i = 255.
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
func RecoverLastSBoxes(constr sas.Construction) (out [16]encoding.Byte) {
	ms := GenerateBalanceMatrices(constr)

	for i, m := range ms {
		out[i] = NewSBox(FindPermutation(m.NullSpace()))
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

// GenerateBalanceMatrices takes an SAS block cipher as input and finds a balance matrix for the trailing S-Box of each
// position.
//
// A balance matrix is...
func GenerateBalanceMatrices(constr sas.Construction) (out [16]matrix.Matrix) {
	// Set defaults for out.
	for i, _ := range out {
		out[i] = matrix.GenerateEmpty(256)
	}

	pointers := [16]int{} // Represents how far we've gotten filling the balance matrix for each S-Box.
	for unfinished(pointers) {
		pts := GenerateRandomPlaintexts(0)
		ct := make([]byte, 16)

		for _, pt := range pts {
			constr.Encrypt(ct, pt) // C[..]PC[..] -(S)-> C[..]PC[..] -(A)->   D[..]   -(S)-> D[..]
			constr.Encrypt(ct, ct) //    D[..]    -(S)->    D[..]    -(A)->   B[..]   -(S)-> x[..]

			// Accumulate the linear relationships of all the ciphertexts in the next empty row of the matrix.
			for i, pointer := range pointers {
				if pointer < 256 {
					out[i][pointer][ct[i]] = out[i][pointer][ct[i]].Add(0x01)
				}
			}
		}

		for i, pointer := range pointers {
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
