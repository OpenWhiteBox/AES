package chow

import (
	"errors"

	"github.com/OpenWhiteBox/primitives/table"

	"github.com/OpenWhiteBox/AES/constructions/common"
)

const (
	fullSize = 770048

	maskTableSize = 256 * 16
	stepTableSize = 256 * 4
	xorTableSize  = 256 / 2
)

// Serialize serializes a white-box construction into a byte slice.
func (constr *Construction) Serialize() []byte {
	out, base := make([]byte, fullSize), 0

	// Input Mask
	base += common.SerializeBlockMatrix(out[base:], constr.InputMask, constr.InputXORTables)

	// First half of round
	base += serializeStepTables(out[base:], constr.TBoxTyiTable)
	base += serializeXORTables(out[base:], constr.HighXORTable)

	// Second half of round
	base += serializeStepTables(out[base:], constr.MBInverseTable)
	base += serializeXORTables(out[base:], constr.LowXORTable)

	// Output Mask
	common.SerializeBlockMatrix(out[base:], constr.TBoxOutputMask, constr.OutputXORTables)

	return out
}

// Parse parses a byte array into a white-box construction. It returns an error if the byte array isn't long enough.
func Parse(in []byte) (constr Construction, err error) {
	var rest []byte

	constr.InputMask, constr.InputXORTables, rest = common.ParseBlockNibbleMatrix(in)

	constr.TBoxTyiTable, rest = parseStepTables(rest)
	constr.HighXORTable, rest = parseXORTables(rest)

	constr.MBInverseTable, rest = parseStepTables(rest)
	constr.LowXORTable, rest = parseXORTables(rest)

	constr.TBoxOutputMask, constr.OutputXORTables, rest = common.ParseBlockNibbleMatrix(rest)

	if rest == nil {
		err = errors.New("Parsing the key failed!")
	}

	return
}

func serializeStepTables(dst []byte, t [9][16]table.Word) int {
	base := 0
	for _, round := range t {
		for _, pos := range round {
			base += copy(dst[base:], table.SerializeWord(pos))
		}
	}

	return base
}

func parseStepTables(in []byte) (out [9][16]table.Word, rest []byte) {
	if in == nil || len(in) < stepTableSize*9*16 {
		return
	}

	for i := 0; i < 9; i++ {
		for j := 0; j < 16; j++ {
			loc := 16*i + j
			out[i][j] = table.ParsedWord(in[stepTableSize*loc : stepTableSize*(loc+1)])
		}
	}

	return out, in[stepTableSize*9*16:]
}

func serializeXORTables(dst []byte, t [9][32][3]table.Nibble) int {
	base := 0
	for _, round := range t {
		for _, pos := range round {
			for _, gate := range pos {
				base += copy(dst[base:], table.SerializeNibble(gate))
			}
		}
	}

	return base
}

func parseXORTables(in []byte) (out [9][32][3]table.Nibble, rest []byte) {
	if in == nil || len(in) < xorTableSize*9*32*3 {
		return
	}

	for i := 0; i < 9; i++ {
		for j := 0; j < 32; j++ {
			for k := 0; k < 3; k++ {
				loc := 32*3*i + 3*j + k
				out[i][j][k] = table.ParsedNibble(in[xorTableSize*loc : xorTableSize*(loc+1)])
			}
		}
	}

	return out, in[xorTableSize*9*32*3:]
}
