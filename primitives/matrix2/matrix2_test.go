package matrix2

import (
	"testing"

	"crypto/rand"

	"github.com/OpenWhiteBox/AES/primitives/number"
)

func TestNullSpace(t *testing.T) {
	m := GenerateTrueRandom(rand.Reader, 32)
	m[2] = m[3].ScalarMul(number.ByteFieldElem(0x03)) // Force matrix to be singular.
	m[4] = m[3].ScalarMul(number.ByteFieldElem(0x07))

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
