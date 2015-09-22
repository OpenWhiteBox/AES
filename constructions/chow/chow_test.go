package chow

import (
	test_vectors "../test/"
	"testing"
)

var key = GenerateTables([16]byte{72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100, 33, 33, 33, 33, 33})

func TestShiftRows(t *testing.T) {
	in := [16]byte{99, 202, 183, 4, 9, 83, 208, 81, 205, 96, 224, 231, 186, 112, 225, 140}
	out := [16]byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}

	constr := Construction{key}
	cand := constr.shiftRows(in)

	for i := 0; i < 16; i++ {
		if out[i] != cand[i] {
			t.Fatalf("Byte %v is wrong! %v != %v", i, out[i], cand[i])
		}
	}
}

func TestMixColumns(t *testing.T) {
	in := [16]byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}
	out := [16]byte{95, 114, 100, 21, 87, 245, 188, 146, 247, 190, 59, 41, 29, 185, 249, 26}

	constr := Construction{key}
	cand := constr.mixColumns(in)

	for i := 0; i < 16; i++ {
		if out[i] != cand[i] {
			t.Fatalf("Byte %v is wrong! %v != %v", i, out[i], cand[i])
		}
	}
}

func TestEncrypt(t *testing.T) {
	for n, vec := range test_vectors.AESVectors {
		table := GenerateTables(vec.Key)
		constr := Construction{table}
		cand := constr.Encrypt(vec.In)

		for i := 0; i < 16; i++ {
			if vec.Out[i] != cand[i] {
				t.Fatalf("Byte %v is wrong in test vector %v! %v != %v", i, n, vec.Out[i], cand[i])
			}
		}
	}
}
