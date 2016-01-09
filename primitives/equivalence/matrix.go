package equivalence

import (
	"github.com/OpenWhiteBox/AES/primitives/matrix"
)

type Elem struct {
	In, Out matrix.Row
}

type Matrix struct {
	Space map[byte]byte
}

// NewMatrix returns a new, empty matrix.
func NewMatrix() Matrix {
	m := Matrix{
		Space: make(map[byte]byte),
	}
	m.Space[0x00] = 0x00

	return m
}

// Assert represents an assertion that A(in) = out. The function will panic if this is inconsistent with previous
// assertions. It it's not, it returns whether or not the assertion contained new information about A.
func (e Matrix) Assert(in, out matrix.Row) (learned bool) {
	learned = false

	f := e.Dup()
	for x, y := range f.Space {
		yGot, ok := e.Space[x^in[0]]
		yExpected := y ^ out[0]

		if ok && yGot != yExpected {
			panic("Inconsistency!")
		} else if !ok {
			e.Space[x^in[0]] = yExpected
			learned = true
		}
	}

	return
}

// NovelInput returns an x such that A(x) is unknown.
func (e Matrix) NovelInput() matrix.Row {
	for x := 1; x < 256; x++ {
		_, ok := e.Space[byte(x)]
		if !ok {
			return matrix.Row{byte(x)}
		}
	}

	return nil
}

// IsInSpan returns whether or not x is in the known span of A.
func (e Matrix) IsInSpan(x matrix.Row) bool {
	for _, v := range e.Space {
		if v == x[0] {
			return true
		}
	}

	return false
}

// Span returns an iterator for all the elements in the span of A.
func (e Matrix) Span() <-chan Elem {
	res := make(chan Elem)

	go func() {
		for k, v := range e.Space {
			res <- Elem{matrix.Row{k}, matrix.Row{v}}
		}
		close(res)
	}()

	return res
}

// FullyDefined returns true if the assertions made give a fully defined matrix.
func (e Matrix) FullyDefined() bool {
	return len(e.Space) == 256
}

// Matrix returns the matrix representation of A.
func (e Matrix) Matrix() matrix.Matrix {
	out := matrix.Matrix{}

	for i := uint(0); i < 8; i++ {
		out = append(out, matrix.Row{e.Space[1<<i]})
	}

	return out.Transpose()
}

// Dup returns a duplicate of e.
func (e Matrix) Dup() Matrix {
	f := NewMatrix()

	for k, v := range e.Space {
		f.Space[k] = v
	}

	return f
}
