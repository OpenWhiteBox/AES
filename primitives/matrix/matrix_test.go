package matrix

import (
  "testing"
)

func TestMatrix(t *testing.T) {
  m := ByteMatrix{ // AES S-Box
    143, // 0b10001111
    199, // 0b11000111
    227, // 0b11100011
    241, // 0b11110001
    248, // 0b11111000
    124, // 0b01111100
     62, // 0b00111110
     31, // 0b00011111
  }
  a := byte(99) // 0b01100011

  if (m.Mul(0x53) ^ a) != 0xed {
    t.Fatalf("Substitution value was wrong!")
  }
}
