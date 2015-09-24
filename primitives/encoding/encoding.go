package encoding

type Nibble interface {
	Encode(i byte) byte
	Decode(i byte) byte
}

type Byte interface {
	Encode(i byte) byte
	Decode(i byte) byte
}

type Word interface {
	Encode(i uint32) uint32
	Decode(i uint32) uint32
}

type IdentityByte struct{}

func (ib IdentityByte) Encode(i byte) byte { return i }
func (ib IdentityByte) Decode(i byte) byte { return i }

type IdentityWord struct{}

func (iw IdentityWord) Encode(i uint32) uint32 { return i }
func (iw IdentityWord) Decode(i uint32) uint32 { return i }

type Shift uint

func (s Shift) Encode(i byte) byte { return (i + byte(s)) % 16 }
func (s Shift) Decode(i byte) byte { return (i + 16 - byte(s)) % 16 }

type ConcatenatedByte struct {
	Left, Right Nibble
}

func (cb ConcatenatedByte) Encode(i byte) byte {
	return (cb.Left.Encode(i>>4) << 4) | cb.Right.Encode(i&0xf)
}

func (cb ConcatenatedByte) Decode(i byte) byte {
	return (cb.Left.Decode(i>>4) << 4) | cb.Right.Decode(i&0xf)
}

type ConcatenatedWord struct {
	A, B, C, D Byte
}

func (cw ConcatenatedWord) Encode(i uint32) uint32 {
	return uint32(cw.A.Encode(byte(i>>24)))<<24 |
		uint32(cw.B.Encode(byte(i>>16)))<<16 |
		uint32(cw.C.Encode(byte(i>>8)))<<8 |
		uint32(cw.D.Encode(byte(i)))
}

func (cw ConcatenatedWord) Decode(i uint32) uint32 {
	return uint32(cw.A.Decode(byte(i>>24)))<<24 |
		uint32(cw.B.Decode(byte(i>>16)))<<16 |
		uint32(cw.C.Decode(byte(i>>8)))<<8 |
		uint32(cw.D.Decode(byte(i)))
}

type ForLocation struct {
	Position, SubPosition int
}

func (fl ForLocation) Encode(i byte) byte { return (i ^ byte(fl.Position+fl.SubPosition)) & 0xf }
func (fl ForLocation) Decode(i byte) byte { return (i ^ byte(fl.Position+fl.SubPosition)) & 0xf }

func WordEncodingForLocation(pos int) Word {
	return ConcatenatedWord{
		ConcatenatedByte{
			ForLocation{pos, 0},
			ForLocation{pos, 1},
		},
		ConcatenatedByte{
			ForLocation{pos, 2},
			ForLocation{pos, 3},
		},
		ConcatenatedByte{
			ForLocation{pos, 4},
			ForLocation{pos, 5},
		},
		ConcatenatedByte{
			ForLocation{pos, 6},
			ForLocation{pos, 7},
		},
	}
}
