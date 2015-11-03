package common

import (
	"crypto/aes"
	"crypto/cipher"
	"io"

	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/matrix"
)

type devNull struct{}

func (dn devNull) Read(p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		p[i] = 0
	}

	return len(p), nil
}

type RandomSource struct {
	Name string
	Seed []byte

	encodingCache map[[16]byte]encoding.Shuffle
	matrixCache   map[[16]byte]matrix.Matrix
}

func NewRandomSource(name string, seed []byte) RandomSource {
	return RandomSource{
		name, seed, make(map[[16]byte]encoding.Shuffle), make(map[[16]byte]matrix.Matrix),
	}
}

// subKey generates a random key from the context and label that can be used for cryptographic primitives.
func (rs *RandomSource) subKey(label []byte) []byte {
	subKey := make([]byte, 16)
	c, _ := aes.NewCipher(rs.Seed)
	c.Encrypt(subKey, label)

	for i, c := range []byte(rs.Name) {
		subKey[i] ^= c
	}

	c.Encrypt(subKey, subKey)

	return subKey
}

// Stream takes a (private) seed and a (possibly public) label and produces an io.Reader giving random bytes, useful for
// deterministically generating random matrices/encodings, in place of (crypto/rand).Reader.
//
// It does this by using the seed as an AES key and the label as the IV in CTR mode.  The io.Reader is providing the
// AES-CTR encryption of /dev/null.
func (rs *RandomSource) Stream(label []byte) io.Reader {
	subKey := rs.subKey(label)

	// Create pseudo-random byte stream keyed by sub-key.
	block, _ := aes.NewCipher(subKey)
	stream := cipher.StreamReader{
		cipher.NewCTR(block, label),
		devNull{},
	}

	return stream
}

// Shuffle takes a (private) seed and a (possibly public) label and produces a random shuffle of the integers [0, 16).
func (rs *RandomSource) Shuffle(label []byte) encoding.Shuffle {
	key := [16]byte{}
	copy(key[:], label)

	cached, ok := rs.encodingCache[key]

	if ok {
		return cached
	} else {
		rs.encodingCache[key] = encoding.GenerateShuffle(rs.Stream(label))
		return rs.encodingCache[key]
	}
}

// Matrix takes a (private) seed and a (possibly public) label and produces a random non-singular 128x128 matrix.
func (rs *RandomSource) Matrix(label []byte, size int) matrix.Matrix {
	key := [16]byte{}
	copy(key[:], label)

	cached, ok := rs.matrixCache[key]

	if ok {
		return cached
	} else {
		rs.matrixCache[key] = matrix.GenerateRandom(rs.Stream(label), size)
		return rs.matrixCache[key]
	}
}
