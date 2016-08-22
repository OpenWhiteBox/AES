// Tables with encoded inputs and outputs.
package encoding

import (
	"github.com/OpenWhiteBox/AES/primitives/table"
)

type NibbleTable struct {
	In     Byte
	Out    Nibble
	Hidden table.Nibble
}

func (nt NibbleTable) Get(i byte) byte {
	return nt.Out.Encode(nt.Hidden.Get(nt.In.Decode(i)))
}

type ByteTable struct {
	In     Byte
	Out    Byte
	Hidden table.Byte
}

func (bt ByteTable) Get(i byte) byte {
	return bt.Out.Encode(bt.Hidden.Get(bt.In.Decode(i)))
}

type WordTable struct {
	In     Byte
	Out    Word
	Hidden table.Word
}

func (wt WordTable) Get(i byte) [4]byte {
	return wt.Out.Encode(wt.Hidden.Get(wt.In.Decode(i)))
}

type BlockTable struct {
	In     Byte
	Out    Block
	Hidden table.Block
}

func (bt BlockTable) Get(i byte) [16]byte {
	return bt.Out.Encode(bt.Hidden.Get(bt.In.Decode(i)))
}

type DoubleToWordTable struct {
	In     Double
	Out    Word
	Hidden table.DoubleToWord
}

func (dtwt DoubleToWordTable) Get(i [2]byte) [4]byte {
	return dtwt.Out.Encode(dtwt.Hidden.Get(dtwt.In.Decode(i)))
}
