// Command generate_key reads a key from the command line and generates a random
// white-box construction for this key. The public white-box is written to
// constr.txt, and the private masks are written to constr.key.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"io/ioutil"
	"log"

	"github.com/OpenWhiteBox/AES/constructions/full"
)

var hexKey = flag.String("key", "", "A hex-encoded 128-bit AES key.")

func main() {
	flag.Parse()
	key, err := hex.DecodeString(*hexKey)
	if err != nil {
		log.Println(err)
		flag.PrintDefaults()
		return
	} else if len(key) != 16 {
		log.Println("Key must be 128 bits.")
		flag.PrintDefaults()
		return
	}

	// GenerateKey is deterministic, so we need to sample a small amount of
	// randomness to get a random white-box construction.
	seed := make([]byte, 16)
	rand.Read(seed)

	// This generates the white-box construction. inputMask and outputMask are
	// the random affine transformations on the input and output of constr.
	constr, inputMask, outputMask := full.GenerateKeys(key, seed)

	// Write the public white-box to disk.
	ioutil.WriteFile("./constr.txt", constr.Serialize(), 0777)

	// Write the private input and output mask to disk.
	buff := make([]byte, 0)
	buff = append(buff, key...)

	for _, row := range inputMask.Forwards {
		buff = append(buff, row...)
	}
	buff = append(buff, inputMask.BlockAdditive[:]...)

	for _, row := range outputMask.Forwards {
		buff = append(buff, row...)
	}
	buff = append(buff, outputMask.BlockAdditive[:]...)

	ioutil.WriteFile("./constr.key", buff, 0777)
}
