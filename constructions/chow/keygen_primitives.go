package chow

import (
	"../../primitives/encoding"
	"../../primitives/matrix"
	"../../primitives/number"
	"../saes"
)

type TBox struct {
	Constr   saes.Construction
	KeyByte1 byte
	KeyByte2 byte
}

func (tbox TBox) Get(i byte) byte {
	return tbox.Constr.SubByte(i^tbox.KeyByte1) ^ tbox.KeyByte2
}

type TyiTable uint

func (tyi TyiTable) Get(i byte) (out uint32) {
	// Calculate dot product of i and [0x02 0x01 0x01 0x03]
	j := number.ByteFieldElem(i)

	a := number.ByteFieldElem(2).Mul(j)
	b := number.ByteFieldElem(1).Mul(j)
	c := number.ByteFieldElem(3).Mul(j)

	// Merge into one output and rotate according to column.
	out = (uint32(a) << 24) | (uint32(b) << 16) | (uint32(b) << 8) | uint32(c)
	out = (out >> (8 * uint(tyi))) | (out << (32 - 8*uint(tyi)))

	return
}

type MBInverseTable struct {
	MBInverse matrix.WordMatrix
	Row       uint
}

func (mbinv MBInverseTable) Get(i byte) uint32 {
	return mbinv.MBInverse.Mul(uint32(i) << (24 - 8*mbinv.Row))
}

type XORTable struct{}

func (xor XORTable) Get(i byte) (out byte) {
	return (i >> 4) ^ (i & 0xf)
}

// Encodes the output of a T-Box/Tyi Table / the input of a top-level XOR.
//
//    position: Position in the state array, counted in *bytes*.
// subPosition: Position in the T-Box/Tyi Table's ouptput for this byte, counted in nibbles.
func TyiEncoding(seed [16]byte, round, position, subPosition int) encoding.Nibble {
	label := [16]byte{}
	label[0], label[1], label[2], label[3] = 'T', byte(round), byte(position), byte(subPosition)

	return encoding.GenerateShuffle(generateStream(seed, label))
}

// Encodes the output of a MB^(-1) Table / the input of a top-level XOR.
//
//    position: Position in the state array, counted in *bytes*.
// subPosition: Position in the MB^(-1) Table's ouptput for this byte, counted in nibbles.
func MBInverseEncoding(seed [16]byte, round, position, subPosition int) encoding.Nibble {
	label := [16]byte{}
	label[0], label[1], label[2], label[3], label[4] = 'M', 'E', byte(round), byte(position), byte(subPosition)

	return encoding.GenerateShuffle(generateStream(seed, label))
}

// Encodes intermediate results between the two top-level XORs and the bottom XOR.
// The bottom XOR decodes its input with a Left and Right XOREncoding and encodes its output with a RoundEncoding.
//
// position: Position in the state array, counted in nibbles.
//  surface: Location relative to the round structure. Inside or Outside.
//     side: "Side" of the circuit. Left or Right.
func XOREncoding(seed [16]byte, round, position int, surface Surface, side Side) encoding.Nibble {
	label := [16]byte{}
	label[0], label[1], label[2], label[3], label[4] = 'X', byte(round), byte(position), byte(surface), byte(side)

	return encoding.GenerateShuffle(generateStream(seed, label))
}

// Encodes the output of an Expand->Squash round. Two Expand->Squash rounds make up one AES round.
//
// position: Position in the state array, counted in nibbles.
//  surface: Location relative to the AES round structure. Inside or Outside.
func RoundEncoding(seed [16]byte, round, position int, surface Surface) encoding.Nibble {
	label := [16]byte{}
	label[0], label[1], label[2], label[3] = 'R', byte(round), byte(position), byte(surface)

	return encoding.GenerateShuffle(generateStream(seed, label))
}
