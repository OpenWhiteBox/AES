package encoding

import (
	"github.com/OpenWhiteBox/AES/primitives/table"
)

// NibbleTable implements the table.Nibble interface over another Nibble table with a Byte encoding on its input and a
// Nibble encoding on its output.
type NibbleTable struct {
	In     Byte
	Out    Nibble
	Hidden table.Nibble
}

func (nt NibbleTable) Get(i byte) byte {
	return nt.Out.Encode(nt.Hidden.Get(nt.In.Decode(i)))
}

// ByteTable implements the table.Byte interface over another Byte table with Byte encodings on its input and output.
type ByteTable struct {
	In     Byte
	Out    Byte
	Hidden table.Byte
}

func (bt ByteTable) Get(i byte) byte {
	return bt.Out.Encode(bt.Hidden.Get(bt.In.Decode(i)))
}

// WordTable implements the table.Word interface over another Word table with a Byte encoding on its input and a Word
// encoding on its output.
type WordTable struct {
	In     Byte
	Out    Word
	Hidden table.Word
}

func (wt WordTable) Get(i byte) [4]byte {
	return wt.Out.Encode(wt.Hidden.Get(wt.In.Decode(i)))
}

// BlockTable implements the table.Block interface over another Block table with a Byte encoding on its input and a
// Block encoding on its output.
type BlockTable struct {
	In     Byte
	Out    Block
	Hidden table.Block
}

func (bt BlockTable) Get(i byte) [16]byte {
	return bt.Out.Encode(bt.Hidden.Get(bt.In.Decode(i)))
}

// DoubleToByteTable implements the table.DoubleToByte interface over another DoubleToByte table with a Double encoding
// on its input and a Byte encoding on its output.
type DoubleToByteTable struct {
	In     Double
	Out    Byte
	Hidden table.DoubleToByte
}

func (dtbt DoubleToByteTable) Get(i [2]byte) byte {
	return dtbt.Out.Encode(dtbt.Hidden.Get(dtbt.In.Decode(i)))
}

// DoubleToWordTable implements the table.DoubleToWord interface over another DoubleToWord table with a Double encoding
// on its input and a Word encoding on its output.
type DoubleToWordTable struct {
	In     Double
	Out    Word
	Hidden table.DoubleToWord
}

func (dtwt DoubleToWordTable) Get(i [2]byte) [4]byte {
	return dtwt.Out.Encode(dtwt.Hidden.Get(dtwt.In.Decode(i)))
}
