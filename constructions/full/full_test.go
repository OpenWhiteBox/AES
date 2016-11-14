package full

import (
	"bytes"
	"testing"

	test_vectors "github.com/OpenWhiteBox/AES/constructions/test"
)

var (
	key   = []byte{72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100, 33, 33, 33, 33, 33}
	seed  = []byte{38, 41, 142, 156, 29, 181, 23, 194, 21, 250, 223, 183, 210, 168, 214, 145}
	input = []byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}
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

func TestPersistence(t *testing.T) {
	constr1, _, _ := GenerateKeys(key, seed)

	serialized := constr1.Serialize()
	constr2, err := Parse(serialized)

	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	cand1, cand2 := make([]byte, 16), make([]byte, 16)

	constr1.Encrypt(cand1, input)
	constr2.Encrypt(cand2, input)

	if !bytes.Equal(cand1, cand2) {
		t.Fatalf("Real disagrees with parsed! %x != %x", cand1, cand2)
	}
}
