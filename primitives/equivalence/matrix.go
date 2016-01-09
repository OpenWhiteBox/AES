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

func NewMatrix() Matrix {
	m := Matrix{
		Space: make(map[byte]byte),
	}
	m.Space[0x00] = 0x00

	return m
}

func (e Matrix) Dup() Matrix {
	f := NewMatrix()

	for k, v := range e.Space {
		f.Space[k] = v
	}

	return f
}

func (e Matrix) NovelInput() matrix.Row {
	for x := 1; x < 256; x++ {
		_, ok := e.Space[byte(x)]
		if !ok {
			return matrix.Row{byte(x)}
		}
	}

	return nil
}

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

func (e Matrix) Matrix() matrix.Matrix {
	out := matrix.Matrix{}

	for i := uint(0); i < 8; i++ {
		out = append(out, matrix.Row{e.Space[1<<i]})
	}

	return out.Transpose()
}

func (e Matrix) IsInSpan(x matrix.Row) bool {
	for _, v := range e.Space {
		if v == x[0] {
			return true
		}
	}

	return false
}

func (e Matrix) Assert(in, out matrix.Row) (learned bool) {
	v, ok := e.Space[in[0]]

	if ok {
		if v == out[0] {
			return false
		} else {
			panic("Inconsistency!")
		}
	}

	f := e.Dup()
	for k, v := range f.Space {
		e.Space[k^in[0]] = v ^ out[0]
	}

	return true
}

func (e Matrix) FullyDefined() bool {
  for x := 0; x < 256; x++ {
    _, ok := e.Space[byte(x)]
    if !ok {
      return false
    }
  }

  return true
}
