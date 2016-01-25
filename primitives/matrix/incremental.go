package matrix

import (
	"crypto/rand"
	"sort"
)

// IncrementalMatrix is an invertible matrix that can be generated incrementally. It implements sort.Interface (but
// don't worry about that).
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

// reduce takes an arbitrary row as input and reduces it according to the Gauss-Jordan method with the current matrix.
// It returns the reduced row and the corresponding row in the inverse matrix.
func (im *IncrementalMatrix) reduce(raw Row) (Row, Row) {
	if raw.Size() != im.n {
		panic("Tried to reduce incorrectly sized with incremental matrix!")
	}

	reduced := raw.Dup()
	inverse := NewRow(im.n)
	inverse.SetBit(len(im.raw), true)

	// Put cand in simplest form.
	for i, _ := range im.simplest {
		if im.simplest[i].Cancels(reduced) {
			reduced = reduced.Add(im.simplest[i])
			inverse = inverse.Add(im.inverse[i])
		}
	}

	return reduced, inverse
}

// addRows adds each row to their respective matrices and puts im.simplest back in simplest form.
func (im *IncrementalMatrix) addRows(raw, reduced, inverse Row) {
	// Cancel every other row in the simplest form with cand.
	for i, _ := range im.simplest {
		if reduced.Cancels(im.simplest[i]) {
			im.simplest[i] = im.simplest[i].Add(reduced)
			im.inverse[i] = im.inverse[i].Add(inverse)
		}
	}

	im.raw = append(im.raw, raw.Dup())
	im.simplest = append(im.simplest, reduced.Dup())
	im.inverse = append(im.inverse, inverse.Dup())
}

// Add tries to add the row to the matrix. It fails if the new row is linearly dependent with another row. Add returns
// success or failure.
func (im *IncrementalMatrix) Add(raw Row) bool {
	reduced, inverse := im.reduce(raw)

	if reduced.IsZero() {
		return false
	}

	im.addRows(raw, reduced, inverse)
	return true
}

// FullyDefined returns true if the matrix has been fully defined and false if it hasn't.
func (im *IncrementalMatrix) FullyDefined() bool {
	return im.n == len(im.raw)
}

// IsInSpan returns whether or not the given row can be expressed as a linear combination of currently known rows.
func (im *IncrementalMatrix) IsInSpan(in Row) bool {
	reduced, _ := im.reduce(in)
	return reduced.IsZero()
}

// Novel returns a random row that is out of the span of the current matrix.
func (im *IncrementalMatrix) Novel() Row {
	if im.FullyDefined() {
		return nil
	}

	for true {
		cand := GenerateRandomNonZeroRow(rand.Reader, im.n)

		if !im.IsInSpan(cand) {
			return cand
		}
	}

	return nil
}

// pad pads an incremental matrix with empty rows until it is square.
func (im *IncrementalMatrix) pad(in Matrix) Matrix {
	out := in.Dup()

	for len(out) < im.n {
		out = append(out, NewRow(im.n))
	}

	return out
}

// Matrix returns the generated matrix.
func (im *IncrementalMatrix) Matrix() Matrix {
	return im.pad(im.raw)
}

// Inverse returns the generated matrix's inverse.
func (im *IncrementalMatrix) Inverse() Matrix {
	sort.Sort(im)
	return im.pad(im.inverse)
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
