package encoding

import (
	"testing"
)

func TestShuffle(t *testing.T) {
	s := GenerateShuffle()

	for i := byte(0); i < 16; i++ {
		if s.Decode(s.Encode(i)) != i {
			t.Fatalf("Shuffle didn't Encode/Decode correctly.")
		}
	}
}
