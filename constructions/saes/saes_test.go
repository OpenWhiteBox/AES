package saes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"

	"fmt"
	"testing"

	test_vectors "github.com/OpenWhiteBox/AES/constructions/test"
)

var key = []byte{72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100, 33, 33, 33, 33, 33}

func Example_encrypt() {
	constr := Construction{
		Key: []byte{100, 17, 10, 146, 79, 7, 67, 213, 0, 204, 173, 174, 114, 193, 52, 39},
	}

	src := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	dst := make([]byte, 16)

	constr.Encrypt(dst, src)

	fmt.Println(dst)
	// Output: [53 135 12 106 87 233 233 35 20 188 184 8 124 222 114 206]
}

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

func TestUnSubByte(t *testing.T) {
	constr := Construction{key}

	for i := 0; i < 256; i++ {
		cand := constr.UnSubByte(constr.SubByte(byte(i)))
		if cand != byte(i) {
			t.Fatalf("UnSubByte didn't properly invert SubByte! %v in, got %v out.", i, cand)
		}
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

func TestUnShiftRows(t *testing.T) {
	in := []byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}
	out := []byte{99, 202, 183, 4, 9, 83, 208, 81, 205, 96, 224, 231, 186, 112, 225, 140}

	constr := Construction{key}
	constr.UnShiftRows(in)

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

func TestUnMixColumns(t *testing.T) {
	in := []byte{95, 114, 100, 21, 87, 245, 188, 146, 247, 190, 59, 41, 29, 185, 249, 26}
	out := []byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}

	constr := Construction{key}
	constr.UnMixColumns(in)

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

func TestDecrypt(t *testing.T) {
	for n, vec := range test_vectors.AESVectors {
		constr := Construction{vec.Key}

		cand := make([]byte, 16)
		constr.Decrypt(cand, vec.Out)

		if !bytes.Equal(vec.In, cand) {
			t.Fatalf("Real disagrees with result in test vector %v! %x != %x", n, vec.In, cand)
		}
	}
}

func TestCBC(t *testing.T) {
	// Vector stolen from crypto/aes/cbc_aes_test.go
	key := []byte{0x2b, 0x7e, 0x15, 0x16, 0x28, 0xae, 0xd2, 0xa6, 0xab, 0xf7, 0x15, 0x88, 0x09, 0xcf, 0x4f, 0x3c}
	iv := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

	in := []byte{
		0x6b, 0xc1, 0xbe, 0xe2, 0x2e, 0x40, 0x9f, 0x96, 0xe9, 0x3d, 0x7e, 0x11, 0x73, 0x93, 0x17, 0x2a,
		0xae, 0x2d, 0x8a, 0x57, 0x1e, 0x03, 0xac, 0x9c, 0x9e, 0xb7, 0x6f, 0xac, 0x45, 0xaf, 0x8e, 0x51,
		0x30, 0xc8, 0x1c, 0x46, 0xa3, 0x5c, 0xe4, 0x11, 0xe5, 0xfb, 0xc1, 0x19, 0x1a, 0x0a, 0x52, 0xef,
		0xf6, 0x9f, 0x24, 0x45, 0xdf, 0x4f, 0x9b, 0x17, 0xad, 0x2b, 0x41, 0x7b, 0xe6, 0x6c, 0x37, 0x10,
	}

	out := []byte{
		0x76, 0x49, 0xab, 0xac, 0x81, 0x19, 0xb2, 0x46, 0xce, 0xe9, 0x8e, 0x9b, 0x12, 0xe9, 0x19, 0x7d,
		0x50, 0x86, 0xcb, 0x9b, 0x50, 0x72, 0x19, 0xee, 0x95, 0xdb, 0x11, 0x3a, 0x91, 0x76, 0x78, 0xb2,
		0x73, 0xbe, 0xd6, 0xb8, 0xe3, 0xc1, 0x74, 0x3b, 0x71, 0x16, 0xe6, 0x9e, 0x22, 0x22, 0x95, 0x16,
		0x3f, 0xf1, 0xca, 0xa1, 0x68, 0x1f, 0xac, 0x09, 0x12, 0x0e, 0xca, 0x30, 0x75, 0x86, 0xe1, 0xa7,
	}

	constr := Construction{key}

	cbcEnc := cipher.NewCBCEncrypter(constr, iv)
	cbcDec := cipher.NewCBCDecrypter(constr, iv)

	cand1, cand2 := make([]byte, 16*4), make([]byte, 16*4)

	cbcEnc.CryptBlocks(cand1, in)
	cbcDec.CryptBlocks(cand2, out)

	if !bytes.Equal(cand2, in) {
		t.Fatalf("CBC decryption was wrong!")
	}

	if !bytes.Equal(cand1, out) {
		t.Fatalf("CBC encryption was wrong!")
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

func BenchmarkStandardDecrypt(b *testing.B) {
	key := test_vectors.AESVectors[50].Key
	input := test_vectors.AESVectors[50].Out

	constr := Construction{key}
	out := make([]byte, 16)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr.Decrypt(out, input)
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
