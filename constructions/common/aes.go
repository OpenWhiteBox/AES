package common

import (
	"github.com/OpenWhiteBox/primitives/number"

	"github.com/OpenWhiteBox/AES/constructions/saes"
)

// A T-Box computes the SubBytes and AddRoundKey steps.
type TBox struct {
	Constr   saes.Construction
	KeyByte1 byte
	KeyByte2 byte
}

func (tbox TBox) Get(i byte) byte {
	return tbox.Constr.SubByte(i^tbox.KeyByte1) ^ tbox.KeyByte2
}

type InvTBox struct {
	Constr   saes.Construction
	KeyByte1 byte
	KeyByte2 byte
}

func (tbox InvTBox) Get(i byte) byte {
	return tbox.Constr.UnSubByte(i^tbox.KeyByte1) ^ tbox.KeyByte2
}

// A Tyi Table computes the MixColumns step.
type TyiTable uint

func (tyi TyiTable) Get(i byte) (out [4]byte) {
	// Calculate dot product of i and [0x02 0x01 0x01 0x03]
	j := number.ByteFieldElem(i)

	a := byte(number.ByteFieldElem(0x02).Mul(j))
	b := byte(number.ByteFieldElem(0x01).Mul(j))
	c := byte(number.ByteFieldElem(0x03).Mul(j))

	// Merge into one output and rotate according to column.
	res := [4]byte{a, b, b, c}
	copy(out[:], append(res[(4-tyi):], res[0:(4-tyi)]...))

	return
}

type InvTyiTable uint

func (tyi InvTyiTable) Get(i byte) (out [4]byte) {
	// Calculate dot product of i and []
	j := number.ByteFieldElem(i)

	a := byte(number.ByteFieldElem(0x0e).Mul(j))
	b := byte(number.ByteFieldElem(0x09).Mul(j))
	c := byte(number.ByteFieldElem(0x0d).Mul(j))
	d := byte(number.ByteFieldElem(0x0b).Mul(j))

	// Merge into one output and rotate according to column.
	res := [4]byte{a, b, c, d}
	copy(out[:], append(res[(4-tyi):], res[0:(4-tyi)]...))

	return
}

func NoShift(i int) int {
	return i
}

// ShiftRows is the block permutation in AES. Index in, index out.
//
// Example: ShiftRows(5) = 1 because ShiftRows(block) returns [16]byte{block[0], block[5], ...
func ShiftRows(i int) int {
	return []int{0, 13, 10, 7, 4, 1, 14, 11, 8, 5, 2, 15, 12, 9, 6, 3}[i]
}

// UnShiftRows is the inverse of ShiftRows.
func UnShiftRows(i int) int {
	return []int{0, 5, 10, 15, 4, 9, 14, 3, 8, 13, 2, 7, 12, 1, 6, 11}[i]
}
