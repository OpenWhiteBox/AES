package toy

import (
	"github.com/OpenWhiteBox/primitives/encoding"
	"github.com/OpenWhiteBox/primitives/gfmatrix"
)

// permutation is a byte-wise permutation of a block.
type permutation encoding.Shuffle

// newPermutation returns an identity permutation.
func newPermutation() *permutation {
	p := &permutation{}

	for i := byte(0); i < 16; i++ {
		p.EncKey[i], p.DecKey[i] = i, i
	}

	return p
}

func (p *permutation) Encode(in [16]byte) (out [16]byte) {
	for i := byte(0); i < 16; i++ {
		out[(*encoding.Shuffle)(p).Encode(i)] = in[i]
	}

	return out
}

func (p *permutation) Decode(in [16]byte) (out [16]byte) {
	for i := byte(0); i < 16; i++ {
		out[(*encoding.Shuffle)(p).Decode(i)] = in[i]
	}

	return out
}

func (p *permutation) swap(i, j int) {
	p.DecKey[p.EncKey[i]], p.DecKey[p.EncKey[j]] = p.DecKey[p.EncKey[j]], p.DecKey[p.EncKey[i]]
	p.EncKey[i], p.EncKey[j] = p.EncKey[j], p.EncKey[i]
}

// smallMatrix is a round matrix compressed to only its MixColumns coefficient.
type smallMatrix gfmatrix.Matrix

func (sm smallMatrix) swapRows(x, y int) {
	sm[x], sm[y] = sm[y], sm[x]
}

func (sm smallMatrix) swapCols(x, y int) {
	for row := 0; row < 16; row++ {
		sm[row][x], sm[row][y] = sm[row][y], sm[row][x]
	}
}

// unpermute transforms sm until it is equal to a compressed round matrix. sm is mutated to the compressed round matrix,
// and permIn and permOut are the permutations that moved it there.
func (sm smallMatrix) unpermute() (permIn, permOut *permutation) {
	permIn, permOut = newPermutation(), newPermutation()
	ok := sm.unpermuteStep(0, permIn, permOut)
	if !ok {
		panic("unable to unpermute matrix")
	}
	return
}

// unpermuteStep is one step of the dynamic programming algorithm. It works by moving different twos to the diagonal and
// checking if this leads to a partially correct matrix.
func (sm smallMatrix) unpermuteStep(step int, permIn, permOut *permutation) bool {
	if step == 16 {
		return true
	}

	// Build a list of all twos below.
	twos := [][2]int{}
	for row := step; row < 16; row++ {
		for col := step; col < 16; col++ {
			if sm[row][col] == 2 {
				twos = append(twos, [2]int{row, col})
			}
		}
	}

	// Foreach two, try to put it in the diagonal. Check neighboring entries for correctness. Recurse.
	for _, two := range twos {
		// Put this two in the diagonal.
		sm.swapRows(step, two[0])
		sm.swapCols(step, two[1])

		permIn.swap(step, two[1])
		permOut.swap(step, two[0])

		// Check for correctness.
		ok := true
		for row := 0; row < step; row++ {
			ok = ok && sm[row][step] == smallRound[row][step]
		}
		for col := 0; col < step; col++ {
			ok = ok && sm[step][col] == smallRound[step][col]
		}
		if !ok {
			continue
		}

		// Recurse down.
		if sm.unpermuteStep(step+1, permIn, permOut) {
			return true
		}

		// Couldn't find lower solution. Undo swaps.
		sm.swapRows(step, two[0])
		sm.swapCols(step, two[1])

		permIn.swap(step, two[1])
		permOut.swap(step, two[0])
	}

	return false
}
