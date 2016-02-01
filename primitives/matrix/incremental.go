package matrix

import (
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
	frees    []int  // Set of free variables.
}

// NewIncrementalMatrix initializes a new n-by-n incremental matrix.
func NewIncrementalMatrix(n int) IncrementalMatrix {
	frees := make([]int, n)
	for i := 0; i < n; i++ {
		frees[i] = i
	}

	return IncrementalMatrix{
		n:        n,
		raw:      Matrix{},
		simplest: Matrix{},
		inverse:  Matrix{},
		frees:    frees,
	}
}

// reduce takes an arbitrary row as input and reduces it according to the Gauss-Jordan method with the current matrix.
// It returns the reduced row and the corresponding row in the inverse matrix.
func (im *IncrementalMatrix) reduce(raw Row) (Row, Row) {
	if raw.Size() != im.n {
		panic("Tried to reduce incorrectly sized row with incremental matrix!")
	}

	reduced := raw.Dup()
	inverse := NewRow(im.n)
	if len(im.raw) < im.n {
		inverse.SetBit(len(im.raw), true)
	}

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

	idx := sort.SearchInts(im.frees, reduced.Height())
	im.frees = append(im.frees[0:idx], im.frees[idx+1:]...)
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

// IsIn returns whether or not the given row can be expressed as a linear combination of known rows.
func (im *IncrementalMatrix) IsIn(in Row) bool {
	reduced, _ := im.reduce(in)
	return reduced.IsZero()
}

// Size returns the number of rows that can be expressed as a linear combination of the known rows of the matrix.
func (im *IncrementalMatrix) Size() int {
	return 1 << uint(len(im.raw))
}

// Row returns the nth row that is a linear combination of the known rows of the matrix. n will be considered modulo
// im.Size().
func (im *IncrementalMatrix) Row(n int) Row {
	out := NewRow(im.n)

	for i := uint(0); i < uint(len(im.raw)); i++ {
		if (n>>i)&1 == 1 {
			out = out.Add(im.raw[i])
		}
	}

	return out
}

func (im *IncrementalMatrix) freeSize() int {
	return (1 << uint(len(im.frees))) - 1 // All combinations of free variables except the empty one.
}

// NovelSize returns the number of rows that are NOT a linear combination of the known rows of the matrix..
func (im *IncrementalMatrix) NovelSize() int {
	return im.Size() * im.freeSize()
}

// NovelRow returns the nth row that is NOT a linear combination of the known rows of the matrix. n will be considered
// modulo im.NovelSize().
func (im *IncrementalMatrix) NovelRow(n int) Row {
	if im.FullyDefined() {
		return nil
	}

	// Extract choices for free variables and rows.
	n = n % im.NovelSize()
	free := (n % im.freeSize()) + 1
	raw := n / im.freeSize()

	out := NewRow(im.n)
	for i := uint(0); i < uint(len(im.frees)); i++ { // Set all chosen free variables to true.
		if (free>>i)&1 == 1 {
			out.SetBit(im.frees[i], true)
		}
	}

	out = out.Add(im.Row(raw)) // Add the chosen rows from the raw matrix.

	return out
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
	frees := make([]int, len(im.frees))
	copy(frees, im.frees)

	return IncrementalMatrix{
		n:        im.n,
		raw:      im.raw.Dup(),
		simplest: im.simplest.Dup(),
		inverse:  im.inverse.Dup(),
		frees:    frees,
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
