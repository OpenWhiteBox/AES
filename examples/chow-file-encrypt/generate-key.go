package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/OpenWhiteBox/AES/constructions/chow"
)

var out = flag.String("out", "key.txt", "Where to write the key.")

func main() {
	flag.Parse()

	// Generate AES key and seed for the white-box generator.
	key, seed := make([]byte, 16), make([]byte, 16)
	rand.Read(key)
	rand.Read(seed)

	fmt.Printf("Key: %x\n", key)

	// Create a white-box version of the above key.
	constr, _, _ := chow.GenerateEncryptionKeys(key, seed, chow.SameMasks(chow.IdentityMask))
	keyData := constr.Serialize()

	ioutil.WriteFile(*out, keyData, os.ModePerm)
}
