package chow

import (
	"bytes"
	"crypto/aes"
	"testing"

	"github.com/OpenWhiteBox/AES/primitives/matrix"

	test_vectors "github.com/OpenWhiteBox/AES/constructions/test"
)

var (
	key   = []byte{72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100, 33, 33, 33, 33, 33}
	seed  = []byte{38, 41, 142, 156, 29, 181, 23, 194, 21, 250, 223, 183, 210, 168, 214, 145}
	input = []byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}
)

func TestShiftRows(t *testing.T) {
	in := []byte{99, 202, 183, 4, 9, 83, 208, 81, 205, 96, 224, 231, 186, 112, 225, 140}
	out := []byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}

	constr, _, _ := GenerateKeys(key, key, SameMasks(IdentityMask))
	constr.ShiftRows(in)

	if !bytes.Equal(out, in) {
		t.Fatalf("Real disagrees with result! %v != %v", out, in)
	}
}

func TestTyiTable(t *testing.T) {
	in := [16]byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}
	out := [16]byte{95, 114, 100, 21, 87, 245, 188, 146, 247, 190, 59, 41, 29, 185, 249, 26}

	a, b, c, d := TyiTable(0), TyiTable(1), TyiTable(2), TyiTable(3)
	cand := [16]byte{}

	for i := 0; i < 16; i += 4 {
		e, f, g, h := a.Get(in[i+0]), b.Get(in[i+1]), c.Get(in[i+2]), d.Get(in[i+3])

		cand[i+0] = e[0] ^ f[0] ^ g[0] ^ h[0]
		cand[i+1] = e[1] ^ f[1] ^ g[1] ^ h[1]
		cand[i+2] = e[2] ^ f[2] ^ g[2] ^ h[2]
		cand[i+3] = e[3] ^ f[3] ^ g[3] ^ h[3]
	}

	if out != cand {
		t.Fatalf("Real disagrees with result! %v != %v", out, cand)
	}
}

func TestUnmaskedEncrypt(t *testing.T) {
	cand, real := make([]byte, 16), make([]byte, 16)

	// Calculate the candidate output.
	constr, _, _ := GenerateKeys(key, seed, SameMasks(IdentityMask))
	constr.Encrypt(cand, input)

	// Calculate the real output.
	c, _ := aes.NewCipher(key)
	c.Encrypt(real, input)

	if !bytes.Equal(real, cand) {
		t.Fatalf("Real disagrees with result! %x != %x", real, cand)
	}
}

func TestMatchedEncrypt(t *testing.T) {
	cand, real := make([]byte, 16), make([]byte, 16)

	// Calculate the candidate output.
	constr, inputMask, outputMask := GenerateKeys(key, seed, MatchingMasks{})

	inputInv, _ := inputMask.Invert()
	outputInv, _ := outputMask.Invert()

	in := make([]byte, 16)
	copy(in, inputInv.Mul(matrix.Row(input))) // Apply input encoding.

	constr.Encrypt(cand, in)
	constr.Encrypt(cand, cand)

	copy(cand, outputInv.Mul(matrix.Row(cand))) // Remove output encoding.

	// Calculate the real output.
	c, _ := aes.NewCipher(key)
	c.Encrypt(real, input)
	c.Encrypt(real, real)

	if !bytes.Equal(real, cand) {
		t.Fatalf("Real disagrees with result! %x != %x", real, cand)
	}
}

func TestEncrypt(t *testing.T) {
	for n, vec := range test_vectors.AESVectors[0:10] {
		constr, inputMask, outputMask := GenerateKeys(vec.Key, vec.Key, IndependentMasks{RandomMask, RandomMask})

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

func TestPersistence(t *testing.T) {
	constr1, _, _ := GenerateKeys(key, seed, IndependentMasks{RandomMask, RandomMask})

	serialized := constr1.Serialize()
	constr2 := Parse(serialized)

	cand1, cand2 := make([]byte, 16), make([]byte, 16)

	constr1.Encrypt(cand1, input)
	constr2.Encrypt(cand2, input)

	if !bytes.Equal(cand1, cand2) {
		t.Fatalf("Real disagrees with parsed! %v != %v", cand1, cand2)
	}
}

// A "Live" Encryption is one based on table abstractions, so many computations are performed on-demand.
func BenchmarkLiveEncrypt(b *testing.B) {
	constr, _, _ := GenerateKeys(key, seed, IndependentMasks{RandomMask, RandomMask})

	out := make([]byte, 16)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr.Encrypt(out, input)
	}
}

// A "Dead" Encryption is one based on serialized tables, like we'd have in a real use case.
func BenchmarkDeadEncrypt(b *testing.B) {
	constr1, _, _ := GenerateKeys(key, seed, IndependentMasks{RandomMask, RandomMask})

	serialized := constr1.Serialize()
	constr2 := Parse(serialized)

	out := make([]byte, 16)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr2.Encrypt(out, input)
	}
}

func BenchmarkShiftRows(b *testing.B) {
	constr, _, _ := GenerateKeys(key, seed, SameMasks(IdentityMask))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr.ShiftRows(input)
	}
}

func BenchmarkExpandWord(b *testing.B) {
	constr1, _, _ := GenerateKeys(key, seed, SameMasks(IdentityMask))

	serialized := constr1.Serialize()
	constr2 := Parse(serialized)

	dst := make([]byte, 16)
	copy(dst, input)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr2.ExpandWord(constr2.TBoxTyiTable[0][0:4], dst[0:4])
	}
}

func BenchmarkExpandBlock(b *testing.B) {
	constr1, _, _ := GenerateKeys(key, seed, IndependentMasks{RandomMask, RandomMask})

	serialized := constr1.Serialize()
	constr2 := Parse(serialized)

	dst := make([]byte, 16)
	copy(dst, input)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr2.ExpandBlock(constr2.InputMask, dst)
	}
}

func BenchmarkSquashWords(b *testing.B) {
	constr1, _, _ := GenerateKeys(key, seed, SameMasks(IdentityMask))

	serialized := constr1.Serialize()
	constr2 := Parse(serialized)

	dst := make([]byte, 16)
	copy(dst, input)

	stretched := constr2.ExpandWord(constr2.TBoxTyiTable[0][0:4], dst[0:4])

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr2.SquashWords(constr2.HighXORTable[0][0:8], stretched, dst[0:4])
		copy(dst[0:4], input)
	}
}

func BenchmarkSquashBlocks(b *testing.B) {
	constr1, _, _ := GenerateKeys(key, seed, IndependentMasks{RandomMask, RandomMask})

	serialized := constr1.Serialize()
	constr2 := Parse(serialized)

	dst := make([]byte, 16)
	copy(dst, input)

	stretched := constr2.ExpandBlock(constr2.InputMask, dst)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr2.SquashBlocks(constr2.InputXORTable, stretched, dst)
		copy(dst, input)
	}
}
