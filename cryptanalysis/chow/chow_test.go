package chow

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/matrix"
	"github.com/OpenWhiteBox/AES/primitives/number"
	"github.com/OpenWhiteBox/AES/primitives/table"

	"github.com/OpenWhiteBox/AES/constructions/chow"
	"github.com/OpenWhiteBox/AES/constructions/saes"

	test_vectors "github.com/OpenWhiteBox/AES/constructions/test"
)

var (
	isCached bool = false
	cached   chow.Construction
)

func testConstruction() (chow.Construction, []byte) {
	key := test_vectors.AESVectors[50].Key
	seed := test_vectors.AESVectors[51].Key
	constr, _, _ := chow.GenerateEncryptionKeys(key, seed, chow.SameMasks(chow.IdentityMask))

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

func exposeRound(constr chow.Construction, round, inPos, outPos int) (encoding.Byte, encoding.Byte, table.InvertibleTable) {
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

func getOutputAffineEncoding(constr chow.Construction, fastConstr chow.Construction, round, pos int) encoding.ByteAffine {
	_, out, _ := exposeRound(constr, round, pos, pos)
	outEncTilde, _ := RecoverAffineEncoded(fastConstr, encoding.IdentityByte{}, round, pos, pos)
	outAff := encoding.ComposedBytes{out, encoding.InverseByte{outEncTilde}}
	A, b := DecomposeAffineEncoding(outAff)

	return encoding.ByteAffine{encoding.ByteLinear(A), b}
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

	if ae.Constant != c {
		t.Fatalf("The affine part is wrong! %v != %v", ae.Constant, c)
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
		t.Fatalf("FindCharacteristic was not invariant!\nM = %2.2x\nA = %2.2x\nN = %2.2x, ", M, A, N)
	}
}

func TestRecoverL(t *testing.T) {
	MC := [][]number.ByteFieldElem{
		[]number.ByteFieldElem{0x02, 0x01, 0x01, 0x03},
		[]number.ByteFieldElem{0x03, 0x02, 0x01, 0x01},
	}

	constr, _ := testConstruction()
	fastConstr := fastTestConstruction()

	for i := 0; i < 16; i++ {
		L := RecoverL(fastConstr, 1, i)

		outAff := getOutputAffineEncoding(constr, fastConstr, 1, i)

		// L is supposed to equal A_i <- D(beta) <- A_i^(-1)
		// We strip the conjugation by A_i and check that D(beta) is multiplication by an element of GF(2^8).
		DEnc := encoding.ComposedBytes{
			outAff,
			encoding.ByteLinear(L),
			encoding.InverseByte{outAff},
		}
		D, _ := DecomposeAffineEncoding(DEnc)
		Dstr := fmt.Sprintf("%x", D)

		// Calculate what beta should be.
		pos0, pos1 := i%4, (i+1)%4
		beta := MC[0][pos0].Mul(MC[1][pos1]).Mul(MC[0][pos1].Mul(MC[1][pos0]).Invert())

		// Calculate the matrix of multiplication by beta and check that it equals what we derived in D.
		E, _ := DecomposeAffineEncoding(encoding.ByteMultiplication(beta))
		Estr := fmt.Sprintf("%x", E)

		if Dstr != Estr {
			t.Fatalf("A_i^(-1) * L * A_i doesn't equal D(beta)! i = %v\nL = %2.2x\nD = %2.2x\nE = %2.2x\n", i, L, D, E)
		}
	}
}

func TestFindAtilde(t *testing.T) {
	fastConstr := fastTestConstruction()

	L := RecoverL(fastConstr, 1, 0)
	Atilde := FindAtilde(fastConstr, L)

	beta := CharToBeta[FindCharacteristic(L)]
	D, _ := DecomposeAffineEncoding(encoding.ByteMultiplication(beta))

	left, right := L.Compose(Atilde), Atilde.Compose(D)

	for i, _ := range left {
		if len(left[i]) != 1 || len(right[i]) != 1 || left[i][0] != right[i][0] {
			t.Fatalf("L * Atilde != Atilde * D(beta)!\nL = %x\nR = %x\n", left, right)
		}
	}
}

func TestRecoverEncodings(t *testing.T) {
	constr, key := testConstruction()
	fastConstr := fastTestConstruction()

	baseConstr := saes.Construction{key}
	roundKeys := baseConstr.StretchedKey()

	outAff := getOutputAffineEncoding(constr, fastConstr, 1, 0)

	// Manually recover the output encoding.
	Q, Ps := RecoverEncodings(fastConstr, 1, 0)

	if fmt.Sprintf("%x %v", outAff.Linear, outAff.Constant) != fmt.Sprintf("%x %v", Q.Linear, Q.Constant) {
		t.Fatalf("RecoverEncodings recovered the wrong output encoding!")
	}

	// Verify that all Ps composed with their corresponding output encoding equals XOR by a key byte.
	id := matrix.GenerateIdentity(8)
	for pos, P := range Ps {
		outAff := getOutputAffineEncoding(constr, fastConstr, 0, unshiftRows(pos))
		A, b := DecomposeAffineEncoding(encoding.ComposedBytes{outAff, P})

		if fmt.Sprintf("%x", id) != fmt.Sprintf("%x", A) {
			t.Fatalf("Linear part of encoding was not identity!")
		}

		if roundKeys[1][unshiftRows(pos)] != b {
			t.Fatalf("Constant part of encoding was not key byte!")
		}
	}
}

func TestBackOneRound(t *testing.T) {
	_, key := testConstruction()
	baseConstr := saes.Construction{key}
	roundKeys := baseConstr.StretchedKey()

	for round := 1; round < 11; round++ {
		a, b := roundKeys[round-1], BackOneRound(roundKeys[round], round)
		if bytes.Compare(a, b) != 0 {
			t.Fatalf("Failed to move back one round on round %v!\nReal: %x\nCand: %x\n", round, a, b)
		}
	}
}

func TestRecoverKey(t *testing.T) {
	_, key := testConstruction()
	fastConstr := fastTestConstruction()

	cand := RecoverKey(fastConstr)

	if bytes.Compare(key, cand) != 0 {
		t.Fatalf("Recovered key was wrong!\nReal: %x\nCand: %x\n", key, cand)
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

func BenchmarkDecoposeAffineEncoding(b *testing.B) {
	aff := encoding.ByteAffine{
		encoding.ByteLinear(matrix.GenerateRandom(rand.Reader, 8)),
		0x60,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		DecomposeAffineEncoding(aff)
	}
}

func BenchmarkFindCharacteristic(b *testing.B) {
	M := matrix.GenerateRandom(rand.Reader, 8)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		FindCharacteristic(M)
	}
}
