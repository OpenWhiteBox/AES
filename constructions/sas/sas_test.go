package sas

import (
	"testing"

	"bytes"
	"crypto/rand"
)

func Example_encrypt() {
	constr := GenerateKeys(rand.Reader) // crypto/rand.Reader

	dst, src := make([]byte, 16), make([]byte, 16)
	constr.Encrypt(dst, src)
}

func TestEncrypt(t *testing.T) {
	constr := GenerateKeys(rand.Reader)

	in := make([]byte, 16)
	out := make([]byte, 16)
	out2 := make([]byte, 16)

	constr.Encrypt(out, in)
	constr.Decrypt(out2, out)

	if bytes.Compare(in, out2) != 0 {
		t.Fatalf("Correctness property is not satisfied.")
	}
}
