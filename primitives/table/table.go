// Package table defines interfaces to be implemented by tabular functions.
// A table maps one (or sometimes two) byte(s) to another primitive type (nibble, byte, word, ...).  They're not
// necessarily injective or surjective.  A series of Byte tables can be composed indefinitely, but a Nibble, Word, or
// Block table will terminate the series.
package table

type Nibble interface {
	Get(i byte) byte // Get(i byte) nibble
}

type Byte interface {
	Get(i byte) byte
}

type Word interface {
	Get(i byte) [4]byte
}

type Block interface {
	Get(i byte) [16]byte
}

type DoubleToByte interface {
	Get(i [2]byte) byte
}

type DoubleToWord interface {
	Get(i [2]byte) [4]byte
}

// ComposedBytes converts an array of Byte tables into one by chaining them. Tables are chained in REVERSE order than
// they would be in function composition notation.
//
// Example:
//   f := ComposedBytes([]Byte{A, B, C}) // f(x) = C(B(A(x)))
type ComposedBytes []Byte

func (cb ComposedBytes) Get(i byte) byte {
	for j, _ := range cb {
		i = cb[j].Get(i)
	}

	return i
}

// IdentityByte is the identity operation on bytes.
type IdentityByte struct{}

func (ib IdentityByte) Get(i byte) byte { return i }

// ComposedToWord composes a Byte table and a Word table, giving a new Word table.
type ComposedToWord struct {
	Heads Byte
	Tails Word
}

func (cw ComposedToWord) Get(i byte) [4]byte {
	return cw.Tails.Get(cw.Heads.Get(i))
}

// ComposedToBlock composes a Byte table and a Block table, giving a new Block table.
type ComposedToBlock struct {
	Heads Byte
	Tails Block
}

func (cb ComposedToBlock) Get(i byte) [16]byte {
	return cb.Tails.Get(cb.Heads.Get(i))
}

// InvertibleTable is a sloppy way to invert permutation tables that aren't encodings for some reason.
type InvertibleTable Byte

func Invert(it InvertibleTable) InvertibleTable {
	out := make([]byte, 256)

	for i := 0; i < 256; i++ {
		out[it.Get(byte(i))] = byte(i)
	}

	return InvertibleTable(ParsedByte(out))
}
