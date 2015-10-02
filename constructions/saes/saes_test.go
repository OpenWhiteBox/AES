package saes

import (
	"crypto/aes"
	"testing"

	test_vectors "github.com/OpenWhiteBox/AES/constructions/test"
)

var key = [16]byte{72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100, 33, 33, 33, 33, 33}

func TestSubByte(t *testing.T) {
	constr := Construction{key}

	if constr.SubByte(0x00) != 0x63 {
		t.Fatalf("Affine component of SubByte is wrong!")
	}

	if constr.SubByte(0x53) != 0xED {
		t.Fatalf("Linear component of SubByte is wrong! #1")
	}

	if constr.SubByte(0x02) != 0x77 {
		t.Fatalf("Linear component of SubByte is wrong! #2")
	}

	// Test subWord
	wordA := uint32((0x10 << 24) | (0x53 << 16) | (0x86 << 8) | 0xED)
	wordB := uint32((0xCA << 24) | (0xED << 16) | (0x44 << 8) | 0x55)

	if constr.SubWord(wordA) != wordB {
		t.Fatalf("constr.subWord gave incorrect output!")
	}
}

func TestKeyStretching(t *testing.T) {
	real := [11][16]byte{
		[16]byte{72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100, 33, 33, 33, 33, 33},
		[16]byte{180, 152, 145, 145, 219, 184, 198, 254, 169, 212, 162, 223, 136, 245, 131, 254},
		[16]byte{80, 116, 42, 85, 139, 204, 236, 171, 34, 24, 78, 116, 170, 237, 205, 138},
		[16]byte{1, 201, 84, 249, 138, 5, 184, 82, 168, 29, 246, 38, 2, 240, 59, 172},
		[16]byte{133, 43, 197, 142, 15, 46, 125, 220, 167, 51, 139, 250, 165, 195, 176, 86},
		[16]byte{187, 204, 116, 136, 180, 226, 9, 84, 19, 209, 130, 174, 182, 18, 50, 248},
		[16]byte{82, 239, 53, 198, 230, 13, 60, 146, 245, 220, 190, 60, 67, 206, 140, 196},
		[16]byte{153, 139, 41, 220, 127, 134, 21, 78, 138, 90, 171, 114, 201, 148, 39, 182},
		[16]byte{59, 71, 103, 1, 68, 193, 114, 79, 206, 155, 217, 61, 7, 15, 254, 139},
		[16]byte{86, 252, 90, 196, 18, 61, 40, 139, 220, 166, 241, 182, 219, 169, 15, 61},
		[16]byte{179, 138, 125, 125, 161, 183, 85, 246, 125, 17, 164, 64, 166, 184, 171, 125},
	}

	constr := Construction{key}
	cand := constr.StretchedKey()

	for i := 0; i < 11; i++ {
		if real[i] != cand[i] {
			t.Fatalf("Real #%v disagrees with result! %x != %x", i, real[i], cand[i])
		}
	}
}

func TestShiftRows(t *testing.T) {
	in := [16]byte{99, 202, 183, 4, 9, 83, 208, 81, 205, 96, 224, 231, 186, 112, 225, 140}
	out := [16]byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}

	constr := Construction{key}
	cand := constr.ShiftRows(in)

	if out != cand {
		t.Fatalf("Real disagrees with result! %x != %x", out, cand)
	}
}

func TestMixColumns(t *testing.T) {
	in := [16]byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}
	out := [16]byte{95, 114, 100, 21, 87, 245, 188, 146, 247, 190, 59, 41, 29, 185, 249, 26}

	constr := Construction{key}
	cand := constr.MixColumns(in)

	if out != cand {
		t.Fatalf("Real disagrees with result! %x != %x", out, cand)
	}
}

func TestEncrypt(t *testing.T) {
	for n, vec := range test_vectors.AESVectors {
		constr := Construction{vec.Key}
		cand := constr.Encrypt(vec.In)

		if vec.Out != cand {
			t.Fatalf("Real disagrees with result in test vector %v! %x != %x", n, vec.Out, cand)
		}
	}
}

func BenchmarkStandardEncrypt(b *testing.B) {
	key := test_vectors.AESVectors[50].Key
	input := test_vectors.AESVectors[50].In

	constr := Construction{key}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr.Encrypt(input)
	}
}

func BenchmarkGolangEncrypt(b *testing.B) {
	key := test_vectors.AESVectors[50].Key
	input := test_vectors.AESVectors[50].In

	constr, _ := aes.NewCipher(key[:])
	out := make([]byte, 16)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr.Encrypt(out, input[:])
	}
}
