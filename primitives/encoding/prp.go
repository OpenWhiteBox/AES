// A PRP is a special form of encoding that pseudo-randomly permutes the input space.
// TODO: Should small-space PRPs be added / used instead of Shuffle?
package encoding

import (
	"io"
)

func marshall(x byte, mod int) int {
	y := int(x)

	// Trim every bit off of y that mod doesn't have.
	for i := uint(2); i < 8; i++ {
		if mod < (1 << i) {
			y &= (1 << i) - 1
			break
		}
	}

	return y
}

func generatePermutation(reader io.Reader, size int) []byte {
	out := make([]byte, size)
	for i := 0; i < size; i++ {
		out[i] = byte(i)
	}

	ptr := 0
	for ptr < size {
		buffer := make([]byte, 32)
		reader.Read(buffer)

		for _, b := range buffer {
			c := marshall(b, size-ptr) + ptr

			if c < size {
				out[ptr], out[c] = out[c], out[ptr]
				ptr++

				if ptr == size {
					break
				}
			}
		}
	}

	return out
}

type Shuffle struct {
	EncKey, DecKey [16]byte
}

func (s Shuffle) Encode(i byte) byte {
	return s.EncKey[i]
}

func (s Shuffle) Decode(i byte) byte {
	return s.DecKey[i]
}

func GenerateShuffle(reader io.Reader) (s Shuffle) {
	// Generate a random permutation as the encryption key.
	copy(s.EncKey[:], generatePermutation(reader, 16))

	// Invert the encryption key; set it as the decryption key.
	for i, j := range s.EncKey {
		s.DecKey[j] = byte(i)
	}

	return
}

type SBox struct {
	EncKey, DecKey [256]byte
}

func (s SBox) Encode(i byte) byte {
	return s.EncKey[i]
}

func (s SBox) Decode(i byte) byte {
	return s.DecKey[i]
}

func GenerateSBox(reader io.Reader) (s SBox) {
	// Generate a random permutation as the encryption key.
	copy(s.EncKey[:], generatePermutation(reader, 256))

	// Invert the encryption key; set it as the decryption key.
	for i, j := range s.EncKey {
		s.DecKey[j] = byte(i)
	}

	return
}
