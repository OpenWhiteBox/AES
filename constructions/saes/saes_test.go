package saes

import (
	"bytes"
	"crypto/aes"
	"testing"

	test_vectors "github.com/OpenWhiteBox/AES/constructions/test"
)

var key = []byte{72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100, 33, 33, 33, 33, 33}

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
	real := [11][]byte{
		[]byte{72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100, 33, 33, 33, 33, 33},
		[]byte{180, 152, 145, 145, 219, 184, 198, 254, 169, 212, 162, 223, 136, 245, 131, 254},
		[]byte{80, 116, 42, 85, 139, 204, 236, 171, 34, 24, 78, 116, 170, 237, 205, 138},
		[]byte{1, 201, 84, 249, 138, 5, 184, 82, 168, 29, 246, 38, 2, 240, 59, 172},
		[]byte{133, 43, 197, 142, 15, 46, 125, 220, 167, 51, 139, 250, 165, 195, 176, 86},
		[]byte{187, 204, 116, 136, 180, 226, 9, 84, 19, 209, 130, 174, 182, 18, 50, 248},
		[]byte{82, 239, 53, 198, 230, 13, 60, 146, 245, 220, 190, 60, 67, 206, 140, 196},
		[]byte{153, 139, 41, 220, 127, 134, 21, 78, 138, 90, 171, 114, 201, 148, 39, 182},
		[]byte{59, 71, 103, 1, 68, 193, 114, 79, 206, 155, 217, 61, 7, 15, 254, 139},
		[]byte{86, 252, 90, 196, 18, 61, 40, 139, 220, 166, 241, 182, 219, 169, 15, 61},
		[]byte{179, 138, 125, 125, 161, 183, 85, 246, 125, 17, 164, 64, 166, 184, 171, 125},
	}

	constr := Construction{key}
	cand := constr.StretchedKey()

	for i := 0; i < 11; i++ {
		if !bytes.Equal(real[i], cand[i]) {
			t.Fatalf("Real #%v disagrees with result! %x != %x", i, real[i], cand[i])
		}
	}
}

func TestShiftRows(t *testing.T) {
	in := []byte{99, 202, 183, 4, 9, 83, 208, 81, 205, 96, 224, 231, 186, 112, 225, 140}
	out := []byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}

	constr := Construction{key}
	constr.ShiftRows(in)

	if !bytes.Equal(out, in) {
		t.Fatalf("Real disagrees with result! %x != %x", out, in)
	}
}

func TestShiftRows2(t *testing.T) {
	in := []byte{83, 146, 140, 83, 213, 138, 7, 139, 50, 163, 16, 51, 66, 55, 140, 174}
	out := []byte{83, 138, 16, 174, 213, 163, 140, 83, 50, 55, 140, 139, 66, 146, 7, 51}

	constr := Construction{key}
	constr.ShiftRows(in)

	if !bytes.Equal(out, in) {
		t.Fatalf("Real disagrees with result! %x != %x", out, in)
	}
}

func TestMixColumns(t *testing.T) {
	in := []byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}
	out := []byte{95, 114, 100, 21, 87, 245, 188, 146, 247, 190, 59, 41, 29, 185, 249, 26}

	constr := Construction{key}
	constr.MixColumns(in)

	if !bytes.Equal(out, in) {
		t.Fatalf("Real disagrees with result! %x != %x", out, in)
	}
}

func TestMixColumns2(t *testing.T) {
	in := [][]byte{
		[]byte{83, 138, 16, 174},
		[]byte{213, 163, 140, 83},
		[]byte{50, 55, 140, 139},
		[]byte{66, 146, 7, 51},
	}

	out := [][]byte{
		[]byte{157, 194, 16, 40},
		[]byte{144, 84, 128, 237},
		[]byte{58, 88, 128, 224},
		[]byte{29, 71, 139, 53},
	}

	constr := Construction{key}

	for i := 0; i < 4; i++ {
		constr.MixColumn(in[i])

		if !bytes.Equal(out[i], in[i]) {
			t.Fatalf("Column %v was mixed wrong! %x != %x", i, out[i], in[i])
		}
	}
}

func TestMixColumns3(t *testing.T) {
	in := []byte{8, 146, 217, 164, 165, 74, 237, 190, 128, 48, 179, 171, 55, 194, 224, 46}
	out := []byte{192, 227, 196, 0, 220, 163, 247, 52, 83, 133, 43, 85, 253, 189, 92, 39}

	constr := Construction{key}
	constr.MixColumns(in)

	if !bytes.Equal(out, in) {
		t.Fatalf("Real disagrees with result! %x != %x", out, in)
	}
}

func TestEncrypt(t *testing.T) {
	for n, vec := range test_vectors.AESVectors {
		constr := Construction{vec.Key}

		cand := make([]byte, 16)
		constr.Encrypt(cand, vec.In)

		if !bytes.Equal(vec.Out, cand) {
			t.Fatalf("Real disagrees with result in test vector %v! %x != %x", n, vec.Out, cand)
		}
	}
}

func BenchmarkStandardEncrypt(b *testing.B) {
	key := test_vectors.AESVectors[50].Key
	input := test_vectors.AESVectors[50].In

	constr := Construction{key}
	out := make([]byte, 16)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr.Encrypt(out, input)
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
