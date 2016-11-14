// Command encrypt reads a block from the command line, loads the serialized
// white-box from disk, and encrypts the block with it. The encrypted block is
// output.
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/OpenWhiteBox/AES/constructions/full"
)

var hexBlock = flag.String("block", "", "A hex-encoded 128-bit block to encrypt.")

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

	// Read construction from disk and parse it into something usable.
	data, err := ioutil.ReadFile("./constr.txt")
	if err != nil {
		log.Fatal(err)
	}
	constr, err := full.Parse(data)
	if err != nil {
		log.Fatal(err)
	}

	// Encrypt block in-place, and print as hex.
	constr.Encrypt(block, block)
	fmt.Printf("%x\n", block)
}
