// Command decrypt reads a block from the command line, loads the white-box
// private key from disk, and decrypts the block. The decrypted block is output.
package main

import (
	"crypto/aes"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/OpenWhiteBox/primitives/encoding"
	"github.com/OpenWhiteBox/primitives/matrix"
)

const keySize = 16 * (1 + 128 + 1 + 128 + 1)

var hexBlock = flag.String("block", "", "A hex-encoded 128-bit block to decrypt.")

func main() {
	flag.Parse()
	block, err := hex.DecodeString(*hexBlock)
	if err != nil {
		log.Println(err)
		flag.PrintDefaults()
		return
	} else if len(block) != 16 {
		log.Println("Block must be 128 bits.")
		flag.PrintDefaults()
		return
	}

	// Read key from disk and parse it.
	data, err := ioutil.ReadFile("./constr.key")
	if err != nil {
		log.Fatal(err)
	} else if len(data) != keySize {
		log.Fatalf("key wrong size: %v (should be %v)", len(data), keySize)
	}

	var key []byte
	inputLinear, outputLinear := matrix.Matrix{}, matrix.Matrix{}
	inputConst, outputConst := [16]byte{}, [16]byte{}

	key, data = data[:16], data[16:]
	for i := 0; i < 128; i++ {
		inputLinear, data = append(inputLinear, data[:16]), data[16:]
	}
	copy(inputConst[:], data)
	data = data[16:]
	for i := 0; i < 128; i++ {
		outputLinear, data = append(outputLinear, data[:16]), data[16:]
	}
	copy(outputConst[:], data)

	inputMask := encoding.NewBlockAffine(inputLinear, inputConst)
	outputMask := encoding.NewBlockAffine(outputLinear, outputConst)

	// Decrypt block and print as hex.
	temp := [16]byte{}
	copy(temp[:], block)

	temp = outputMask.Decode(temp)

	c, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}
	c.Decrypt(temp[:], temp[:])

	temp = inputMask.Decode(temp)
	fmt.Printf("%x\n", temp)
}
