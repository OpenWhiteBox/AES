package encoding

import (
	"crypto/rand"
	"testing"

	"github.com/OpenWhiteBox/AES/primitives/matrix"
)

func TestShuffle(t *testing.T) {
	s := GenerateShuffle(rand.Reader)

	for i := byte(0); i < 16; i++ {
		if s.Decode(s.Encode(i)) != i {
			t.Fatalf("Shuffle didn't Encode/Decode correctly.")
		}
	}
}

func TestByteLinear(t *testing.T) {
	M := matrix.GenerateRandom(rand.Reader, 8)
	MInv, _ := M.Invert()

	m := ByteLinear{M, MInv}

	for i := byte(0); i < 250; i++ {
		for j := byte(0); j < 250; j++ {
			if m.Decode(m.Encode(i)^m.Encode(j)) != i^j {
				t.Fatalf("Linear encoding didn't Encode/Decode correctly.")
			}
		}
	}
}
