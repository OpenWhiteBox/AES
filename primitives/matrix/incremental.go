package matrix

import (
	"sort"
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
	inverseRow := NewRow(im.n)
	inverseRow.SetBit(len(im.raw), true)

	// Put cand in simplest form.
	for i, _ := range im.simplest {
		if im.simplest[i].Cancels(cand) {
			cand = cand.Add(im.simplest[i])
			inverseRow = inverseRow.Add(im.inverse[i])
		}
	}

	if cand.IsZero() {
		return false
	}

	// Cancel every other row in the simplest form with cand.
	for i, _ := range im.simplest {
		if cand.Cancels(im.simplest[i]) {
			im.simplest[i] = im.simplest[i].Add(cand)
			im.inverse[i] = im.inverse[i].Add(inverseRow)
		}
	}

	im.raw = append(im.raw, candM.Dup())
	im.simplest = append(im.simplest, cand)
	im.inverse = append(im.inverse, inverseRow)

	return true
}

// FullyDefined returns true if the matrix has been fully defined and false if it hasn't.
func (im *IncrementalMatrix) FullyDefined() bool {
	n, m := im.raw.Size()
	return n == m
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

// Dup returns a duplicate of im.
func (im *IncrementalMatrix) Dup() IncrementalMatrix {
	return IncrementalMatrix{
		n:        im.n,
		raw:      im.raw.Dup(),
		simplest: im.simplest.Dup(),
		inverse:  im.inverse.Dup(),
	}
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
