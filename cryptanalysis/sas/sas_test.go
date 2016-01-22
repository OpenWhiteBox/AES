package sas

import (
	"testing"

	"bytes"
	"crypto/rand"

	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/gfmatrix"
	"github.com/OpenWhiteBox/AES/primitives/number"

	"github.com/OpenWhiteBox/AES/constructions/sas"
)

func testAffine(cipher encoding.Block) bool {
	aff := encoding.DecomposeBlockAffine(cipher)
	return encoding.ProbablyEquivalentBlocks(cipher, aff)
}

func TestDecomposeSAS(t *testing.T) {
	constr1 := sas.GenerateKeys(rand.Reader)
	constr2 := DecomposeSAS(constr1)

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

func TestRecoverLastSBoxes(t *testing.T) {
	constr := sas.GenerateKeys(rand.Reader)
	cipher := Encoding{constr}

	last := RecoverLastSBoxes(cipher)

	aff := encoding.ComposedBlocks{
		encoding.InverseBlock{constr.First},
		cipher,
		encoding.InverseBlock{last},
	}

	if !testAffine(aff) {
		t.Fatalf("Affine test on encryption without tailing s-boxes failed.")
	}
}

func TestRecoverFirstSBoxes(t *testing.T) {
	constr := sas.GenerateKeys(rand.Reader)
	cipher := encoding.ComposedBlocks{
		Encoding{constr},
		encoding.InverseBlock{constr.Last},
	}

	first := RecoverFirstSBoxes(cipher)

	aff := encoding.ComposedBlocks{
		encoding.InverseBlock{first},
		cipher,
	}

	if !testAffine(aff) {
		t.Fatalf("Affine test on encryption without s-boxes on either side failed.")
	}
}

func TestGenerateFirstBalance(t *testing.T) {
	pos := 0
	constr := sas.GenerateKeys(rand.Reader)
	cipher := encoding.ComposedBlocks{
		Encoding{constr},
		encoding.InverseBlock{constr.Last},
	}

	a, b := cipher.Encode(XatY(0x00, pos)), cipher.Encode(XatY(0x01, pos))
	target := [16]byte{}
	encoding.XOR(target[:], a[:], b[:])

	m := FirstSBoxConstraints(cipher, pos, target)

	sbox := constr.First[pos].(encoding.SBox)
	c := sbox.Encode(0x00) ^ sbox.Encode(0x01)

	sboxRow := gfmatrix.Row(make([]number.ByteFieldElem, 256))
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
