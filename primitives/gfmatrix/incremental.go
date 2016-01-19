package gfmatrix

import (
	"sort"

	"github.com/OpenWhiteBox/AES/primitives/number"
)

// IncrementalMatrix is an invertible matrix that can be generated incrementally. Implements sort.Interface.
//
// For example, in cryptanalyses, we might be able to do some work and discover some rows of a matrix. We want to stop
// working as soon as its fully defined, but we also can't just work until we have n rows because we might have
// recovered duplicate or linearly dependent rows.
type IncrementalMatrix struct {
	n        int    // The dimension of the matrix.
	raw      Matrix // The collection of rows as they were put in.
	simplest Matrix // The matrix in Gauss-Jordan eliminated form.
	inverse  Matrix // The inverse matrix of raw.
}

// NewIncrementalMatrix initializes a new n-by-n incremental matrix.
func NewIncrementalMatrix(n int) IncrementalMatrix {
	return IncrementalMatrix{
		n:        n,
		raw:      Matrix{},
		simplest: Matrix{},
		inverse:  Matrix{},
	}
}

// Add tries to add the row to the matrix. It fails if the new row is linearly dependent with another row. Add returns
// success or failure.
func (im *IncrementalMatrix) Add(candM Row) bool {
	if candM.Size() != im.n {
		panic("Tried to add incorrectly sized row to incremental matrix!")
	}

	cand := candM.Dup()
	inverseRow := Row(make([]number.ByteFieldElem, im.n))
	inverseRow[len(im.raw)] = 0x01

	// Put cand in simplest form.
	for i, _ := range im.simplest {
		height := im.simplest[i].Height()
		if !cand[height].IsZero() {
			correction := cand[height]

			cand = cand.Add(im.simplest[i].ScalarMul(correction))
			inverseRow = inverseRow.Add(im.inverse[i].ScalarMul(correction))
		}
	}

	if cand.IsZero() {
		return false
	}

	height := cand.Height()

	correction := cand[height].Invert()
	cand, inverseRow = cand.ScalarMul(correction), inverseRow.ScalarMul(correction)

	// Cancel every other row in the simplest form with cand.
	for i, _ := range im.simplest {
		if !im.simplest[i][height].IsZero() {
			correction := im.simplest[i][height]

			im.simplest[i] = im.simplest[i].Add(cand.ScalarMul(correction))
			im.inverse[i] = im.inverse[i].Add(inverseRow.ScalarMul(correction))
		}
	}

	im.raw = append(im.raw, candM.Dup())
	im.simplest = append(im.simplest, cand)
	im.inverse = append(im.inverse, inverseRow)

	return true
}

// FullyDefined returns true if the matrix has been fully defined and false if it hasn't.
func (im *IncrementalMatrix) FullyDefined() bool {
	return im.n == len(im.raw)
}

// Matrix returns the generated matrix.
func (im *IncrementalMatrix) Matrix() Matrix {
	return im.raw
}

// Inverse returns the generated matrix's inverse.
func (im *IncrementalMatrix) Inverse() Matrix {
	sort.Sort(im)
	return im.inverse
}

// Implementation of sort.Interface

func (im *IncrementalMatrix) Len() int {
	return len(im.raw)
}

func (im *IncrementalMatrix) Less(i, j int) bool {
	return LessThan(im.simplest[i], im.simplest[j])
}

func (im *IncrementalMatrix) Swap(i, j int) {
	im.simplest[i], im.simplest[j] = im.simplest[j], im.simplest[i]
	im.inverse[i], im.inverse[j] = im.inverse[j], im.inverse[i]
}
