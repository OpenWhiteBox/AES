package matrix

import (
	"../number"
	"crypto/rand"
	"testing"
)

func TestByteMatrix(t *testing.T) {
	m := ByteMatrix{ // AES S-Box
		0xF1, // 0b11110001
		0xE3, // 0b11100011
		0xC7, // 0b11000111
		0x8F, // 0b10001111
		0x1F, // 0b00011111
		0x3E, // 0b00111110
		0x7C, // 0b01111100
		0xF8, // 0b11111000
	}
	a := byte(0x63) // 0b01100011

	if m.Mul(byte(number.ByteFieldElem(0x53).Invert()))^a != 0xED {
		t.Fatalf("Substitution value was wrong! 0x53 -> 0xED")
	}

	if m.Mul(byte(number.ByteFieldElem(0x10).Invert()))^a != 0xCA {
		t.Fatalf("Substitution value was wrong! 0x10 -> 0xCA")
	}
}

func TestByteInvert(t *testing.T) {
	m := GenerateRandomByte(rand.Reader)
	n, _ := m.Invert()

	for i := 0; i < 256; i++ {
		nm := n.Mul(m.Mul(byte(i)))
		mn := m.Mul(n.Mul(byte(i)))

		if nm != byte(i) || mn != byte(i) {
			t.Fatalf("M * M^-1 != M^-1 * M != I")
		}
	}
}

func TestWordInvert(t *testing.T) {
	m := GenerateRandomWord(rand.Reader)
	n, _ := m.Invert()

	for i := 0; i < 256; i++ {
		nm := n.Mul(m.Mul(uint32(i)))
		mn := m.Mul(n.Mul(uint32(i)))

		if nm != uint32(i) || mn != uint32(i) {
			t.Fatalf("M * M^-1 != M^-1 * M != I")
		}
	}
}
