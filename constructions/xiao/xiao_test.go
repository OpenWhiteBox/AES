package xiao

import (
	"testing"

	"github.com/OpenWhiteBox/AES/constructions/common"
	"github.com/OpenWhiteBox/AES/constructions/saes"
)

func TestTBoxMixCol(t *testing.T) {
	in := [4]byte{0x4c, 0x7c, 0x84, 0xb3}
	out := [4]byte{0xcd, 0x5d, 0xa1, 0xc9}

	baseConstr := saes.Construction{}

	left := TBoxMixCol{
		[2]common.TBox{
			common.TBox{baseConstr, 0xea, 0x00},
			common.TBox{baseConstr, 0x8d, 0x00},
		},
		Left,
	}

	right := TBoxMixCol{
		[2]common.TBox{
			common.TBox{baseConstr, 0xf5, 0x00},
			common.TBox{baseConstr, 0x2f, 0x00},
		},
		Right,
	}

	cand := left.Get([2]byte{in[0], in[1]})
	for k, v := range right.Get([2]byte{in[2], in[3]}) {
		cand[k] ^= v
	}

	if out != cand {
		t.Fatalf("TBoxMixCol does not calculate T-Box composed with MC!\nReal: %x\nCand: %x\n", out, cand)
	}
}
