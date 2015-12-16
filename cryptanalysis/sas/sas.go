// Cryptanalysis of SAS block ciphers.
package sas

import (
	"crypto/rand"

	"github.com/OpenWhiteBox/AES/primitives/encoding"
	matrix "github.com/OpenWhiteBox/AES/primitives/matrix2"
	"github.com/OpenWhiteBox/AES/primitives/number"

	"github.com/OpenWhiteBox/AES/constructions/sas"
)

// NewSBox returns a new S-Box from a permutation vector.
func NewSBox(v matrix.Row) (out encoding.SBox) {
	for i, v_i := range v {
		out.DecKey[i] = byte(v_i) // EncKey and DecKey are backwards because we recover the inverse of the S-box.
	}

	for i, j := range out.EncKey {
		out.EncKey[j] = byte(i)
	}

	return
}

// RecoverLastSBoxes takes an SAS block cipher as input and returns the trailing S-boxes of each position.
func RecoverLastSBoxes(constr sas.Construction) (out [16]encoding.Byte) {
	ms := GenerateBalanceMatrices(constr.Encrypt)

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
func GenerateBalanceMatrices(f func([]byte, []byte)) (out [16]matrix.Matrix) {
	// Set defaults for out.
	for i, _ := range out {
		out[i] = matrix.GenerateEmpty(256)
	}

	pointers := [16]int{} // Represents how far we've gotten filling the balance matrix for each S-Box.
	for unfinished(pointers) {
		pts := GenerateRandomPlaintexts(0)
		ct := make([]byte, 16)

		for _, pt := range pts {
			f(ct, pt) // C[..]PC[..] -(S)-> C[..]PC[..] -(A)->   D[..]   -(S)-> D[..]
			f(ct, ct) //    D[..]    -(S)->    D[..]    -(A)->   B[..]   -(S)-> x[..]

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

// GenerateRandomPlaintexts returns a random multiset of C[..]PC[..] plaintexts with the P at the given position.
func GenerateRandomPlaintexts(pos int) (out [][]byte) {
	master := make([]byte, 16)
	rand.Read(master)

	for i := 0; i < 256; i++ {
		pt := make([]byte, 16)
		copy(pt, master)

		pt[pos] = byte(i)

		out = append(out, pt)
	}

	return
}

// FindPermutation takes a set of vectors and finds a linear combination of them that gives a permutation vector.
func FindPermutation(basis []matrix.Row) matrix.Row {
	for true {
		v := RandomLinearCombination(basis)

		if v.IsPermutation() {
			return v
		}
	}

	return nil
}

// RandomLinearCombination returns a random linear combination of a set of basis vectors.
func RandomLinearCombination(basis []matrix.Row) matrix.Row {
	coeffs := make([]byte, len(basis))
	rand.Read(coeffs)

	v := matrix.Row(make([]number.ByteFieldElem, basis[0].Size()))

	for i, c_i := range coeffs {
		v = v.Add(basis[i].ScalarMul(number.ByteFieldElem(c_i)))
	}

	return v
}
