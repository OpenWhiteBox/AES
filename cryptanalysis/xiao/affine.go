package xiao

import (
	"github.com/OpenWhiteBox/primitives/encoding"
	"github.com/OpenWhiteBox/primitives/matrix"
)

// When you compose a matrix with 16-by-16 blocks with ShiftRows, you get a matrix with one 16-by-8 block in each
// column. The value in blockPos at position i is the vertical position of the ith block.
var blockPos = []int{0, 6, 5, 3, 2, 0, 7, 5, 4, 2, 1, 7, 6, 4, 3, 1}

// blockOfInverse computes a block of the something something something.
func blockOfInverse(swap [2]int, eqs [4]int) matrix.Matrix {
	unmixcol := unMixColumn.Dup()

	// Swap chosen rows.
	for i, ok := range swap {
		if ok == 1 {
			for row := 16 * i; row < 16*i+8; row++ {
				unmixcol[row], unmixcol[8+row] = unmixcol[8+row], unmixcol[row]
			}
		}
	}

	// Generate matrix corresponding to self-equivalence noise from S-box.
	noise := matrix.GenerateEmpty(32, 32)

	for i, eq := range eqs {
		for row := 0; row < 8; row++ {
			noise[8*i+row][i] = equivalences[eq][row][0]
		}
	}

	return noise.Compose(unmixcol)
}

// affineLayer implements methods for disambiguating an affine layer of the SPN.
type affineLayer encoding.BlockAffine

func (al affineLayer) Encode(in [16]byte) [16]byte {
	return encoding.BlockAffine(al).Encode(in)
}

func (al affineLayer) Decode(in [16]byte) [16]byte {
	return encoding.BlockAffine(al).Decode(in)
}

// leftCompose composes a Block encoding on the left.
func (al *affineLayer) leftCompose(left encoding.Block) {
	temp, _ := encoding.DecomposeBlockAffine(encoding.ComposedBlocks{
		left, encoding.BlockAffine(*al),
	})

	*al = affineLayer(temp)
}

// rightCompose composes a Block encoding on the right.
func (al *affineLayer) rightCompose(right encoding.Block) {
	temp, _ := encoding.DecomposeBlockAffine(encoding.ComposedBlocks{
		encoding.BlockAffine(*al), right,
	})

	*al = affineLayer(temp)
}

// FindPermutation is called on the first affine layer. It returns the permutation matrix corresponding to the row
// permutation that has occured to the first layer.
func (al *affineLayer) findPermutation() matrix.Matrix {
	permed := (*al).BlockLinear.Forwards
	unpermed := matrix.Matrix{}

	for i := 0; i < 8; i++ {
		for pos := 0; pos < 16; pos++ {
			if h := permed[8*pos].Height(); 16*i <= h && h < 16*(i+1) {
				unpermed = append(unpermed, permed[8*pos:8*(pos+1)]...)
			}
		}
	}

	unpermed, _ = unpermed.Invert()
	perm := permed.Compose(unpermed)

	return perm
}

// cleanLeft gets the last affine layer back to a matrix with 16-by-16 blocks along the diagonal, times ShiftRows, times
// MixColumns and returns the matrix on the input encoding that it used to do this.
func (al *affineLayer) cleanLeft() encoding.Block {
	inverse := matrix.GenerateEmpty(128, 128)
	mixcols := matrix.GenerateEmpty(128, 128)

	// Combine individual blocks of the inverse matrix into the full inverse matrix. Also build the matrix corresponding
	// to the full-block MixColumns operation.
	for block := 0; block < 4; block++ {
		inv := al.findBlockOfInverse(block)

		for row := 0; row < 32; row++ {
			copy(inverse[32*block+row][4*block:], inv[row])
			copy(mixcols[32*block+row][4*block:], mixColumn[row])
		}
	}

	out := encoding.NewBlockLinear(inverse.Compose(mixcols))
	al.leftCompose(out)
	return encoding.InverseBlock{out}
}

// findBlockOfInverse finds any S-box transpositions or self-equivalence noise that may be hiding in the given block of
// the last affine layer and returns them.
func (al *affineLayer) findBlockOfInverse(block int) matrix.Matrix {
	for swap1 := 0; swap1 < 2; swap1++ {
		for swap2 := 0; swap2 < 2; swap2++ {
			for p1 := 0; p1 < 8; p1++ {
				for p2 := 0; p2 < 8; p2++ {
					for p3 := 0; p3 < 8; p3++ {
						for p4 := 0; p4 < 8; p4++ {
							cand := blockOfInverse([2]int{swap1, swap2}, [4]int{p1, p2, p3, p4})

							if al.isBlockOfInverse(block, cand) {
								return cand
							}
						}
					}
				}
			}
		}
	}

	panic("Could not find block of inverse!")
}

// isBlockOfInverse takes a candidate solution for the given block of the matrix and returns true if it is valid and
// false if it isn't.
func (al *affineLayer) isBlockOfInverse(block int, cand matrix.Matrix) bool {
	// Pad matrix.
	inv := matrix.GenerateEmpty(32*block, 32)
	for _, row := range cand {
		inv = append(inv, row)
	}
	for row := 0; row < 96-32*block; row++ {
		inv = append(inv, matrix.NewRow(32))
	}

	// Test if this is consistent with inverse.
	res := (*al).BlockLinear.Forwards.Compose(inv).Transpose()

	for i := 0; i < 4; i++ {
		row, pos := res[8*i], blockPos[4*block+i]

		if h := row.Height(); !(16*pos <= h && h < 16*(pos+1)) {
			return false
		}

		if !row[2*(pos+1):].IsZero() {
			return false
		}
	}

	return true
}

// getBlock returns the 8-by-8 block of the affine layer at the given position.
func (al *affineLayer) getBlock(row, col int) matrix.Matrix {
	out := matrix.Matrix{}

	for i := 0; i < 8; i++ {
		out = append(out, matrix.Row{al.BlockLinear.Forwards[8*row+i][col]})
	}

	return out
}
