package saes

import (
	"../../primitives/matrix"
	"../../primitives/number"
)

// Powers of x mod M(x).
var powx = [16]byte{
	0x01,
	0x02,
	0x04,
	0x08,
	0x10,
	0x20,
	0x40,
	0x80,
	0x1b,
	0x36,
	0x6c,
	0xd8,
	0xab,
	0x4d,
	0x9a,
	0x2f,
}

type Construction struct {
	Key [16]byte
}

func (constr *Construction) Encrypt(block [16]byte) [16]byte {
	roundKeys := constr.stretchedKey()

	block = constr.addRoundKey(roundKeys[0], block)

	for i := 1; i <= 9; i++ {
		block = constr.subBytes(block)
		block = constr.shiftRows(block)
		// block = constr.mixColumns(block)
		block = constr.addRoundKey(roundKeys[i], block)
	}

	block = constr.subBytes(block)
	block = constr.shiftRows(block)
	block = constr.addRoundKey(roundKeys[10], block)

	return block
}

func rotw(w uint32) uint32 { return w<<8 | w>>24 }

func (constr *Construction) stretchedKey() [11][16]byte {
	var (
		i         int            = 0
		temp      uint32         = 0
		stretched [4 * 11]uint32 // Stretched key
		split     [11][16]byte   // Each round key is combined and its uint32s are turned into 4 bytes
	)

	for ; i < 4; i++ { // First key-length of stretched is the raw key.
		stretched[i] = (uint32(constr.Key[4*i]) << 24) |
			(uint32(constr.Key[4*i+1]) << 16) |
			(uint32(constr.Key[4*i+2]) << 8) |
			uint32(constr.Key[4*i+3])
	}

	for ; i < (4 * 11); i++ {
		temp = stretched[i-1]

		if (i % 4) == 0 {
			temp = constr.subWord(rotw(temp)) ^ (uint32(powx[i/4-1]) << 24)
		} else {
			temp = constr.subWord(temp)
		}

		stretched[i] = stretched[i-4] ^ temp
	}

	for j := 0; j < 11; j++ {
		for k := 0; k < 4; k++ {
			word := stretched[4*j+k]

			split[j][4*k] = byte(word >> 24)
			split[j][4*k+1] = byte(word >> 16)
			split[j][4*k+2] = byte(word >> 8)
			split[j][4*k+3] = byte(word)
		}
	}

	return split
}

func (constr *Construction) addRoundKey(roundKey, block [16]byte) (out [16]byte) {
	for i := 0; i < 16; i++ {
		out[i] = roundKey[i] ^ block[i]
	}

	return
}

func (constr *Construction) subBytes(block [16]byte) (out [16]byte) {
	for i, _ := range block {
		out[i] = constr.subByte(block[i])
	}

	return block
}

func (constr *Construction) subWord(w uint32) uint32 {
	return (uint32(constr.subByte(byte(w>>24))) << 24) |
		(uint32(constr.subByte(byte(w>>16))) << 16) |
		(uint32(constr.subByte(byte(w>>8))) << 8) |
		uint32(constr.subByte(byte(w)))
}

func (constr *Construction) subByte(e byte) byte {
	// AES S-Box
	m := matrix.ByteMatrix{ // Linear component.
		0xF1, // 0b11110001
		0xE3, // 0b11100011
		0xC7, // 0b11000111
		0x8F, // 0b10001111
		0x1F, // 0b00011111
		0x3E, // 0b00111110
		0x7C, // 0b01111100
		0xF8, // 0b11111000
	}
	a := byte(0x63) // 0b01100011 - Affine component.

	return m.Mul(byte(number.ByteFieldElem(e).Invert())) ^ a
}

func (constr *Construction) shiftRows(block [16]byte) [16]byte {
	return [16]byte{
		block[0], block[5], block[10], block[15], block[4], block[9], block[14], block[3], block[8], block[13], block[2],
		block[7], block[12], block[1], block[6], block[11],
	}
}
