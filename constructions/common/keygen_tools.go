package common

import (
	"crypto/aes"
	"crypto/cipher"
	"io"

	"github.com/OpenWhiteBox/AES/primitives/encoding"
)

type Surface int

const (
	Inside Surface = iota
	Outside
)

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

func GenerateStream(seed, label []byte) io.Reader {
	// Generate sub-key
	subKey := make([]byte, 16)
	c, _ := aes.NewCipher(seed)
	c.Encrypt(subKey, label)

	// Create pseudo-random byte stream keyed by sub-key.
	block, _ := aes.NewCipher(subKey)
	stream := cipher.StreamReader{
		cipher.NewCTR(block, label),
		devNull{},
	}

	return stream
}

func GetShuffle(seed, label []byte) encoding.Shuffle {
	key := [32]byte{}
	copy(key[0:16], seed)
	copy(key[16:32], label)

	cached, ok := encodingCache[key]

	if ok {
		return cached
	} else {
		encodingCache[key] = encoding.GenerateShuffle(GenerateStream(seed, label))
		return encodingCache[key]
	}
}
