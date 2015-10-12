// A table maps one byte to another primitive type (nibble, a second byte, ...).  They're not necessarily injective or
// surjective.  A series of byte tables can be composed indefinitely, but a nibble, word, or block table will terminate
// the series.
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

type ComposedBytes []Byte

func (cb ComposedBytes) Get(i byte) byte {
	for j, _ := range cb {
		i = cb[j].Get(i)
	}

	return i
}

type IdentityByte struct{}

func (ib IdentityByte) Get(i byte) byte { return i }

// ComposedToNibble isn't needed because you can use ComposedSmalls.

type ComposedToWord struct {
	Heads Byte
	Tails Word
}

func (cw ComposedToWord) Get(i byte) [4]byte {
	return cw.Tails.Get(cw.Heads.Get(i))
}

type ComposedToBlock struct {
	Heads Byte
	Tails Block
}

func (cb ComposedToBlock) Get(i byte) [16]byte {
	return cb.Tails.Get(cb.Heads.Get(i))
}

// Sloppy way to invert permutation tables that aren't encodings for some reason.
type InvertibleTable Byte

func Invert(it InvertibleTable) InvertibleTable {
	out := make([]byte, 256)

	for i := 0; i < 256; i++ {
		out[it.Get(byte(i))] = byte(i)
	}

	return InvertibleTable(ParsedByte(out))
}
