package chow

import (
	"errors"

	"github.com/OpenWhiteBox/AES/primitives/table"
)

const (
	fullSize = 770048

	maskTableSize = 256 * 16
	stepTableSize = 256 * 4
	xorTableSize  = 256 / 2
)

func (constr *Construction) Serialize() []byte {
	out, base := make([]byte, fullSize), 0

	// Input Mask
	base += serializeMaskTables(out[base:], constr.InputMask)
	base += serializeLargeXORTables(out[base:], constr.InputXORTable)

	// First half of round
	base += serializeStepTables(out[base:], constr.TBoxTyiTable)
	base += serializeXORTables(out[base:], constr.HighXORTable)

	// Second half of round
	base += serializeStepTables(out[base:], constr.MBInverseTable)
	base += serializeXORTables(out[base:], constr.LowXORTable)

	// Output Mask
	base += serializeMaskTables(out[base:], constr.TBoxOutputMask)
	serializeLargeXORTables(out[base:], constr.OutputXORTable)

	return out
}

func Parse(in []byte) (constr Construction, err error) {
	var rest []byte

	constr.InputMask, rest = parseMaskTables(in)
	constr.InputXORTable, rest = parseLargeXORTables(rest)

	constr.TBoxTyiTable, rest = parseStepTables(rest)
	constr.HighXORTable, rest = parseXORTables(rest)

	constr.MBInverseTable, rest = parseStepTables(rest)
	constr.LowXORTable, rest = parseXORTables(rest)

	constr.TBoxOutputMask, rest = parseMaskTables(rest)
	constr.OutputXORTable, rest = parseLargeXORTables(rest)

	if rest == nil {
		err = errors.New("Parsing the table failed!")
	}

	return
}

func serializeMaskTables(dst []byte, t [16]table.Block) int {
	base := 0
	for _, mask := range t {
		base += copy(dst[base:], table.SerializeBlock(mask))
	}

	return base
}

func parseMaskTables(in []byte) (out [16]table.Block, rest []byte) {
	if in == nil || len(in) < maskTableSize*16 {
		return
	}

	for i := 0; i < 16; i++ {
		out[i] = table.ParsedBlock(in[maskTableSize*i : maskTableSize*(i+1)])
	}

	return out, in[maskTableSize*16:]
}

func serializeLargeXORTables(dst []byte, t [32][15]table.Nibble) int {
	base := 0
	for _, rack := range t {
		for _, xorTable := range rack {
			base += copy(dst[base:], table.SerializeNibble(xorTable))
		}
	}

	return base
}

func parseLargeXORTables(in []byte) (out [32][15]table.Nibble, rest []byte) {
	if in == nil || len(in) < xorTableSize*9*16 {
		return
	}

	for i := 0; i < 32; i++ {
		for j := 0; j < 15; j++ {
			loc := 15*i + j
			out[i][j] = table.ParsedNibble(in[xorTableSize*loc : xorTableSize*(loc+1)])
		}
	}

	return out, in[xorTableSize*32*15:]
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
