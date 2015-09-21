package matrix

import (
	"../number"
	"testing"
)

func TestMatrix(t *testing.T) {
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
