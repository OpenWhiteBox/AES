package equivalence

import (
	"sort"

	"github.com/OpenWhiteBox/AES/primitives/matrix"
)

type Elem struct {
	In, Out matrix.Row
}

type IncrementalMatrix struct {
	Raw, Simplest matrix.Matrix
}

func NewIncrementalMatrix() IncrementalMatrix {
	return IncrementalMatrix{
		Raw:      matrix.Matrix{},
		Simplest: matrix.Matrix{},
	}
}

// AddRow adds the row to the matrix if and only if it's linearly independent from the rest of the rows. It returns
// success or failure.
func (im *IncrementalMatrix) Add(candM matrix.Row) bool {
	cand := candM.Dup()

	for _, row := range im.Simplest {
		// I want to know if lambda(row & cand) = lambda(row)
		// (row & cand) ^ row <= (row & cand) & row
		// row & !cand <= cand & row

		if row[0]&^cand[0] <= row[0]&cand[0] {
			cand = cand.Add(row)
		}
	}

	if cand.IsZero() {
		return false
	} else {
		im.Raw = append(im.Raw, candM.Dup())
		im.Simplest = append(im.Simplest, cand)

		sort.Sort(im)

		return true
	}
}

func (im *IncrementalMatrix) Matrix() matrix.Matrix {
	return im.Raw.Transpose()
}

// Implementation of sort.Interface
func (im *IncrementalMatrix) Len() int { return len(im.Raw) }
func (im *IncrementalMatrix) Less(i, j int) bool {
	return matrix.LessThan(im.Simplest[i], im.Simplest[j])
}
func (im *IncrementalMatrix) Swap(i, j int) {
	im.Simplest[i], im.Simplest[j] = im.Simplest[j], im.Simplest[i]
}

type Matrix struct {
	Space     map[byte]byte
	NotInSpan map[byte]bool
}

// NewMatrix returns a new, empty matrix.
func NewMatrix() *Matrix {
	m := Matrix{
		Space:     make(map[byte]byte),
		NotInSpan: make(map[byte]bool),
	}
	m.Space[0x00] = 0x00

	for x := 1; x < 256; x++ {
		m.NotInSpan[byte(x)] = true
	}

	return &m
}

// Assert represents an assertion that A(in) = out. The function will panic if this is inconsistent with previous
// assertions. It it's not, it returns whether or not the assertion contained new information about A.
func (e *Matrix) Assert(in, out matrix.Row) (learned bool) {
	learned = false

	f := e.Dup()
	for x, y := range f.Space {
		k := x ^ in[0]

		yGot, ok := e.Space[k]
		yExpected := y ^ out[0]

		if ok && yGot != yExpected {
			panic("Inconsistency!")
		} else if !ok {
			e.Space[k] = yExpected

			delete(e.NotInSpan, yExpected)
			learned = true
		}
	}

	return
}

// NovelInput returns an x such that A(x) is unknown.
func (e *Matrix) NovelInput() matrix.Row {
	for x := 1; x < 256; x++ {
		_, ok := e.Space[byte(x)]
		if !ok {
			return matrix.Row{byte(x)}
		}
	}

	return nil
}

// IsInSpan returns whether or not x is in the known span of A.
func (e *Matrix) IsInSpan(x matrix.Row) bool {
	for _, v := range e.Space {
		if v == x[0] {
			return true
		}
	}

	return false
}

// FullyDefined returns true if the assertions made give a fully defined matrix.
func (e *Matrix) FullyDefined() bool {
	return len(e.Space) == 256
}

// Matrix returns the matrix representation of A.
func (e *Matrix) Matrix() matrix.Matrix {
	out := matrix.Matrix{}

	for i := uint(0); i < 8; i++ {
		out = append(out, matrix.Row{e.Space[1<<i]})
	}

	return out.Transpose()
}

// Dup returns a duplicate of e.
func (e *Matrix) Dup() *Matrix {
	f := NewMatrix()

	for k, v := range e.Space {
		f.Space[k] = v
	}

	return f
}
