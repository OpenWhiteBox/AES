package toy

import (
	"testing"

	"bytes"
	"crypto/rand"

	"github.com/OpenWhiteBox/AES/constructions/toy"
)

func TestRecoverKey(t *testing.T) {
	key := make([]byte, 16)
	rand.Read(key)

	constr, _, _ := toy.GenerateKeys(key, key)

	cand := RecoverKey(&constr)
	if !bytes.Equal(cand, key) {
		t.Fatalf("Recovered wrong key!\nreal=%x\ncand=%x", key, cand)
	}
}
