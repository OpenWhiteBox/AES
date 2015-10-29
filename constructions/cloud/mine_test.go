package cloud

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/OpenWhiteBox/AES/primitives/matrix"

	"github.com/OpenWhiteBox/AES/constructions/saes"

	test_vectors "github.com/OpenWhiteBox/AES/constructions/test"
)

func TestSubBytes(t *testing.T) {
	vect := make([]byte, 16)
	rand.Read(vect)

	baseConstr := saes.Construction{make([]byte, 16)}
	constr := Construction{baseConstr.StretchedKey()}

	real := make([]byte, 16)
	copy(real, vect)
	baseConstr.SubBytes(real)

	cand := make([]byte, 16)
	copy(cand, vect)

	constr.Invert(cand)
	copy(cand, SubBytes.Mul(matrix.Row(cand)))
	constr.AddConstant(sbConst, cand)

	if bytes.Compare(real, cand) != 0 {
		t.Fatalf("SubBytes matrix was wrong for input: %x\nReal: %x\nCand: %x", vect, real, cand)
	}
}

func TestMixColumns(t *testing.T) {
	vect := make([]byte, 16)
	rand.Read(vect)

	baseConstr := saes.Construction{make([]byte, 16)}

	real := make([]byte, 16)
	copy(real, vect)
	baseConstr.MixColumns(real)

	cand := make([]byte, 16)
	copy(cand, MixColumns.Mul(matrix.Row(vect)))

	if bytes.Compare(real, cand) != 0 {
		t.Fatalf("MixColumns matrix was wrong for input: %x\nReal: %x\nCand: %x", vect, real, cand)
	}
}

func TestShiftRows(t *testing.T) {
	vect := make([]byte, 16)
	rand.Read(vect)

	baseConstr := saes.Construction{make([]byte, 16)}

	real := make([]byte, 16)
	copy(real, vect)
	baseConstr.ShiftRows(real)

	cand := make([]byte, 16)
	copy(cand, ShiftRows.Mul(matrix.Row(vect)))

	if bytes.Compare(real, cand) != 0 {
		t.Fatalf("ShiftRows matrix was wrong for input: %x\nReal: %x\nCand: %x", vect, real, cand)
	}
}

func TestEncrypt(t *testing.T) {
	for n, vec := range test_vectors.AESVectors {
		baseConstr := saes.Construction{vec.Key}
		constr := Construction{baseConstr.StretchedKey()}

		in, out := make([]byte, 16), make([]byte, 16)
		copy(in, vec.In)

		constr.Encrypt(out, in)

		if !bytes.Equal(vec.Out, out) {
			t.Fatalf("Real disagrees with result in test vector %v! %x != %x", n, vec.Out, out)
		}
	}
}
