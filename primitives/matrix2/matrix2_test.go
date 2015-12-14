package matrix2

import (
	"testing"
)

func TestGaussJordan(t *testing.T) {
	m := Matrix{
		Row{0x01, 0x02},
		Row{0x02, 0x01},
	}

	aug, _, _, ok := m.gaussJordan()

	if !ok {
		t.Fatalf("gaussJordan returned ok=false")
	}
	in := Row{0x01, 0x01}
	out := aug.Mul(m.Mul(in))

	if !in.Add(out).IsZero() {
		t.Fatalf("gaussJordan returned an incorrect inverse matrix")
	}
}

func TestNullSpace(t *testing.T) {
	m := Matrix{
		Row{0x01, 0x02},
		Row{0x02, 0x04},
	}

	r := m.NullSpace()
	mr := m.Mul(r)

	if r.IsZero() {
		t.Fatalf("NullSpace returned trivial element of nullspace")
	}

	if !mr.IsZero() {
		t.Fatalf("NullSpace returned an incorrect element of nullspace")
	}
}
