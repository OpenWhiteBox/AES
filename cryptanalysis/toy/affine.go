package toy

import (
	"fmt"

	"github.com/OpenWhiteBox/primitives/encoding"
	"github.com/OpenWhiteBox/primitives/gfmatrix"
	"github.com/OpenWhiteBox/primitives/matrix"
)

// toString converts an 8-by-8 matrix into a hex string.
func toString(m matrix.Matrix) (out string) {
	for _, r := range m {
		out = out + fmt.Sprintf("%2.2x", r[0])
	}
	return out
}

// afineLayer implements methods for disambiguating an affine layer of the SPN.
type affineLayer encoding.BlockAffine

func (al affineLayer) Encode(in [16]byte) [16]byte {
	return encoding.BlockAffine(al).Encode(in)
}

func (al affineLayer) Decode(in [16]byte) [16]byte {
	return encoding.BlockAffine(al).Decode(in)
}

// adjust fixes the affine layer for two block encodings which will be moved somewhere else.
func (al *affineLayer) adjust(input, output encoding.Block) {
	temp, _ := encoding.DecomposeBlockAffine(encoding.ComposedBlocks{
		encoding.InverseBlock{input},
		encoding.BlockAffine(*al),
		encoding.InverseBlock{output},
	})

	*al = affineLayer(temp)
}

// cleanParasites removes the given parasites from the input and output of al. in are the input parasites for this
// affine layer, and nextIn are the input parasites of the following affine layer.
func (al *affineLayer) cleanParasites(in, nextIn [16]*parasite) {
	input, output := encoding.ConcatenatedBlock{}, encoding.ConcatenatedBlock{}

	for pos := 0; pos < 16; pos++ {
		input[pos], output[pos] = in[pos], nextIn[pos].Convert()
	}

	al.adjust(input, output)
}

// parasites returns the input parasite for each byte of the input of al.
func (al *affineLayer) parasites() (input [16]*parasite) {
	for col := 0; col < 16; col++ {
		for row := 0; row < 16; row++ {
			input[col] = al.blockParasite(row, col)
			if input[col] != nil {
				break
			}
		}

		if input[col] == nil {
			panic("one column of matrix has all zero blocks")
		}
	}

	return
}

// blockParasites returns the input parasite of the block at the given row and column of the linear part of al, or nil
// if the block is zero.
func (al *affineLayer) blockParasite(row, col int) (input *parasite) {
	block := al.getBlock(row, col)
	if toString(block) == "0000000000000000" {
		return nil
	}
	enc := encoding.NewByteLinear(block)

	for _, input = range parasites {
		cand, _ := encoding.DecomposeByteLinear(encoding.ComposedBytes{
			encoding.InverseByte{subBytes}, encoding.InverseByte{input}, enc,
		})
		_, ok := parasites[toString(cand.Forwards)]
		if ok {
			return
		}
	}

	panic("unable to determine input parasite")
}

// compress maps a permuted round matrix to a matrix over F_{2^8} by converting each block to its MixColumns coefficient.
func (al *affineLayer) compress() smallMatrix {
	sm := gfmatrix.GenerateEmpty(16, 16)

	var ok bool
	for row := 0; row < 16; row++ {
		for col := 0; col < 16; col++ {
			sm[row][col], ok = blocks[toString(al.getBlock(row, col))]
			if !ok {
				sm[row][col] = 255
				// panic("unknown block in matrix")
			}
		}
	}

	return smallMatrix(sm)
}

// getBlock returns the 8-by-8 block of the affine layer at the given position.
func (al *affineLayer) getBlock(row, col int) matrix.Matrix {
	out := matrix.Matrix{}

	for i := 0; i < 8; i++ {
		out = append(out, matrix.Row{al.BlockLinear.Forwards[8*row+i][col]})
	}

	return out
}
