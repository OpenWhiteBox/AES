package xiao

import (
	"errors"

	"github.com/OpenWhiteBox/primitives/matrix"
	"github.com/OpenWhiteBox/primitives/table"
)

const (
	fullSize = 20994048

	matrixSize = 16 * 128
	tmcSize    = 65536 * 4
)

// Serialize serializes a white-box construction into a byte slice.
func (constr *Construction) Serialize() []byte {
	out, base := make([]byte, fullSize), 0

	base += serializeMatrix(out[base:], constr.FinalMask)

	for _, sr := range constr.ShiftRows {
		base += serializeMatrix(out[base:], sr)
	}

	for _, round := range constr.TBoxMixCol {
		for _, tmc := range round {
			base += copy(out[base:], table.SerializeDoubleToWord(tmc))
		}
	}

	return out
}

// Parse parses a byte array into a white-box construction. It returns an error if the byte array isn't long enough.
func Parse(in []byte) (constr Construction, err error) {
	var rest []byte

	constr.FinalMask, rest = parseMatrix(in)

	for i, _ := range constr.ShiftRows {
		constr.ShiftRows[i], rest = parseMatrix(rest)
	}

	for i, _ := range constr.TBoxMixCol {
		for j, _ := range constr.TBoxMixCol[i] {
			constr.TBoxMixCol[i][j] = table.ParsedDoubleToWord(rest[:tmcSize])
			rest = rest[tmcSize:]
		}
	}

	if rest == nil {
		err = errors.New("Parsing the key failed!")
	}

	return
}

func serializeMatrix(dst []byte, m matrix.Matrix) int {
	base := 0
	for _, row := range m {
		base += copy(dst[base:], row)
	}

	return base
}

func parseMatrix(in []byte) (out matrix.Matrix, rest []byte) {
	if in == nil || len(in) < matrixSize {
		return
	}

	out = matrix.Matrix(make([]matrix.Row, 128))
	for row, _ := range out {
		out[row] = in[16*row : 16*(row+1)]
	}

	return out, in[matrixSize:]
}
