package sas

import (
	"testing"

	"bytes"
)

func TestEncrypt(t *testing.T) {
	constr := GenerateKeys()

	in := make([]byte, 16)
	out := make([]byte, 16)
  out2 := make([]byte, 16)

	constr.Encrypt(out, in)
  constr.Decrypt(out2, out)

	if bytes.Compare(in, out2) != 0 {
		t.Fatalf("Correctness property is not satisfied.")
	}
}
