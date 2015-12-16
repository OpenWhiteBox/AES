package sas

import (
	"testing"

	"bytes"

	"github.com/OpenWhiteBox/AES/primitives/encoding"

	"github.com/OpenWhiteBox/AES/constructions/sas"
)

func encryptWithoutSBoxes(constr sas.Construction, sboxes [16]encoding.Byte, input []byte) []byte {
	in, out := [16]byte{}, [16]byte{}
	copy(in[:], input)

	in = constr.First.Decode(in)
	constr.Encrypt(out[:], in[:])
	out = encoding.ConcatenatedBlock(sboxes).Decode(out)

	return out[:]
}

func xor(a, b, c []byte) []byte {
	out := make([]byte, 16)

	for i := 0; i < 16; i++ {
		out[i] = a[i] ^ b[i] ^ c[i]
	}

	return out
}

func TestRecoverLastSBoxes(t *testing.T) {
	constr := sas.GenerateKeys()
	sboxes := RecoverLastSBoxes(constr)

	one, two, three, final := make([]byte, 16), make([]byte, 16), make([]byte, 16), make([]byte, 16)
	one[0], two[1], three[2] = 0x01, 0x01, 0x01
	final[0], final[1], final[2] = 0x01, 0x01, 0x01

	a := encryptWithoutSBoxes(constr, sboxes, one)
	b := encryptWithoutSBoxes(constr, sboxes, two)
	c := encryptWithoutSBoxes(constr, sboxes, three)

	// If encryptWithoutSBoxes is an affine transformation (which it should be), then three ciphertexts XORed together
	// should equal the encrypt of the three plaintexts XORed together.
	x := encryptWithoutSBoxes(constr, sboxes, final)
	y := xor(a, b, c)

	if bytes.Compare(x, y) != 0 {
		t.Fatalf("Affine test on encryption without s-boxes failed.")
	}
}
