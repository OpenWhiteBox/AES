package chow

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/OpenWhiteBox/AES/constructions/chow"
	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/matrix"
	"github.com/OpenWhiteBox/AES/primitives/table"

	test_vectors "github.com/OpenWhiteBox/AES/constructions/test"
)

var (
	isCached bool = false
	cached   chow.Construction
)

func testConstruction() (chow.Construction, []byte) {
	key := test_vectors.AESVectors[50].Key
	seed := test_vectors.AESVectors[51].Key
	constr, _, _ := chow.GenerateKeys(key, seed, chow.SameMasks(chow.IdentityMask))

	return constr, key
}

func fastTestConstruction() chow.Construction {
	if !isCached {
		constr1, _ := testConstruction()
		serialized := constr1.Serialize()
		cached, _ = chow.Parse(serialized)

		isCached = true
	}

	return cached
}

func exposeRound(constr chow.Construction, round, inPos, outPos int) (encoding.Byte, encoding.Byte, table.InvertibleTable){
	// Actual input and output encoding for round 1 in position i.
	in := constr.TBoxTyiTable[round][inPos].(encoding.WordTable).In
	out := constr.TBoxTyiTable[round+1][shiftRows(outPos)].(encoding.WordTable).In

	f := F{constr, round, inPos, outPos, 0x00}

	// R corresponds to an exposed round.
	// Marginally more intuitive way to write it: V(x) := out.Decode(f.Get(in.Encode(x)))
	R := table.InvertibleTable(encoding.ByteTable{
		encoding.InverseByte{in},
		encoding.InverseByte{out},
		f,
	})

	return in, out, R
}

func verifyIsAffine(t *testing.T, aff encoding.Byte, err string) {
	m, c := DecomposeAffineEncoding(aff)
	test := encoding.ByteAffine{encoding.ByteLinear(m), c}

	for i := 0; i < 256; i++ {
		a, b := aff.Encode(byte(i)), test.Encode(byte(i))
		if a != b {
			t.Fatalf(err, i, a, b)
		}
	}
}

func TestFindBasisAndSort(t *testing.T) {
	constr := fastTestConstruction()
	S := GenerateS(constr, 0, 0, 0)

	basis := FindBasisAndSort(S)

	// Test that each vector is correct in place.
	for i, perm := range S {
		vect := [256]byte{}
		copy(vect[:], table.SerializeByte(FunctionFromBasis(i, basis)))

		if vect != perm {
			t.Fatalf("FindBasisAndSort, position #%v is wrong!\n%v\n%v", i, vect, perm)
		}
	}

	// Test composition of two functions.
	x := table.SerializeByte(table.ComposedBytes{
		table.ParsedByte(S[39][:]),
		table.ParsedByte(S[120][:]),
	})

	y := [256]byte{}
	copy(y[:], x)

	if y != S[39^120] {
		t.Fatalf("Function composition was wrong!\n%v\n%v", x, S[39^120])
	}
}

func TestQtilde(t *testing.T) {
	constr, _ := testConstruction()

	S := GenerateS(constr, 0, 0, 0)
	_ = FindBasisAndSort(S)

	qtilde := Qtilde{S}

	for i := 0; i < 256; i++ { // Test first identity, that f = Qtilde <- ...
		cand := qtilde.Encode(qtilde.Decode(byte(i)) ^ 37)

		if S[37][i] != cand {
			t.Fatalf("Identity broken! f = Qtilde <- xor psi <- Qtilde^(-1)\nf(%v) = %v, g(%v) = %v", i, S[37][i], i, cand)
		}
	}

	// Test second identity that Qtilde -> Qinverse is an affine transformation.
	q := constr.TBoxTyiTable[1][0].(encoding.WordTable).In // Actual output encoding at round 0, position 0.
	cand := encoding.ComposedBytes{qtilde, encoding.InverseByte{q}}

	// Decompose cand as an affine encoding and reconstruct it.  If cand is affine, the two will agree exactly.
	verifyIsAffine(t, cand, "Identity is broken! f = Q^(-1) <- Qtilde isn't affine! At point %v, %v != %v.")
}

func TestDecomposeAffineEncoding(t *testing.T) {
	ae := encoding.ByteAffine{
		encoding.ByteLinear(matrix.Matrix{
			matrix.Row{0xA4},
			matrix.Row{0x49},
			matrix.Row{0x92},
			matrix.Row{0x25},
			matrix.Row{0x4a},
			matrix.Row{0x94},
			matrix.Row{0x29},
			matrix.Row{0x52},
		}),
		0x63,
	}

	m, c := DecomposeAffineEncoding(ae)

	for i := 0; i < 8; i++ {
		if ae.Linear[i][0] != m[i][0] {
			t.Fatalf("Row %v in the linear part is wrong! %v != %v", i, ae.Linear[i][0], m[i][0])
		}
	}

	if ae.Affine != c {
		t.Fatalf("The affine part is wrong! %v != %v", ae.Affine, c)
	}
}

func TestRecoverAffineEncoded(t *testing.T) {
	constr, _ := testConstruction()
	fastConstr := fastTestConstruction()

	inEncTilde := [4]encoding.Byte{} // Small amount of caching. Makes test go from 10s to 7s.

	for i := 0; i < 16; i++ {
		in, out, R := exposeRound(constr, 1, i/4*4, i)

		if i/4*4 == i {
			inEncTilde[i/4], _ = RecoverAffineEncoded(fastConstr, encoding.IdentityByte{}, 0, i, i)
		}
		outEncTilde, outTilde := RecoverAffineEncoded(fastConstr, inEncTilde[i/4], 1, i/4*4, i)

		// Expected affine input and output encodings for round 1 in position i after extraction.
		inAff := encoding.ComposedBytes{in, encoding.InverseByte{inEncTilde[i/4]}}
		outAff := encoding.ComposedBytes{out, encoding.InverseByte{outEncTilde}}

		// Verify that expected encodings are affine.
		verifyIsAffine(t, inAff, fmt.Sprintf("inAff for position %v is not affine! At point %%v, %%v != %%v.", i))
		verifyIsAffine(t, outAff, fmt.Sprintf("outAff for position %v is not affine! At point %%v, %%v != %%v.", i))

		// Verify that outTilde is R with affine input and output encodings.
		RAff := encoding.ByteTable{inAff, outAff, R}

		for j := 0; j < 256; j++ {
			a, b := RAff.Get(byte(j)), outTilde.Get(byte(j))
			if a != b {
				t.Fatalf("RAff and r1 at position %v disagree at point %v! %v != %v", i, j, a, b)
			}
		}
	}
}

func TestFindCharacteristic(t *testing.T) {
	M := matrix.GenerateRandom(rand.Reader, 8)

	A := matrix.GenerateRandom(rand.Reader, 8)
	AInv, _ := A.Invert()

	N, _ := DecomposeAffineEncoding(encoding.ComposedBytes{
		encoding.ByteLinear(A),
		encoding.ByteLinear(M),
		encoding.ByteLinear(AInv),
	})

	if FindCharacteristic(M) != FindCharacteristic(N) {
		t.Fatalf("FindCharacteristic was not invariant!\nM = %x\nA = %x\nN = %x, ", M, A, N)
	}
}

func BenchmarkGenerateS(b *testing.B) {
	constr := fastTestConstruction()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GenerateS(constr, 0, 0, 0)
	}
}

func BenchmarkQtilde(b *testing.B) {
	constr := fastTestConstruction()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		S := GenerateS(constr, 0, 0, 0)
		_ = FindBasisAndSort(S)
	}
}
