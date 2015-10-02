package table

import (
	"testing"
)

type XORTable struct{}

func (xor XORTable) Get(i byte) (out byte) {
	return (i >> 4) ^ (i & 0xf)
}

type AddTable byte

func (at AddTable) Get(i byte) byte { return i + byte(at) }

type TimesTable byte

func (tt TimesTable) Get(i byte) byte { return i * byte(tt) }

type ShiftTable uint

func (st ShiftTable) Get(i byte) [4]byte {
	val := uint32(i) << uint(st)
	return [4]byte{byte(val >> 24), byte(val >> 16), byte(val >> 8), byte(val)}
}

func TestCompose(t *testing.T) {
	x := ComposedBytes{TimesTable(5), AddTable(3)}
	y := ComposedBytes{AddTable(3), TimesTable(5)}

	a := ComposedToWord{x, ShiftTable(24)}

	if x.Get(7) != ((7 * 5) + 3) {
		t.Fatalf("X's Composition is wrong.")
	}

	if y.Get(7) != ((7 + 3) * 5) {
		t.Fatalf("Y's Composition is wrong.")
	}

	temp := a.Get(7)
	val := uint32(temp[0])<<24 | uint32(temp[1])<<16 | uint32(temp[2])<<8 | uint32(temp[3])

	if val != ((7*5)+3)<<24 {
		t.Fatalf("A's Composition is wrong.")
	}

	// Incorrect table compositions should refuse to compile.
}

func TestPersistNibble(t *testing.T) {
	w := XORTable{}
	serializedW := SerializeNibble(w)
	parsedW := ParsedNibble(serializedW)

	for i := 0; i < 256; i++ {
		if w.Get(byte(i)) != parsedW.Get(byte(i)) {
			t.Fatalf("W and ParsedW disagree at point %v: %v != %v", i, w.Get(byte(i)), parsedW.Get(byte(i)))
		}
	}
}

func TestPersistByte(t *testing.T) {
	x := AddTable(3)
	serializedX := SerializeByte(x)
	parsedX := ParsedByte(serializedX)

	for i := 0; i < 256; i++ {
		if x.Get(byte(i)) != parsedX.Get(byte(i)) {
			t.Fatalf("X and ParsedX disagree at point %v: %v != %v", i, x.Get(byte(i)), parsedX.Get(byte(i)))
		}
	}
}

func TestPersistWord(t *testing.T) {
	y := ShiftTable(13)
	serializedY := SerializeWord(y)
	parsedY := ParsedWord(serializedY)

	for i := 0; i < 256; i++ {
		if y.Get(byte(i)) != parsedY.Get(byte(i)) {
			t.Fatalf("Y and ParsedY disagree at point %v: %v != %v", i, y.Get(byte(i)), parsedY.Get(byte(i)))
		}
	}
}
