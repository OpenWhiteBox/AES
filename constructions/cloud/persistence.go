package cloud

import (
	"errors"

	"github.com/OpenWhiteBox/AES/primitives/table"

	"github.com/OpenWhiteBox/AES/constructions/common"
)

const fullSize = 65536 + 15728640

func (constr *Construction) Serialize() []byte {
	out, base := make([]byte, len(*constr)*fullSize), 0

	for _, M := range *constr {
		base += common.SerializeBlockMatrix(out[base:], M.Slices, M.XORs)
	}

	return out
}

func Parse(in []byte) (constr Construction, err error) {
	rest := in[:]

	var (
		slices [16]table.Block
		xors   common.ByteXORTables
	)

	for rest != nil && len(rest) > 0 {
		slices, xors, rest = common.ParseBlockByteMatrix(rest)
		constr = append(constr, Matrix{slices, xors})
	}

	if rest == nil {
		err = errors.New("Parsing the key failed!")
	}

	return
}
