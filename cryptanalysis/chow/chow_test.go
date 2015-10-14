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

func TestRecoverKey(t *testing.T) {
	constr := fastTestConstruction()
	fmt.Println(RecoverKey(constr))
}

func TestFindBasisAndSort(t *testing.T) {
	constr := fastTestConstruction()
	S := GenerateS(constr, 0, 0)

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

	S := GenerateS(constr, 0, 0)
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
	m, c := DecomposeAffineEncoding(cand)
	reconstr := encoding.ByteAffine{encoding.ByteLinear(m), c}

	for i := 0; i < 256; i++ {
		a, b := cand.Encode(byte(i)), reconstr.Encode(byte(i))

		if a != b {
			t.Fatalf("Identity is broken! f = Q^(-1) <- Qtilde isn't affine! %v != %v at point %v", a, b, i)
		}
	}
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

func TestMakeAffineRound(t *testing.T) {
	constr, _ := testConstruction()
	fastConstr := fastTestConstruction()

	// Remove non-linear parts of input and output encodings on round 1.
	idEnc := [16]encoding.Byte{}
	for i, _ := range idEnc {
		idEnc[i] = encoding.IdentityByte{}
	}

	r0Enc, _ := MakeAffineRound(fastConstr, idEnc, 0)
	r1Enc, r1 := MakeAffineRound(fastConstr, r0Enc, 1)

	// Test that remaining encodings are affine.
	for i := 0; i < 16; i++ { // for i := 0; i < 16; i++ {
		// Actual input and output encoding for round 1 in position i.
		in := constr.TBoxTyiTable[1][i/4*4].(encoding.WordTable).In
		out := constr.TBoxTyiTable[2][shiftRows(i)].(encoding.WordTable).In

		f := F{constr, 1, i, 0x00}

		// R corresponds to an exposed round.
		// Marginally more intuitive way to write it: V(x) := out.Decode(f.Get(in.Encode(x)))
		R := table.InvertibleTable(encoding.ByteTable{
			encoding.InverseByte{in},
			encoding.InverseByte{out},
			f,
		})

		// Expected affine input and output encodings for round 1 in position i after extraction.
		inAff := encoding.ComposedBytes{in, encoding.InverseByte{r0Enc[i/4*4]}}
		outAff := encoding.ComposedBytes{out, encoding.InverseByte{r1Enc[shiftRows(i)]}}

		m, c := DecomposeAffineEncoding(inAff)
		inAffTest := encoding.ByteAffine{encoding.ByteLinear(m), c}

		n, d := DecomposeAffineEncoding(outAff)
		outAffTest := encoding.ByteAffine{encoding.ByteLinear(n), d}

		// Verify that expected encodings are affine.
		for j := 0; j < 256; j++ {
			a, b := inAff.Encode(byte(j)), inAffTest.Encode(byte(j))
			if a != b {
				t.Fatalf("inAff for position %v is not affine! %v != %v at point %v", i, a, b, j)
			}

			c, d := outAff.Encode(byte(j)), outAffTest.Encode(byte(j))
			if c != d {
				t.Fatalf("outAff for position %v is not affine! %v != %v at point %v", i, c, d, j)
			}
		}

		// Verify that r1 is R with affine input and output encodings.
		RAff := encoding.ByteTable{inAff, outAff, R}

		for j := 0; j < 256; j++ {
			a, b := RAff.Get(byte(j)), r1[shiftRows(i)].Get(byte(j))
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

	if FindCharacteristic(M) == FindCharacteristic(N) {
		t.Fatalf("FindCharacteristic was not invariant!\nM = %x\nA = %x\nN = %x, ", M, A, N)
	}
}

func BenchmarkGenerateS(b *testing.B) {
	constr := fastTestConstruction()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GenerateS(constr, 0, 0)
	}
}

func BenchmarkQtilde(b *testing.B) {
	constr := fastTestConstruction()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		S := GenerateS(constr, 0, 0)
		_ = FindBasisAndSort(S)
	}
}
