// Cryptanalysis of SAS block ciphers.
package sas

import (
	"fmt"

	"crypto/rand"

	"github.com/OpenWhiteBox/AES/primitives/number"
	matrix "github.com/OpenWhiteBox/AES/primitives/matrix2"
	"github.com/OpenWhiteBox/AES/primitives/table"

	"github.com/OpenWhiteBox/AES/constructions/sas"
)

func equal(left, right matrix.Row) bool {
	for j := 0; j < 256; j++ {
		if left[j] != right[j] {
			return false
		}
	}

	return true
}

func RecoverLastSBox(constr sas.Construction, sbox int) table.Byte {
	m := matrix.GenerateEmpty(256)

	round := 0
	for round < 256 {
		pts := GenerateRandomPlaintexts(sbox)
		ct := make([]byte, 16)

		for _, pt := range pts {
			constr.Encrypt(ct, pt) // C[..]PC[..] -(S)-> C[..]PC[..] -(A)->   D[..]   -(S)-> D[..]
			constr.Encrypt(ct, ct) //    D[..]    -(S)->    D[..]    -(A)->   B[..]   -(S)-> x[..]

			m[round][ct[sbox]] = m[round][ct[sbox]].Add(number.ByteFieldElem(0x01))
		}

		bad := false
		for i := 0; i < round; i++ {
			if equal(m[round], m[i]) {
				m[round] = matrix.Row(make([]number.ByteFieldElem, 256))
				bad = true
				break
			}
		}

		if !bad {
			round++
		}
	}

	b := m.NullSpace()

	for true {
		v := RandomLinearCombination(b)

		if IsPermutation(v) {
			fmt.Println(v)
			fmt.Println(m.Mul(v))
			break
		}
	}

	return table.IdentityByte{}
}

func RandomLinearCombination(basis []matrix.Row) matrix.Row {
	coeffs := make([]byte, len(basis))
	rand.Read(coeffs)

	v := matrix.Row(make([]number.ByteFieldElem, basis[0].Size()))

	for i, c_i := range coeffs {
		v = v.Add(basis[i].ScalarMul(number.ByteFieldElem(c_i)))
	}

	return v
}

func IsPermutation(v matrix.Row) bool {
	sums := [256]int{}
	for _, v_i := range v {
		sums[byte(v_i)]++
	}

	for _, x := range sums {
		if x != 1 {
			return false
		}
	}
	return true
}

func GenerateRandomPlaintexts(sbox int) (out [][]byte) {
	master := make([]byte, 16)
	rand.Read(master)

	for i := 0; i < 256; i++ {
		pt := make([]byte, 16)
		copy(pt, master)

		pt[sbox] = byte(i)

		out = append(out, pt)
	}

	return
}
