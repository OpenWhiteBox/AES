package sas

import (
	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/gfmatrix"
	"github.com/OpenWhiteBox/AES/primitives/number"
)

// LastSBoxConstraints takes an SAS block cipher as input and, for each trailing S-box, it returns a matrix containing
// constraints on what the last S-boxes can be. Each 'constraint' is a row containing zeroes and ones, asserting that
// the dot product of it and the row [ S'(0) S'(1) ... S'(255) ] equals zero, where S' is the inverse S-box.
//
// The nullspace of this matrix describes every S-box that satisfies these conditions. One of the elements in the
// nullspace describe the 'real' trailing S-box of the cipher, but taking any bijective S-box will lead to an equivalent
// solution.
func LastSBoxConstraints(cipher encoding.Block) []gfmatrix.Matrix {
	ims := NewIncrementalMatrices(16, 256)

	for !ims.SufficientlyDefined() {
		pts := GenerateRandomPlaintexts(0)
		rows := gfmatrix.GenerateEmpty(16, 256)

		for _, pt := range pts {
			// C[..]PC[..] -(S)-> C[..]PC[..] -(A)->   D[..]   -(S)-> D[..] (See: Biryukov's Calculus)
			//    D[..]    -(S)->    D[..]    -(A)->   B[..]   -(S)-> x[..]
			ct := cipher.Encode(cipher.Encode(pt))

			// Accumulate the linear relationships of all the ciphertexts in the next empty row of the matrix.
			for i, j := range ct {
				rows[i][j] = rows[i][j].Add(0x01)
			}
		}

		for i, _ := range ims {
			ims[i].Add(rows[i])
		}
	}

	return ims.Matrices()
}

// FirstSBoxConstraints takes an SAS block cipher stripped of its trailing S-boxes and returns a slice of rows
// containing constraints on what the first S-box as position pos can be. Each 'constraint' is a row containing two ones
// and the rest zeroes. If the two ones are in positions i, and j, this means that
//   E(00...0i0...00) ^ E(00...0j0...00)                             // i and j are in the {pos}th position
//   = A * [ P_1(x_1) P_2(x_2) ... ] ^ A * [ P_1(y_1) P_2(y_2) ... ] // A is a matrix, P_i is an S-box
//   = A_pos * P_pos(i) + A_pos * P_pos(j) = target
//
// Knowing many pairs (i, j) such that P_pos(i) ^ P_pos(j) = x creates a system of linear equations that we can solve to
// find a candidate solution for P_pos.
func FirstSBoxConstraints(cipher encoding.Block, pos int) (out gfmatrix.Matrix) {
	im := gfmatrix.NewIncrementalMatrix(256)

	x := cipher.Encode(XatY(0x00, pos))
	y := cipher.Encode(XatY(target, pos))
	target := [16]byte{}
	encoding.XOR(target[:], x[:], y[:])

	out = gfmatrix.Matrix{}

	for i := 0; i < 255; i++ { // For every set of choices {i, j}:
		if out.FindPivot(0, i) != -1 { // Skip this choice for i if it was discovered as a j before.
			continue
		}

		x := cipher.Encode(XatY(byte(i), pos))

		for j := i + 1; j < 256; j++ {
			y := cipher.Encode(XatY(byte(j), pos))

			z := [16]byte{}
			encoding.XOR(z[:], x[:], y[:])

			if z == target {
				row := gfmatrix.Row(make([]number.ByteFieldElem, 256))
				row[i], row[j] = 0x01, 0x01

				out = append(out, row)
				break
			}
		}
	}

	return
}
