// Package random implements the generation of random objects with controlled, cryptographically secure randomness.
//
// Three parameters matter: the name and seed of the randomness source, and the label given when requesting an object.
// If all of these three parameters are the same, the objects returned will be the same--if they're different, the
// returned object will likely be different. To prevent an adversary from being able to predict what will be returned,
// the only parameter that needs to be kept secret is the seed of the randomness source.
package random

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

// Source implements generators of random objects. It also maintains a cache to speed up generation, in cases where the
// same object may be requested many times.
type Source struct {
	// The name of the randomness source--an arbitrary string.
	Name string
	// A 16-byte truly random seed.
	Seed []byte

	shuffleCache map[[16]byte]encoding.Shuffle
	sboxCache    map[[16]byte]encoding.SBox
	matrixCache  map[[16]byte]matrix.Matrix
}

// NewSource initializes a Source object with the given name and seed.
func NewSource(name string, seed []byte) Source {
	return Source{
		Name:         name,
		Seed:         seed,
		shuffleCache: make(map[[16]byte]encoding.Shuffle),
		sboxCache:    make(map[[16]byte]encoding.SBox),
		matrixCache:  make(map[[16]byte]matrix.Matrix),
	}
}

// subKey generates a random key from the context and label that can be used for cryptographic primitives.
func (rs *Source) subKey(label []byte) []byte {
	subKey := make([]byte, 16)
	c, _ := aes.NewCipher(rs.Seed)
	c.Encrypt(subKey, label)

	for i, c := range []byte(rs.Name) {
		subKey[i] ^= c
	}

	c.Encrypt(subKey, subKey)

	return subKey
}

// Stream takes a (possibly public) label and produces an io.Reader giving random bytes, useful for deterministically
// generating random matrices/encodings, in place of (crypto/rand).Reader.
//
// It does this by using the seed as an AES key and the label as the IV in CTR mode.  The io.Reader is providing the
// AES-CTR encryption of /dev/null.
func (rs *Source) Stream(label []byte) io.Reader {
	subKey := rs.subKey(label)

	// Create pseudo-random byte stream keyed by sub-key.
	block, _ := aes.NewCipher(subKey)
	stream := cipher.StreamReader{
		S: cipher.NewCTR(block, label),
		R: devNull{},
	}

	return stream
}

// Shuffle takes a (possibly public) label and produces a random shuffle of the integers [0, 16).
func (rs *Source) Shuffle(label []byte) encoding.Shuffle {
	key := [16]byte{}
	copy(key[:], label)

	cached, ok := rs.shuffleCache[key]

	if ok {
		return cached
	} else {
		rs.shuffleCache[key] = encoding.GenerateShuffle(rs.Stream(label))
		return rs.shuffleCache[key]
	}
}

// SBox takes a (possibly public) label and produces a random S-box.
func (rs *Source) SBox(label []byte) encoding.SBox {
	key := [16]byte{}
	copy(key[:], label)

	cached, ok := rs.sboxCache[key]

	if ok {
		return cached
	} else {
		rs.sboxCache[key] = encoding.GenerateSBox(rs.Stream(label))
		return rs.sboxCache[key]
	}
}

// Matrix takes a (possibly public) label and produces a random non-singular matrix.
func (rs *Source) Matrix(label []byte, size int) matrix.Matrix {
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

// Dirichlet takes a (possibly public) label and produces the output of a uniform dirichlet distribution with `length`
// variables, summing to `sum`.
func (rs *Source) Dirichlet(label []byte, length, sum int) []int {
	if length == 0 && sum != 0 {
		panic("Dirichlet: Can't sample distribution of zero variables and get a non-zero sum!")
	} else if sum == 0 {
		return make([]int, length)
	}

	out := make([]int, length)
	s := rs.Stream(label)

	// Typical way of sampling a Dirichlet distribution:
	// http://stats.stackexchange.com/questions/69210/drawing-from-dirichlet-distribution
	// Takes multiple guesses because of rounding errors.
	guess := func() {
		buff := make([]byte, length)
		s.Read(buff)

		buff[length-1] /= 3

		candSum := 0
		for pos := 0; pos < length; pos++ {
			out[pos] = int(buff[pos])
			candSum += out[pos]
		}

		if candSum == 0 { // Avoid division by zero error.
			return
		}

		for pos := 0; pos < length; pos++ {
			out[pos] = out[pos] * sum

			r := out[pos] % candSum
			out[pos] = out[pos] / candSum

			if r >= candSum/2 {
				out[pos]++
			}
		}
	}

	candSum := 0
	for candSum != sum {
		candSum = 0
		guess()

		for _, x := range out {
			candSum += x
		}
	}

	return out
}

// Monotone takes a (possibly public) label and produces a random monotone function which is `length` units long and
// maximizes at `max`.
func (rs *Source) Monotone(label []byte, length, max int) []int {
	out := rs.Dirichlet(label, length, max)

	for i := 1; i < length; i++ {
		out[i] += out[i-1]
	}

	return out
}
