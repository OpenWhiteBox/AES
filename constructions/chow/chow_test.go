package chow

import (
	"testing"

	"github.com/OpenWhiteBox/AES/primitives/matrix"

	test_vectors "github.com/OpenWhiteBox/AES/constructions/test"
)

var key = [16]byte{72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100, 33, 33, 33, 33, 33}

func TestShiftRows(t *testing.T) {
	in := [16]byte{99, 202, 183, 4, 9, 83, 208, 81, 205, 96, 224, 231, 186, 112, 225, 140}
	out := [16]byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}

	constr, _, _ := GenerateKeys(key, key)
	cand := constr.ShiftRows(in)

	if out != cand {
		t.Fatalf("Real disagrees with result! %v != %v", out, cand)
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

func TestEncrypt(t *testing.T) {
	for n, vec := range test_vectors.AESVectors[0:10] {
		constr, input, output := GenerateKeys(vec.Key, vec.Key)

		inputInv, _ := input.Invert()
		outputInv, _ := output.Invert()

		in, out := [16]byte{}, [16]byte{}

		copy(in[:], inputInv.Mul(matrix.Row(vec.In[:]))) // Apply input encoding.

		cand := constr.Encrypt(in)

		copy(out[:], outputInv.Mul(matrix.Row(cand[:]))) // Remove output encoding.

		if vec.Out != out {
			t.Fatalf("Real disagrees with result in test vector %v! %x != %x", n, vec.Out, cand)
		}
	}
}

func TestPersistence(t *testing.T) {
	key := test_vectors.AESVectors[50].Key
	seed := test_vectors.AESVectors[51].Key
	input := test_vectors.AESVectors[50].In

	constr, _, _ := GenerateKeys(key, seed)

	serialized := constr.Serialize()
	constr2 := Parse(serialized)

	cand1 := constr.Encrypt(input)
	cand2 := constr2.Encrypt(input)

	if cand1 != cand2 {
		t.Fatalf("Real disagrees with parsed! %v != %v", cand1, cand2)
	}
}

// A "Live" Encryption is one based on table abstractions, so many computations are performed on-demand.
func BenchmarkLiveEncrypt(b *testing.B) {
	key := test_vectors.AESVectors[50].Key
	seed := test_vectors.AESVectors[51].Key
	input := test_vectors.AESVectors[50].In

	constr, _, _ := GenerateKeys(key, seed)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr.Encrypt(input)
	}
}

// A "Dead" Encryption is one based on serialized tables, like we'd have in a real use case.
func BenchmarkDeadEncrypt(b *testing.B) {
	key := test_vectors.AESVectors[50].Key
	seed := test_vectors.AESVectors[51].Key
	input := test_vectors.AESVectors[50].In

	constr, _, _ := GenerateKeys(key, seed)
	serialized := constr.Serialize()
	constr2 := Parse(serialized)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr2.Encrypt(input)
	}
}
