// Contains tools for key generation that don't fit anywhere else.
package chow

import (
	"../../primitives/matrix"
	"crypto/aes"
	"crypto/cipher"
	"io"
)

// Two Expand->Squash rounds comprise one AES round.  The Inside Surface is the output of the first E->S round (or the
// input of the second), and the Outside Surface is the output of the second E->S round (the whole AES round's output,
// feeding into the next round).
type Surface int

const (
	Inside Surface = iota
	Outside
)

// In a Squash step, we take four words and XOR them together into one word with 3 pairwise XORs: (a ^ b) ^ (c ^ d)
// This requires two top-level XOR tables, a Left computing (a ^ b) and a Right computing (c ^ d).  A bottom-level XOR
// takes the output of a Left and a Right gate and computes (a ^ b) ^ (c ^ d).
type Side int

const (
	Left Side = iota
	Right
)

// Index in, index out.  Example: shiftRows(5) = 1 because ShiftRows(block) returns [16]byte{block[0], block[5], ...
func shiftRows(i int) int {
	return []int{0, 13, 10, 7, 4, 1, 14, 11, 8, 5, 2, 15, 12, 9, 6, 3}[i]
}

// generateStream takes a (private) seed and a (possibly public) label and produces an io.Reader giving random bytes,
// useful for deterministically generating random matrices/encodings, in place of (crypto/rand).Reader.
//
// It does this by using the seed as an AES key and the label as the IV in CTR mode.  The io.Reader is providing the
// AES-CTR encryption of /dev/null.
type devNull struct{}

func (dn devNull) Read(p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		p[i] = 0
	}

	return len(p), nil
}

func generateStream(seed, label [16]byte) io.Reader {
	// Generate sub-key
	subKey := [16]byte{}
	c, _ := aes.NewCipher(seed[:])
	c.Encrypt(subKey[:], label[:])

	// Create pseudo-random byte stream keyed by sub-key.
	block, _ := aes.NewCipher(subKey[:])
	stream := cipher.StreamReader{
		cipher.NewCTR(block, label[:]),
		devNull{},
	}

	return stream
}

// Generate byte/word mixing bijections.
// TODO: Ensure that blocks are full-rank.
func ByteMixingBijection(seed [16]byte, round, position int) matrix.Matrix {
	label := [16]byte{}
	label[0], label[1], label[2], label[3] = 'M', 'B', byte(round), byte(position)

	return matrix.GenerateRandom(generateStream(seed, label), 8)
}

func WordMixingBijection(seed [16]byte, round, column int) matrix.Matrix {
	label := [16]byte{}
	label[0], label[1], label[2], label[3] = 'M', 'W', byte(round), byte(column)

	return matrix.GenerateRandom(generateStream(seed, label), 32)
}
