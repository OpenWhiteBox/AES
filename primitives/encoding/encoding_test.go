package encoding

import (
	"crypto/rand"
	"testing"
)

func TestShuffle(t *testing.T) {
	s := GenerateShuffle(rand.Reader)

	for i := byte(0); i < 16; i++ {
		if s.Decode(s.Encode(i)) != i {
			t.Fatalf("Shuffle didn't Encode/Decode correctly.")
		}
	}
}
