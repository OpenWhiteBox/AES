package equivalence

import (
	"testing"

	"github.com/OpenWhiteBox/AES/primitives/encoding"
	"github.com/OpenWhiteBox/AES/primitives/matrix"
	"github.com/OpenWhiteBox/AES/primitives/number"
)

type InvertEncoding struct{}

func (ie InvertEncoding) Encode(in byte) byte { return byte(number.ByteFieldElem(in).Invert()) }
func (ie InvertEncoding) Decode(in byte) byte { return ie.Encode(in) }

func TestLinearEquivalence(t *testing.T) {
	cap := 2041
	if testing.Short() {
		cap = 20
	}

	f := InvertEncoding{}
	eqs := LinearEquivalence(f, f, cap)

	if !(testing.Short() && len(eqs) == 20 || !testing.Short() && len(eqs) == 2040) {
		t.Fatalf("LinearEquivalence found the wrong number of equivalences! Wanted %v, got %v.", 2040, len(eqs))
	}

	for _, eq := range eqs {
		fA := encoding.ComposedBytes{encoding.ByteLinear{eq.A, nil}, f}
		Bf := encoding.ComposedBytes{f, encoding.ByteLinear{eq.B, nil}}

		if !encoding.EquivalentBytes(fA, Bf) {
			t.Fatalf("LinearEquivalence found an incorrect equivalence.")
		}
	}
}

func TestLearnConsistent(t *testing.T) {
	f := InvertEncoding{}
	A, B := matrix.NewDeductiveMatrix(8), matrix.NewDeductiveMatrix(8)

	for i := uint(0); i < 7; i++ {
		A.Assert(matrix.Row{byte(1 << i)}, matrix.Row{byte(1 << i)})
	}

	consistent := learn(f, f, A, B)
	if !consistent {
		t.Fatal("Learn said identity matrix was inconsistent.")
	}

	if !A.FullyDefined() {
		t.Fatal("Learn did not propagate knowledge into A.")
	} else if !B.FullyDefined() {
		t.Fatal("Learn did not propagate knowledge into B.")
	}

	if A.Matrix()[7][0] != 1<<7 {
		t.Fatal("Learn determined A incorrectly.")
	} else if B.Matrix()[7][0] != 1<<7 {
		t.Fatal("Learn determined B incorrectly.")
	}
}

func TestLearnInconsistent(t *testing.T) {
	f := InvertEncoding{}
	A, B := matrix.NewDeductiveMatrix(8), matrix.NewDeductiveMatrix(8)

	for i := uint(0); i < 6; i++ {
		A.Assert(matrix.Row{byte(1 << i)}, matrix.Row{byte(1 << i)})
	}
	A.Assert(matrix.Row{byte(1 << 7)}, matrix.Row{byte(1 << 6)})

	consistent := learn(f, f, A, B)

	if consistent {
		t.Fatal("Learn said inconsistent matrix was consistent.")
	}
}
