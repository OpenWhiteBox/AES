package xiao

import (
	"bytes"
	"testing"

	"github.com/OpenWhiteBox/primitives/matrix"
	"github.com/OpenWhiteBox/primitives/table"

	"github.com/OpenWhiteBox/AES/constructions/common"
	"github.com/OpenWhiteBox/AES/constructions/saes"

	test_vectors "github.com/OpenWhiteBox/AES/constructions/test"
)

var (
	key   = []byte{72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100, 33, 33, 33, 33, 33}
	seed  = []byte{38, 41, 142, 156, 29, 181, 23, 194, 21, 250, 223, 183, 210, 168, 214, 145}
	input = []byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}
)

func TestTBoxMixCol(t *testing.T) {
	in := [4]byte{0x4c, 0x7c, 0x84, 0xb3}
	out := [4]byte{0xcd, 0x5d, 0xa1, 0xc9}

	baseConstr := saes.Construction{}

	left := tBoxMixCol{
		[2]table.Byte{
			common.TBox{baseConstr, 0xea, 0x00},
			common.TBox{baseConstr, 0x8d, 0x00},
		},
		mixColumns,
		left,
	}

	right := tBoxMixCol{
		TBoxes: [2]table.Byte{
			common.TBox{baseConstr, 0xf5, 0x00},
			common.TBox{baseConstr, 0x2f, 0x00},
		},
		MixCol: mixColumns,
		Side:   right,
	}

	cand := left.Get([2]byte{in[0], in[1]})
	for k, v := range right.Get([2]byte{in[2], in[3]}) {
		cand[k] ^= v
	}

	if out != cand {
		t.Fatalf("TBoxMixCol does not calculate T-Box composed with MC!\nReal: %x\nCand: %x\n", out, cand)
	}
}

func TestEncrypt(t *testing.T) {
	for n, vec := range test_vectors.GetAESVectors(testing.Short()) {
		constr, inputMask, outputMask := GenerateEncryptionKeys(
			vec.Key, vec.Key, common.IndependentMasks{common.RandomMask, common.RandomMask},
		)

		inputInv, _ := inputMask.Invert()
		outputInv, _ := outputMask.Invert()

		in, out := make([]byte, 16), make([]byte, 16)

		copy(in, inputInv.Mul(matrix.Row(vec.In))) // Apply input encoding.

		constr.Encrypt(out, in)

		copy(out, outputInv.Mul(matrix.Row(out))) // Remove output encoding.

		if !bytes.Equal(vec.Out, out) {
			t.Fatalf("Real disagrees with result in test vector %v! %x != %x", n, vec.Out, out)
		}
	}
}

func TestDecrypt(t *testing.T) {
	for n, vec := range test_vectors.GetAESVectors(testing.Short()) {
		constr, inputMask, outputMask := GenerateDecryptionKeys(
			vec.Key, vec.Key, common.IndependentMasks{common.RandomMask, common.RandomMask},
		)

		inputInv, _ := inputMask.Invert()
		outputInv, _ := outputMask.Invert()

		in, out := make([]byte, 16), make([]byte, 16)

		copy(in, inputInv.Mul(matrix.Row(vec.Out))) // Apply input encoding.

		constr.Encrypt(out, in)

		copy(out, outputInv.Mul(matrix.Row(out))) // Remove output encoding.

		if !bytes.Equal(vec.In, out) {
			t.Fatalf("Real disagrees with result in test vector %v! %x != %x", n, vec.Out, out)
		}
	}
}

func TestPersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping the persistence test in short mode!")
	}

	constr1, _, _ := GenerateEncryptionKeys(key, seed, common.IndependentMasks{common.RandomMask, common.RandomMask})

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

func BenchmarkGenerateEncryptionKeys(b *testing.B) {
	for i := 0; i < b.N; i++ {
		constr, _, _ := GenerateEncryptionKeys(key, seed, common.IndependentMasks{common.RandomMask, common.RandomMask})
		constr.Serialize()
	}
}

// A "Live" Encryption is one based on table abstractions, so many computations are performed on-demand.
func BenchmarkLiveEncrypt(b *testing.B) {
	constr, _, _ := GenerateEncryptionKeys(key, seed, common.IndependentMasks{common.RandomMask, common.RandomMask})

	out := make([]byte, 16)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr.Encrypt(out, input)
	}
}

// A "Dead" Encryption is one based on serialized tables, like we'd have in a real use case.
func BenchmarkDeadEncrypt(b *testing.B) {
	constr1, _, _ := GenerateEncryptionKeys(key, seed, common.IndependentMasks{common.RandomMask, common.RandomMask})

	serialized := constr1.Serialize()
	constr2, _ := Parse(serialized)

	out := make([]byte, 16)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr2.Encrypt(out, input)
	}
}
