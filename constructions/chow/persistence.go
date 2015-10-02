package chow

import (
	"github.com/OpenWhiteBox/AES/primitives/table"
)

func (constr *Construction) Serialize() []byte {
	out := []byte{}

	// Input Mask
	out = append(out, serializeMaskTables(constr.InputMask)...)
	out = append(out, serializeLargeXORTables(constr.InputXORTable)...)

	// First half of round
	out = append(out, serializeStepTables(constr.TBoxTyiTable)...)
	out = append(out, serializeXORTables(constr.HighXORTable)...)

	// Second half of round
	out = append(out, serializeStepTables(constr.MBInverseTable)...)
	out = append(out, serializeXORTables(constr.LowXORTable)...)

	// Output Mask
	out = append(out, serializeMaskTables(constr.TBoxOutputMask)...)
	out = append(out, serializeLargeXORTables(constr.OutputXORTable)...)

	return out
}

func Parse(in []byte) (constr Construction) {
	var rest []byte

	constr.InputMask, rest = parseMaskTables(in)
	constr.InputXORTable, rest = parseLargeXORTables(rest)

	constr.TBoxTyiTable, rest = parseStepTables(rest)
	constr.HighXORTable, rest = parseXORTables(rest)

	constr.MBInverseTable, rest = parseStepTables(rest)
	constr.LowXORTable, rest = parseXORTables(rest)

	constr.TBoxOutputMask, rest = parseMaskTables(rest)
	constr.OutputXORTable, rest = parseLargeXORTables(rest)

	return
}

func serializeMaskTables(t [16]table.Block) []byte {
	out := []byte{}
	for _, mask := range t {
		out = append(out, table.SerializeBlock(mask)...)
	}

	return out
}

func parseMaskTables(in []byte) (out [16]table.Block, rest []byte) {
	size := 256 * 16
	for i := 0; i < 16; i++ {
		out[i] = table.ParsedBlock(in[size*i : size*(i+1)])
	}

	return out, in[size*16:]
}

func serializeLargeXORTables(t [32][15]table.Nibble) []byte {
	out := []byte{}
	for _, rack := range t {
		for _, xorTable := range rack {
			out = append(out, table.SerializeNibble(xorTable)...)
		}
	}

	return out
}

func parseLargeXORTables(in []byte) (out [32][15]table.Nibble, rest []byte) {
	size := 256 / 2
	for i := 0; i < 32; i++ {
		for j := 0; j < 15; j++ {
			loc := 15*i + j
			out[i][j] = table.ParsedNibble(in[size*loc : size*(loc+1)])
		}
	}

	return out, in[size*32*15:]
}

func serializeStepTables(t [9][16]table.Word) []byte {
	out := []byte{}
	for _, round := range t {
		for _, pos := range round {
			out = append(out, table.SerializeWord(pos)...)
		}
	}

	return out
}

func parseStepTables(in []byte) (out [9][16]table.Word, rest []byte) {
	size := 256 * 4
	for i := 0; i < 9; i++ {
		for j := 0; j < 16; j++ {
			loc := 16*i + j
			out[i][j] = table.ParsedWord(in[size*loc : size*(loc+1)])
		}
	}

	return out, in[size*9*16:]
}

func serializeXORTables(t [9][32][3]table.Nibble) []byte {
	out := []byte{}
	for _, round := range t {
		for _, pos := range round {
			for _, gate := range pos {
				out = append(out, table.SerializeNibble(gate)...)
			}
		}
	}

	return out
}

func parseXORTables(in []byte) (out [9][32][3]table.Nibble, rest []byte) {
	size := 256 / 2
	for i := 0; i < 9; i++ {
		for j := 0; j < 32; j++ {
			for k := 0; k < 3; k++ {
				loc := 32*3*i + 3*j + k
				out[i][j][k] = table.ParsedNibble(in[size*loc : size*(loc+1)])
			}
		}
	}

	return out, in[size*9*32*3:]
}
