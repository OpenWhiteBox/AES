package sas

import (
	"crypto/rand"

	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/matrix"
)

func GenerateKeys() (constr Construction) {
	first, last := encoding.ConcatenatedBlock{}, encoding.ConcatenatedBlock{}
	for i := 0; i < 16; i++ {
		first[i] = encoding.GenerateSBox(rand.Reader)
		last[i] = encoding.GenerateSBox(rand.Reader)
	}

	constr.First, constr.Last = first, last

	constr.Linear = matrix.GenerateRandom(rand.Reader, 128)
	rand.Read(constr.Constant[:])

	return
}
