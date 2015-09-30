package saes

import (
	"github.com/OpenWhiteBox/AES/primitives/matrix"
	"github.com/OpenWhiteBox/AES/primitives/number"
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
	roundKeys := constr.StretchedKey()

	block = constr.AddRoundKey(roundKeys[0], block)

	for i := 1; i <= 9; i++ {
		block = constr.SubBytes(block)
		block = constr.ShiftRows(block)
		block = constr.MixColumns(block)
		block = constr.AddRoundKey(roundKeys[i], block)
	}

	block = constr.SubBytes(block)
	block = constr.ShiftRows(block)
	block = constr.AddRoundKey(roundKeys[10], block)

	return block
}

func rotw(w uint32) uint32 { return w<<8 | w>>24 }

func (constr *Construction) StretchedKey() [11][16]byte {
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
			temp = constr.SubWord(rotw(temp)) ^ (uint32(powx[i/4-1]) << 24)
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

func (constr *Construction) AddRoundKey(roundKey, block [16]byte) (out [16]byte) {
	for i := 0; i < 16; i++ {
		out[i] = roundKey[i] ^ block[i]
	}

	return
}

func (constr *Construction) SubBytes(block [16]byte) (out [16]byte) {
	for i, _ := range block {
		out[i] = constr.SubByte(block[i])
	}

	return out
}

func (constr *Construction) SubWord(w uint32) uint32 {
	return (uint32(constr.SubByte(byte(w>>24))) << 24) |
		(uint32(constr.SubByte(byte(w>>16))) << 16) |
		(uint32(constr.SubByte(byte(w>>8))) << 8) |
		uint32(constr.SubByte(byte(w)))
}

func (constr *Construction) SubByte(e byte) byte {
	// AES S-Box
	m := matrix.Matrix{ // Linear component.
		matrix.Row{0xF1}, // 0b11110001
		matrix.Row{0xE3}, // 0b11100011
		matrix.Row{0xC7}, // 0b11000111
		matrix.Row{0x8F}, // 0b10001111
		matrix.Row{0x1F}, // 0b00011111
		matrix.Row{0x3E}, // 0b00111110
		matrix.Row{0x7C}, // 0b01111100
		matrix.Row{0xF8}, // 0b11111000
	}
	a := byte(0x63) // 0b01100011 - Affine component.

	return m.Mul(matrix.Row{byte(number.ByteFieldElem(e).Invert())})[0] ^ a
}

func (constr *Construction) ShiftRows(block [16]byte) [16]byte {
	return [16]byte{
		block[0], block[5], block[10], block[15], block[4], block[9], block[14], block[3], block[8], block[13], block[2],
		block[7], block[12], block[1], block[6], block[11],
	}
}

func (constr *Construction) MixColumns(block [16]byte) (out [16]byte) {
	for i := 0; i < 4; i++ {
		copy(out[4*i:4*(i+1)], constr.MixColumn(block[4*i:4*(i+1)]))
	}

	return out
}

func (constr *Construction) MixColumn(slice []byte) (out []byte) {
	column := number.ArrayFieldElem{}
	for i := 0; i < 4; i++ {
		column = append(column, number.ByteFieldElem(slice[i]))
	}

	column = column.Mul(number.ArrayFieldElem{
		number.ByteFieldElem(0x02), number.ByteFieldElem(0x01),
		number.ByteFieldElem(0x01), number.ByteFieldElem(0x03),
	})

	for i := 0; i < len(column); i++ {
		out = append(out, byte(column[i]))
	}

	return
}
