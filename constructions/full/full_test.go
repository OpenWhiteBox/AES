package full

import (
	"bytes"
	"testing"

	test_vectors "github.com/OpenWhiteBox/AES/constructions/test"
)

func TestEncrypt(t *testing.T) {
	for n, vec := range test_vectors.GetAESVectors(testing.Short()) {
		constr, inputMask, outputMask := GenerateKeys(vec.Key, vec.Key)

		in, out := [16]byte{}, [16]byte{}

		copy(in[:], vec.In)
		in = inputMask.Decode(in) // Apply input encoding.

		constr.Encrypt(out[:], in[:])

		out = outputMask.Decode(out) // Remove output encoding.

		if !bytes.Equal(vec.Out, out[:]) {
			t.Fatalf("Real disagrees with result in test vector %v! %x != %x", n, vec.Out, out)
		}

		break // Only do one. GenerateKeys is really slow.
	}
}
