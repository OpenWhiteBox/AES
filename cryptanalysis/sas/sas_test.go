package sas

import (
	"testing"

	"bytes"
	"crypto/rand"

	"github.com/OpenWhiteBox/AES/primitives/encoding"
	matrix "github.com/OpenWhiteBox/AES/primitives/matrix2"
	"github.com/OpenWhiteBox/AES/primitives/number"

	"github.com/OpenWhiteBox/AES/constructions/sas"
)

func strippedEncrypt(constr sas.Construction, inner, outer encoding.Block) func([]byte, []byte) {
	return func(dst, src []byte) {
		in, out := [16]byte{}, [16]byte{}
		copy(in[:], src)

		in = inner.Decode(in)
		constr.Encrypt(out[:], in[:])
		out = outer.Decode(out)

		copy(dst, out[:])
	}
}

func testAffine(constr sas.Construction, inner, outer encoding.Block) bool {
	one, two, three, final := make([]byte, 16), make([]byte, 16), make([]byte, 16), make([]byte, 16)
	one[0], two[1], three[2] = 0x01, 0x01, 0x01
	final[0], final[1], final[2] = 0x01, 0x01, 0x01

	strippedEncrypt(constr, inner, outer)(one, one)
	strippedEncrypt(constr, inner, outer)(two, two)
	strippedEncrypt(constr, inner, outer)(three, three)

	// If stripped constr.Encrypt is an affine transformation, then three ciphertexts XORed together should equal the
	// encryption of the three plaintexts XORed together.
	strippedEncrypt(constr, inner, outer)(final, final)
	cand := xor(xor(one, two), three)

	return bytes.Compare(final, cand) == 0
}

func xor(a, b []byte) []byte {
	out := make([]byte, 16)

	for i := 0; i < 16; i++ {
		out[i] = a[i] ^ b[i]
	}

	return out
}

func TestDecomposeSAS(t *testing.T) {
	constr1 := sas.GenerateKeys()

	first, linear, constant, last := DecomposeSAS(constr1)

	constr2 := sas.Construction{
		First:    first,
		Linear:   linear,
		Constant: constant,
		Last:     last,
	}

	// Check that both constructions give the same output on a random challenge.
	challenge := make([]byte, 16)
	rand.Read(challenge)

	cand1, cand2 := make([]byte, 16), make([]byte, 16)
	constr1.Encrypt(cand1, challenge)
	constr2.Encrypt(cand2, challenge)

	if bytes.Compare(cand1, cand2) != 0 {
		t.Fatalf("Construction #1 and construction #2 didn't agree at a random challenge!")
	}
}

func TestRecoverFirstSBoxes(t *testing.T) {
	constr := sas.GenerateKeys()
	outer := constr.Last
	inner := encoding.ConcatenatedBlock(RecoverFirstSBoxes(constr, outer))

	if !testAffine(constr, inner, outer) {
		t.Fatalf("Affine test on encryption without s-boxes on either side failed.")
	}
}

func TestRecoverLastSBoxes(t *testing.T) {
	constr := sas.GenerateKeys()
	outer := encoding.ConcatenatedBlock(RecoverLastSBoxes(constr))

	if !testAffine(constr, constr.First, outer) {
		t.Fatalf("Affine test on encryption without tailing s-boxes failed.")
	}
}

func TestGenerateInnerBalance(t *testing.T) {
	constr := sas.GenerateKeys()
	outer := constr.Last
	pos := 0

	target := xorArray(
		outer.Decode(EncryptAtPosition(constr, pos, 0x00)),
		outer.Decode(EncryptAtPosition(constr, pos, 0x01)),
	)

	m := GenerateInnerBalance(constr, outer, pos, target)

	sbox := constr.First.(encoding.ConcatenatedBlock)[pos].(encoding.SBox)
	c := sbox.Encode(0x00) ^ sbox.Encode(0x01)

	sboxRow := matrix.Row(make([]number.ByteFieldElem, 256))
	for i := 0; i < 256; i++ {
		sboxRow[i] = number.ByteFieldElem(sbox.EncKey[i])
	}

	// Verify that the dot product of each balance row with the S-Box equals the correct constant.
	for _, row := range m {
		cand := byte(row.DotProduct(sboxRow))
		if cand != c {
			t.Fatalf("GenerateInnerBalance gave an incorrect relation! %v != %v", cand, c)
		}
	}
}
