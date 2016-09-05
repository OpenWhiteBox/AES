package chow

import (
	"github.com/OpenWhiteBox/primitives/encoding"
	"github.com/OpenWhiteBox/primitives/matrix"
	"github.com/OpenWhiteBox/primitives/number"
)

// findMatrix finds an invertible matrix in a basis.
func findMatrix(basis []matrix.Row) matrix.Matrix {
	im := matrix.NewIncrementalMatrix(64)
	for _, row := range basis {
		im.Add(row)
	}

	size := im.Size()
	for i := 0; i < size; i++ {
		row := im.Row(i)
		cand := matrix.Matrix{}

		for _, v := range row {
			cand = append(cand, matrix.Row{v})
		}

		if _, ok := cand.Invert(); ok {
			return cand
		}
	}

	panic("Couldn't find an invertible matrix in the given basis!")
}

// affineLayer implements methods for disambiguating an affine layer of the SPN.
type affineLayer encoding.BlockAffine

func (al affineLayer) Encode(in [16]byte) [16]byte {
	return encoding.BlockAffine(al).Encode(in)
}

func (al affineLayer) Decode(in [16]byte) [16]byte {
	return encoding.BlockAffine(al).Decode(in)
}

// clean gets the affine layer back to MixColumns and returns the input and output parasites.
func (al *affineLayer) clean() (input, output encoding.ConcatenatedBlock) {
	// Clean off the non-GF(2^8) noise.
	for pos := 0; pos < 16; pos++ {
		input[pos] = al.inputParasite(pos)
		output[pos] = al.outputParasite(pos)
	}

	al.adjust(input, output)

	// Clean off as much of the GF(2^8) noise as possible.
	in, out := al.stripScalars()
	al.adjust(in, out)

	for pos := 0; pos < 16; pos++ {
		input[pos] = encoding.ComposedBytes{input[pos], in[pos]}
		output[pos] = encoding.ComposedBytes{out[pos], output[pos]}
	}

	return
}

// adjust fixes the affine layer for two concatenated block encodings which will be moved into the S-box layer.
func (al *affineLayer) adjust(input, output encoding.ConcatenatedBlock) {
	temp, _ := encoding.DecomposeBlockAffine(encoding.ComposedBlocks{
		encoding.InverseBlock{input},
		encoding.BlockAffine(*al),
		encoding.InverseBlock{output},
	})

	*al = affineLayer(temp)
}

// inputParasite returns the non-GF(2^8) part of the parasite on the output at position col.
func (al *affineLayer) inputParasite(col int) encoding.Byte {
	block := col / 4
	row := 4 * (col / 4)

	row0, col0 := row%4, col%4
	row1, col1 := (row0+1)%4, (col0+1)%4

	rowT, colT := row1+(4*block), col1+(4*block)

	blockAA, blockAB := al.getBlock(row, col), al.getBlock(row, colT)
	blockBA, blockBB := al.getBlock(rowT, col), al.getBlock(rowT, colT)

	blockAA, _ = blockAA.Invert()
	blockBB, _ = blockBB.Invert()

	B := blockAA.Compose(blockAB).Compose(blockBB).Compose(blockBA)
	lambda := unMixColumns[row0][col0].Compose(mixColumns[row0][col1]).
		Compose(unMixColumns[row1][col1]).Compose(mixColumns[row1][col0])

	return encoding.NewByteLinear(findMatrix(
		B.LeftStretch().Add(lambda.RightStretch()).NullSpace(),
	))
}

// outputParasite returns the non-GF(2^8) part of the parasite on the output at position row.
func (al *affineLayer) outputParasite(row int) encoding.Byte {
	block := row / 4
	col := 4 * (row / 4)

	row0, col0 := row%4, col%4
	row1, col1 := (row0+1)%4, (col0+1)%4

	rowT, colT := row1+(4*block), col1+(4*block)

	blockAA, blockAB := al.getBlock(row, col), al.getBlock(row, colT)
	blockBA, blockBB := al.getBlock(rowT, col), al.getBlock(rowT, colT)

	blockAB, _ = blockAB.Invert()
	blockBA, _ = blockBA.Invert()

	B := blockAA.Compose(blockBA).Compose(blockBB).Compose(blockAB)
	lambda := mixColumns[row0][col0].Compose(unMixColumns[row1][col0]).
		Compose(mixColumns[row1][col1]).Compose(unMixColumns[row0][col1])

	return encoding.NewByteLinear(findMatrix(
		B.RightStretch().Add(lambda.LeftStretch()).NullSpace(),
	))
}

// stripScalars gets rid of unknown scalars in each block of the affine layer. It leaves it exactly equal to MixColumns,
// but there is an unknown scalar in each block that will move into the S-box layers.
func (al *affineLayer) stripScalars() (encoding.ConcatenatedBlock, encoding.ConcatenatedBlock) {
	input, output := [16]encoding.ByteLinear{}, [16]encoding.ByteLinear{}

	for pos := 0; pos < 16; pos += 4 {
		found := false

		for guess := 1; guess < 256 && !found; guess++ { // Take a guess for the input scalar on the first column.
			input[pos], _ = encoding.DecomposeByteLinear(encoding.NewByteMultiplication(number.ByteFieldElem(guess)))

			// Given input scalar on first column, calculate output scalars on all rows.
			for i := pos; i < pos+4; i++ {
				mc, _ := mixColumns[i%4][0].Invert()
				output[i] = encoding.NewByteLinear(
					al.getBlock(i, pos).Compose(input[pos].Backwards).Compose(mc),
				)
			}

			// Given output scalar on each row, calculate input scalars on all columns.
			for i := pos + 1; i < pos+4; i++ {
				mc, _ := mixColumns[0][i%4].Invert()
				input[i] = encoding.NewByteLinear(
					al.getBlock(pos, i).Compose(output[pos].Backwards).Compose(mc),
				)
			}

			// Verify that guess is consistent.
			found = true

			for i := pos; i < pos+4 && found; i++ {
				for j := pos; j < pos+4 && found; j++ {
					cand := al.getBlock(i, j).Compose(output[i].Backwards).Compose(input[j].Backwards)
					real := mixColumns[i%4][j%4]

					if !cand.Equals(real) {
						found = false
					}
				}
			}
		}

		if !found {
			panic("Failed to disambiguate block affine layer!")
		}
	}

	in, out := encoding.ConcatenatedBlock{}, encoding.ConcatenatedBlock{}
	for pos := 0; pos < 16; pos++ {
		in[pos], out[pos] = input[pos], output[pos]
	}

	return in, out
}

// getBlock returns the 8-by-8 block of the affine layer at the given position.
func (al *affineLayer) getBlock(row, col int) matrix.Matrix {
	out := matrix.Matrix{}

	for i := 0; i < 8; i++ {
		out = append(out, matrix.Row{al.BlockLinear.Forwards[8*row+i][col]})
	}

	return out
}
