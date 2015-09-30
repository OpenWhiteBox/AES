// Contains tables and encodings necessary for key generation.
package chow

import (
	"../../primitives/encoding"
	"../../primitives/matrix"
	"../../primitives/number"
	"../saes"
)

var encodingCache = make(map[[32]byte]encoding.Shuffle)

type MaskTable struct {
	Mask     matrix.Matrix
	Position int
}

func (mt MaskTable) Get(i byte) (out [16]byte) {
	r := make([]byte, 16)
	r[mt.Position] = i

	res := mt.Mask.Mul(matrix.Row(r))
	copy(out[:], res)

	return
}

// A T-Box computes the SubBytes and AddRoundKey steps.
type TBox struct {
	Constr   saes.Construction
	KeyByte1 byte
	KeyByte2 byte
}

func (tbox TBox) Get(i byte) byte {
	return tbox.Constr.SubByte(i^tbox.KeyByte1) ^ tbox.KeyByte2
}

// A Tyi Table computes the MixColumns step.
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

// A MB^(-1) Table inverts the mixing bijection on the Tyi Table.
type MBInverseTable struct {
	MBInverse matrix.Matrix
	Row       uint
}

func (mbinv MBInverseTable) Get(i byte) uint32 {
	r := matrix.Row{0, 0, 0, 0}
	r[mbinv.Row] = i

	out := mbinv.MBInverse.Mul(r)

	return uint32(out[0])<<24 | uint32(out[1])<<16 | uint32(out[2])<<8 | uint32(out[3])
}

// An XOR Table computes the XOR of two nibbles.
type XORTable struct{}

func (xor XORTable) Get(i byte) (out byte) {
	return (i >> 4) ^ (i & 0xf)
}

// Abstraction over the Tyi and MB^(-1) encodings, to match the pattern of the XOR and round encodings.
func StepEncoding(seed [16]byte, round, position, subPosition int, surface Surface) encoding.Nibble {
	if surface == Inside {
		return TyiEncoding(seed, round, position, subPosition)
	} else {
		return MBInverseEncoding(seed, round, position, subPosition)
	}
}

// Encodes the output of a T-Box/Tyi Table / the input of a top-level XOR.
//
//    position: Position in the state array, counted in *bytes*.
// subPosition: Position in the T-Box/Tyi Table's ouptput for this byte, counted in nibbles.
func TyiEncoding(seed [16]byte, round, position, subPosition int) encoding.Nibble {
	label := [16]byte{}
	label[0], label[1], label[2], label[3] = 'T', byte(round), byte(position), byte(subPosition)

	return getShuffle(seed, label)
}

// Encodes the output of a MB^(-1) Table / the input of a top-level XOR.
//
//    position: Position in the state array, counted in *bytes*.
// subPosition: Position in the MB^(-1) Table's ouptput for this byte, counted in nibbles.
func MBInverseEncoding(seed [16]byte, round, position, subPosition int) encoding.Nibble {
	label := [16]byte{}
	label[0], label[1], label[2], label[3], label[4] = 'M', 'E', byte(round), byte(position), byte(subPosition)

	return getShuffle(seed, label)
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

	return getShuffle(seed, label)
}

// Encodes the output of an Expand->Squash round. Two Expand->Squash rounds make up one AES round.
//
// position: Position in the state array, counted in nibbles.
//  surface: Location relative to the AES round structure. Inside or Outside.
func RoundEncoding(seed [16]byte, round, position int, surface Surface) encoding.Nibble {
	label := [16]byte{}
	label[0], label[1], label[2], label[3] = 'R', byte(round), byte(position), byte(surface)

	return getShuffle(seed, label)
}

func getShuffle(seed, label [16]byte) encoding.Shuffle {
	key := [32]byte{}
	copy(key[0:16], seed[:])
	copy(key[16:32], label[:])

	cached, ok := encodingCache[key]

	if ok {
		return cached
	} else {
		encodingCache[key] = encoding.GenerateShuffle(generateStream(seed, label))
		return encodingCache[key]
	}
}
