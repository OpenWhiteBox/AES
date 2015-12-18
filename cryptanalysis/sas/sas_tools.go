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
	for i, v_i := range v[0:256] {
		out.DecKey[i] = byte(v_i) // EncKey and DecKey are backwards because we recover the inverse of the S-box.
	}

	for i, j := range out.EncKey {
		out.EncKey[j] = byte(i)
	}

	return
}

func xorArray(a, b [16]byte) (out [16]byte) {
	for i := 0; i < 16; i++ {
		out[i] = a[i] ^ b[i]
	}

	return
}

func EncryptAtPosition(constr sas.Construction, pos int, val byte) (out [16]byte) {
	in := [16]byte{}
	in[pos] = val

	constr.Encrypt(out[:], in[:])

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

		if v[:256].IsPermutation() {
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
