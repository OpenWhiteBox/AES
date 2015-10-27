package xiao

import (
	"bytes"
	"testing"

	"github.com/OpenWhiteBox/AES/primitives/matrix"
	"github.com/OpenWhiteBox/AES/primitives/table"

	"github.com/OpenWhiteBox/AES/constructions/common"
	"github.com/OpenWhiteBox/AES/constructions/saes"

	test_vectors "github.com/OpenWhiteBox/AES/constructions/test"
)

func TestTBoxMixCol(t *testing.T) {
	in := [4]byte{0x4c, 0x7c, 0x84, 0xb3}
	out := [4]byte{0xcd, 0x5d, 0xa1, 0xc9}

	baseConstr := saes.Construction{}

	left := TBoxMixCol{
		[2]table.Byte{
			common.TBox{baseConstr, 0xea, 0x00},
			common.TBox{baseConstr, 0x8d, 0x00},
		},
		MixColumns,
		Left,
	}

	right := TBoxMixCol{
		[2]table.Byte{
			common.TBox{baseConstr, 0xf5, 0x00},
			common.TBox{baseConstr, 0x2f, 0x00},
		},
		MixColumns,
		Right,
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
	for n, vec := range test_vectors.AESVectors {
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
	for n, vec := range test_vectors.AESVectors {
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
