package sas

import (
	"io"

	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/matrix"
)

// GenerateKeys generates a random SAS instance using the random source random (for example, crypto/rand.Reader).
func GenerateKeys(rand io.Reader) (constr Construction) {
	first, last := encoding.ConcatenatedBlock{}, encoding.ConcatenatedBlock{}
	for i := 0; i < 16; i++ {
		first[i] = encoding.GenerateSBox(rand)
		last[i] = encoding.GenerateSBox(rand)
	}

	constr.First, constr.Last = first, last

	c := [16]byte{}
	rand.Read(c[:])

	constr.Affine = encoding.NewBlockAffine(matrix.GenerateRandom(rand, 128), c)

	return
}
