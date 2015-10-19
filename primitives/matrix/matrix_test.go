package matrix

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/OpenWhiteBox/AES/primitives/number"
)

var SBox = Matrix{ // AES S-Box
	Row{0xF1}, // 0b11110001
	Row{0xE3}, // 0b11100011
	Row{0xC7}, // 0b11000111
	Row{0x8F}, // 0b10001111
	Row{0x1F}, // 0b00011111
	Row{0x3E}, // 0b00111110
	Row{0x7C}, // 0b01111100
	Row{0xF8}, // 0b11111000
}

func TestByteMatrix(t *testing.T) {
	m := SBox
	a := byte(0x63) // 0b01100011

	if m.Mul(Row{byte(number.ByteFieldElem(0x53).Invert())})[0]^a != 0xED {
		t.Fatalf("Substitution value was wrong! 0x53 -> 0xED")
	}

	if m.Mul(Row{byte(number.ByteFieldElem(0x10).Invert())})[0]^a != 0xCA {
		t.Fatalf("Substitution value was wrong! 0x10 -> 0xCA")
	}
}

func TestByteInvert(t *testing.T) {
	m := GenerateRandom(rand.Reader, 8)
	n, _ := m.Invert()

	for i := 0; i < 256; i++ {
		nm := n.Mul(m.Mul(Row{byte(i)}))[0]
		mn := m.Mul(n.Mul(Row{byte(i)}))[0]

		if nm != byte(i) || mn != byte(i) {
			t.Fatalf("M * M^-1 != M^-1 * M != I")
		}
	}
}

func TestWordInvert(t *testing.T) {
	m := GenerateRandom(rand.Reader, 32)
	n, _ := m.Invert()

	for i := 0; i < 1024; i++ {
		r := Row{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)}

		nm := n.Mul(m.Mul(r))
		mn := m.Mul(n.Mul(r))

		for j := 0; j < 4; j++ {
			if nm[j] != r[j] || mn[j] != r[j] {
				t.Fatalf("M * M^-1 != M^-1 * M != I")
			}
		}
	}
}

func TestBlockInvert(t *testing.T) {
	m := GenerateRandom(rand.Reader, 128)
	n, _ := m.Invert()

	for i := 0; i < 1024; i++ {
		r := Row{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)}

		nm := n.Mul(m.Mul(r))
		mn := m.Mul(n.Mul(r))

		for j := 0; j < 16; j++ {
			if nm[j] != r[j] || mn[j] != r[j] {
				t.Fatalf("M * M^-1 != M^-1 * M != I")
			}
		}
	}
}

func TestTranspose(t *testing.T) {
	m := GenerateIdentity(8)
	m[0].SetBit(7, true)

	mT := m.Transpose()

	if mT[7].GetBit(0) != 1 {
		t.Fatalf("Transpose is didn't flip the coordinates of a bit off the diagonal!")
	}

	m[0].SetBit(7, false)
	mT[7].SetBit(0, false)

	for i := 0; i < 8; i++ {
		if m[i][0] != mT[i][0] {
			t.Fatalf("Transpose didn't fix elements on the diagonal!")
		}
	}
}

func TestCompose(t *testing.T) {
	m := SBox
	n, _ := m.Invert()

	x, y, z := m.Compose(n), n.Compose(m), GenerateIdentity(8)

	for i := 0; i < 8; i++ {
		if x[i][0] != y[i][0] || y[i][0] != z[i][0] {
			t.Fatalf("Matrix composition is wrong!\nX = %x\nY = %x\nZ = %x\n", x, y, z)
		}
	}
}

func TestRightStretch(t *testing.T) {
	M := GenerateRandom(rand.Reader, 8)
	sboxRow := Row{0xF1, 0xE3, 0xC7, 0x8F, 0x1F, 0x3E, 0x7C, 0xF8}

	MS := M.RightStretch()

	real, cand := M.Compose(SBox), MS.Mul(sboxRow)

	for i := 0; i < 8; i++ {
		if real[i][0] != cand[i] {
			t.Fatalf("cand is not the inlining of real!\nreal = %v\ncand = %v", real, cand)
		}
	}
}

func TestLeftStretch(t *testing.T) {
	M := GenerateRandom(rand.Reader, 8)
	sboxRow := Row{0xF1, 0xE3, 0xC7, 0x8F, 0x1F, 0x3E, 0x7C, 0xF8}

	MS := M.LeftStretch()

	real, cand := SBox.Compose(M), MS.Mul(sboxRow)

	for i := 0; i < 8; i++ {
		if real[i][0] != cand[i] {
			t.Fatalf("cand is not the inlining of real!\nreal = %v\ncand = %v", real, cand)
		}
	}
}

func TestNullSpace(t *testing.T) {
	for i := 0; i < 100; i++ {
		m := GenerateTrueRandom(rand.Reader, 64)
		x := m.NullSpace()

		if fmt.Sprintf("%x", m.Mul(x)) != "0000000000000000" {
			t.Fatalf("Didn't find an actual element of the nullspace!\n x = %x\nMx = %x\n", x, m.Mul(x))
		}
	}
}

func TestTrace(t *testing.T) {
	m := Matrix{Row{12}, Row{20}, Row{41}, Row{94}, Row{176}, Row{97}, Row{195}, Row{134}}
	n := Matrix{Row{53}, Row{95}, Row{191}, Row{75}, Row{163}, Row{70}, Row{141}, Row{26}}

	if m.Trace() != 1 {
		t.Fatalf("Reported wrong trace for m!  Should be 1, got %v.", m.Trace())
	}

	if n.Trace() != 0 {
		t.Fatalf("Reported wrong trace for n! Should be 0, got %v", n.Trace())
	}
}
