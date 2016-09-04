package bes

import (
	"testing"

	"bytes"
	"crypto/rand"

	"github.com/OpenWhiteBox/primitives/gfmatrix"

	"github.com/OpenWhiteBox/AES/constructions/saes"
)

func testinit() (gfmatrix.Row, []byte, saes.Construction) {
	a := make([]byte, 16)
	rand.Read(a)

	b := make([]byte, 16)
	copy(b, a)

	return Expand(a), b, saes.Construction{}
}

func TestExpandContract(t *testing.T) {
	v := make([]byte, 16)
	rand.Read(v)

	u := Expand(v)
	w := Contract(u)

	if !bytes.Equal(v, w) {
		t.Fatal("Contract is not inverse of expand!")
	}
}

func TestSubBytes(t *testing.T) {
	cand, real, constr := testinit()

	for pos := 0; pos < 128; pos++ {
		cand[pos] = cand[pos].Invert()
	}
	cand = subBytes.Mul(cand).Add(subBytesConst)

	constr.SubBytes(real)

	if !Expand(real).Equals(cand) {
		t.Fatal("SubBytes wrong")
	}
}

func TestShiftRows(t *testing.T) {
	cand, real, constr := testinit()

	cand = shiftRows.Mul(cand)
	constr.ShiftRows(real)

	if !Expand(real).Equals(cand) {
		t.Fatal("ShiftRows wrong")
	}
}

func TestMixColumns(t *testing.T) {
	cand, real, constr := testinit()

	cand = mixColumns.Mul(cand)
	constr.MixColumns(real)

	if !Expand(real).Equals(cand) {
		t.Fatal("MixColumns wrong")
	}
}

func TestEncrypt(t *testing.T) {
	key := make([]byte, 16)
	rand.Read(key)

	constr1 := saes.Construction{Key: key}
	constr2 := Construction{Key: Expand(key)}

	in, out := make([]byte, 16), make([]byte, 16)
	rand.Read(in)
	constr1.Encrypt(out, in)

	if !constr2.encrypt(Expand(in)).Equals(Expand(out)) {
		t.Fatal("BES Encrypt didn't agree with AES Encrypt!")
	}
}

func TestDecrypt(t *testing.T) {
	key := make([]byte, 16)
	rand.Read(key)

	constr1 := saes.Construction{Key: key}
	constr2 := Construction{Key: Expand(key)}

	in, out := make([]byte, 16), make([]byte, 16)
	rand.Read(in)
	constr1.Decrypt(out, in)

	if !constr2.decrypt(Expand(in)).Equals(Expand(out)) {
		t.Fatal("BES Decrypt didn't agree with AES Decrypt!")
	}
}
