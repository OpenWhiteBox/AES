package xiao

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/OpenWhiteBox/AES/constructions/common"
	"github.com/OpenWhiteBox/AES/constructions/xiao"
)

func TestRecoverKey(t *testing.T) {
	key := make([]byte, 16)
	rand.Read(key)

	constr, _, _ := xiao.GenerateEncryptionKeys(
		key, key, common.IndependentMasks{common.RandomMask, common.RandomMask},
	)

	serialized := constr.Serialize()
	constr2, err := xiao.Parse(serialized)
	if err != nil {
		t.Fatalf("xiao.Parse returned error: %v", err)
	}

	cand := RecoverKey(&constr2)

	if !bytes.Equal(key, cand) {
		t.Fatal("Generated key does not equal recovered key!")
	}
}
