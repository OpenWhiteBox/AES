package common

import (
	"testing"
)

func TestTyiTable(t *testing.T) {
	in := [16]byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}
	out := [16]byte{95, 114, 100, 21, 87, 245, 188, 146, 247, 190, 59, 41, 29, 185, 249, 26}

	a, b, c, d := TyiTable(0), TyiTable(1), TyiTable(2), TyiTable(3)
	cand := [16]byte{}

	for i := 0; i < 16; i += 4 {
		e, f, g, h := a.Get(in[i+0]), b.Get(in[i+1]), c.Get(in[i+2]), d.Get(in[i+3])

		cand[i+0] = e[0] ^ f[0] ^ g[0] ^ h[0]
		cand[i+1] = e[1] ^ f[1] ^ g[1] ^ h[1]
		cand[i+2] = e[2] ^ f[2] ^ g[2] ^ h[2]
		cand[i+3] = e[3] ^ f[3] ^ g[3] ^ h[3]
	}

	if out != cand {
		t.Fatalf("Real disagrees with result! %v != %v", out, cand)
	}
}
