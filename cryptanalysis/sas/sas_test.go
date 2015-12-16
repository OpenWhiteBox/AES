package sas

import (
	"testing"

	"github.com/OpenWhiteBox/AES/constructions/sas"

	"fmt"
)

func TestRecoverLastSBox(t *testing.T) {
	constr := sas.GenerateKeys()
	fmt.Println(RecoverLastSBox(constr, 4))
}
