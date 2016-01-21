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

// Shuffle implements a random 4-bit bijection.
type Shuffle struct {
	EncKey, DecKey [16]byte
}

// GenerateShuffle generates a random 4-bit bijection using the random source random (for example, crypto/rand.Reader).
func GenerateShuffle(reader io.Reader) (s Shuffle) {
	// Generate a random permutation as the encryption key.
	copy(s.EncKey[:], generatePermutation(reader, 16))

	// Invert the encryption key; set it as the decryption key.
	for i, j := range s.EncKey {
		s.DecKey[j] = byte(i)
	}

	return
}

func (s Shuffle) Encode(i byte) byte {
	return s.EncKey[i]
}

func (s Shuffle) Decode(i byte) byte {
	return s.DecKey[i]
}

// Permutation returns the shuffle's permutation.
func (s Shuffle) Permutation() (out []int) {
	for _, x := range s.DecKey {
		out = append(out, int(x))
	}

	return
}

// SBox implements a random 8-bit bijection.
type SBox struct {
	EncKey, DecKey [256]byte
}

// GenerateSBox generates a random 8-bit bijection using the random source random (for example, crypto/rand.Reader).
func GenerateSBox(reader io.Reader) (s SBox) {
	// Generate a random permutation as the encryption key.
	copy(s.EncKey[:], generatePermutation(reader, 256))

	// Invert the encryption key; set it as the decryption key.
	for i, j := range s.EncKey {
		s.DecKey[j] = byte(i)
	}

	return
}

func (s SBox) Encode(i byte) byte {
	return s.EncKey[i]
}

func (s SBox) Decode(i byte) byte {
	return s.DecKey[i]
}

// Permutation returns the SBox's permutation.
func (s SBox) Permutation() (out []int) {
	for _, x := range s.DecKey {
		out = append(out, int(x))
	}

	return
}
