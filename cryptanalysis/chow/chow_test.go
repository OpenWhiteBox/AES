package chow

import (
	"fmt"
	"testing"

	"github.com/OpenWhiteBox/AES/constructions/chow"
	"github.com/OpenWhiteBox/AES/primitives/table"

	test_vectors "github.com/OpenWhiteBox/AES/constructions/test"
)

func testConstruction() (chow.Construction, [16]byte) {
	key := test_vectors.AESVectors[50].Key
	seed := test_vectors.AESVectors[51].Key
	constr, _, _ := chow.GenerateKeys(key, seed)

	return constr, key
}

func TestRecoverKey(t *testing.T) {
	constr, _ := testConstruction()
	fmt.Println(RecoverKey(constr))
}

func TestFindBasisAndSort(t *testing.T) {
	constr, _ := testConstruction()
	S := GenerateS(constr)

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

	S := GenerateS(constr)
	_ = FindBasisAndSort(S)

  qtilde := Qtilde{S}

	for i := 0; i < 256; i++ {
		cand := qtilde.Encode(qtilde.Decode(byte(i)) ^ 37)

		if S[37][i] != cand {
			t.Fatalf("Identity broken! f = Qtilde <- xor psi <- Qtilde^(-1)\nf(%v) = %v, g(%v) = %v", i, S[37][i], i, cand)
		}
  }
}
