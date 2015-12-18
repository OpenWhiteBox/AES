package sas

import (
	// "fmt"
	"testing"

	"bytes"

	"github.com/OpenWhiteBox/AES/primitives/encoding"
	binmatrix "github.com/OpenWhiteBox/AES/primitives/matrix"

	"github.com/OpenWhiteBox/AES/constructions/sas"
)

func strippedEncrypt(constr sas.Construction, inner encoding.Block, linear binmatrix.Matrix, outer encoding.Block) func([]byte, []byte) {
	return func(dst, src []byte) {
		in, out := [16]byte{}, [16]byte{}
		copy(in[:], src)

		in = inner.Decode(in)

		constr.Encrypt(out[:], in[:])

		copy(out[:], linear.Mul(binmatrix.Row(out[:])))

		out = outer.Decode(out)

		copy(dst, out[:])
	}
}

func xor(a, b []byte) []byte {
	out := make([]byte, 16)

	for i := 0; i < 16; i++ {
		out[i] = a[i] ^ b[i]
	}

	return out
}

func TestRecoverLinear(t *testing.T) {
	constr := sas.GenerateKeys()
	sboxes := encoding.ConcatenatedBlock(RecoverLastSBoxes(constr))

	RecoverFirstSBox(constr, sboxes, 0)
	// _, ok := lin.Invert()
	// fmt.Println(ok)

	// one, two := make([]byte, 16), make([]byte, 16)
	// two[0] = 0x01
	//
	// strippedEncrypt(constr, constr.First, x, sboxes)(one, one)
	// strippedEncrypt(constr, constr.First, x, sboxes)(two, two)
	//
	// fmt.Println(one)
	// fmt.Println(two)

	// fmt.Println(lin)
	// fmt.Println(lin.Compose(constr.Linear))

	// cand := make([]byte, 16)
	//
	// zero := make([]byte, 16)
	// strippedEncrypt(constr, constr.First, c, sboxes)(cand, zero)

	// fmt.Println(cand)
	// one, two, final := make([]byte, 16), make([]byte, 16), make([]byte, 16)
	// one[0], two[1] = 0x01, 0x01, 0x01
	// final[0], final[1], final[2] = 0x01, 0x01, 0x01
	//
	//
	// strippedEncrypt(constr, constr.First, [16]byte{}, sboxes)(one, one)
	// strippedEncrypt(constr, constr.First, [16]byte{}, sboxes)(two, two)
	// strippedEncrypt(constr, constr.First, [16]byte{}, sboxes)(three, three)
}

func TestRecoverLastSBoxes(t *testing.T) {
	constr := sas.GenerateKeys()
	sboxes := encoding.ConcatenatedBlock(RecoverLastSBoxes(constr))

	one, two, three, final := make([]byte, 16), make([]byte, 16), make([]byte, 16), make([]byte, 16)
	one[0], two[1], three[2] = 0x01, 0x01, 0x01
	final[0], final[1], final[2] = 0x01, 0x01, 0x01

	strippedEncrypt(constr, constr.First, binmatrix.GenerateIdentity(128), sboxes)(one, one)
	strippedEncrypt(constr, constr.First, binmatrix.GenerateIdentity(128), sboxes)(two, two)
	strippedEncrypt(constr, constr.First, binmatrix.GenerateIdentity(128), sboxes)(three, three)

	// If encryptWithoutSBoxes is an affine transformation (which it should be), then three ciphertexts XORed together
	// should equal the encrypt of the three plaintexts XORed together.
	strippedEncrypt(constr, constr.First, binmatrix.GenerateIdentity(128), sboxes)(final, final)
	cand := xor(xor(one, two), three)

	if bytes.Compare(final, cand) != 0 {
		t.Fatalf("Affine test on encryption without s-boxes failed.")
	}
}
