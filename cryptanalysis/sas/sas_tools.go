package sas

import (
	"crypto/rand"

	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/gfmatrix"
	"github.com/OpenWhiteBox/AES/primitives/number"
)

// SufficientlyDefined returns true if the incremental matrix has a 9-dimensional nullspace or smaller. This way, it
// is small enough to search, but not so small that we have nowhere to look for solutions.
func SufficientlyDefined(im gfmatrix.IncrementalMatrix) bool {
	return im.Len() >= 247
}

// IncrementalMatrices is implements succint operations over a slice of incremental matrices.
type IncrementalMatrices []gfmatrix.IncrementalMatrix

// NewIncrementalMatrices returns a new slice of x n-by-n incremental matrices.
func NewIncrementalMatrices(x, n int) (ims IncrementalMatrices) {
	ims = make([]gfmatrix.IncrementalMatrix, x)
	for i, _ := range ims {
		ims[i] = gfmatrix.NewIncrementalMatrix(n)
	}

	return
}

// SufficientlyDefined returns true if every incremental matrix is sufficiently defined.
func (ims IncrementalMatrices) SufficientlyDefined() bool {
	for _, im := range ims {
		if !SufficientlyDefined(im) {
			return false
		}
	}

	return true
}

// Matrices returns a slice of matrices, one for each incremental matrix.
func (ims IncrementalMatrices) Matrices() (out []gfmatrix.Matrix) {
	out = make([]gfmatrix.Matrix, len(ims))
	for i, im := range ims {
		out[i] = im.Matrix()
	}

	return out
}

// NewSBox takes a permutation vector as input and returns its corresponding S-Box. It inverts the S-Box if backwards is
// true (because the permutation vector we found was for the inverse S-box).
func NewSBox(v gfmatrix.Row, backwards bool) (out encoding.SBox) {
	for i, v_i := range v[0:256] {
		out.EncKey[i] = byte(v_i)
	}

	for i, j := range out.EncKey {
		out.DecKey[j] = byte(i)
	}

	if backwards { // Reverse EncKey and DecKey if we recover S^-1
		out.EncKey, out.DecKey = out.DecKey, out.EncKey
	}

	return
}

// XatY returns a byte array which is zero, except for the value x as position y.
func XatY(x byte, y int) (out [16]byte) {
	out[y] = x
	return
}

// GenerateRandomPlaintexts returns a random multiset of C[..]PC[..] plaintexts with the P at the given position.
func GenerateRandomPlaintexts(pos int) (out [][16]byte) {
	master := make([]byte, 16)
	rand.Read(master)

	for i := 0; i < 256; i++ {
		pt := [16]byte{}
		copy(pt[:], master)

		pt[pos] = byte(i)

		out = append(out, pt)
	}

	return
}

// FindPermutation takes a set of vectors and finds a linear combination of them that gives a permutation vector.
//
// Currently, this just takes random guesses until it finds a suitable choice. We'll normally be looking at a space of
// size 256^9, so traversing it isn't feasible and I don't know of a better search algorithm. Sorry for the
// non-determinism.
func FindPermutation(basis []gfmatrix.Row) gfmatrix.Row {
	for true {
		v := RandomLinearCombination(basis)

		if v[:256].IsPermutation() {
			return v
		}
	}

	return nil
}

// RandomLinearCombination returns a random linear combination of a set of basis vectors.
func RandomLinearCombination(basis []gfmatrix.Row) gfmatrix.Row {
	coeffs := make([]byte, len(basis))
	rand.Read(coeffs)

	v := gfmatrix.Row(make([]number.ByteFieldElem, basis[0].Size()))

	for i, c_i := range coeffs {
		v = v.Add(basis[i].ScalarMul(number.ByteFieldElem(c_i)))
	}

	return v
}
