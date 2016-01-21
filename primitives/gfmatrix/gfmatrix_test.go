package gfmatrix

import (
	"testing"

	"crypto/rand"
)

func TestNullSpace(t *testing.T) {
	m := GenerateTrueRandom(rand.Reader, 32)
	m[2] = m[3].ScalarMul(0x03) // Force matrix to be singular.
	m[4] = m[3].ScalarMul(0x07)

	basis := m.NullSpace()

	if len(basis) < 2 {
		t.Fatalf("NullSpace returned a basis that's too small")
	}

	a := m.Mul(basis[0])
	b := m.Mul(basis[1])
	c := m.Mul(
		basis[0].Add(basis[1]),
	)

	if !a.IsZero() || !b.IsZero() || !c.IsZero() {
		t.Fatalf("NullSpace returned a malformed basis.")
	}
}

func TestInvert(t *testing.T) {
	m := Matrix{
		Row{0x01, 0x00, 0x00, 0x00},
		Row{0x07, 0x01, 0x00, 0x00},
		Row{0x00, 0x03, 0x01, 0x00},
		Row{0x00, 0x00, 0x06, 0x01},
	}

	mInv, ok := m.Invert()
	if !ok {
		t.Fatalf("Failed to invert invertable matrix.")
	}

	in := GenerateRandomRow(rand.Reader, 4)
	out := mInv.Mul(m.Mul(in))

	if !in.Equals(out) {
		t.Fatalf("Inverse was wrong!")
	}
}
