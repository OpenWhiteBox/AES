package matrix

import (
	"crypto/rand"
	"testing"

	"github.com/OpenWhiteBox/AES/primitives/number"
)

func TestByteMatrix(t *testing.T) {
	m := Matrix{ // AES S-Box
		Row{0xF1}, // 0b11110001
		Row{0xE3}, // 0b11100011
		Row{0xC7}, // 0b11000111
		Row{0x8F}, // 0b10001111
		Row{0x1F}, // 0b00011111
		Row{0x3E}, // 0b00111110
		Row{0x7C}, // 0b01111100
		Row{0xF8}, // 0b11111000
	}
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
