package main

import (
	"crypto/cipher"
	"crypto/rand"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/OpenWhiteBox/AES/constructions/chow"
)

var (
	key = flag.String("key", "key.txt", "White-Box AES Key to encrypt with.")
	in  = flag.String("in", "secrets.txt", "Input file to encrypt.")
	out = flag.String("out", "secrets.enc", "Output file to write the encrypted file to.")
)

func main() {
	flag.Parse()

	// Load in key data and initialize the block cipher.
	keyData, err := ioutil.ReadFile(*key)
	if err != nil {
		panic(err)
	}

	block := chow.Parse(keyData)

	// Put block cipher in CBC mode.
	iv := make([]byte, 16)
	rand.Read(iv)

	mode := cipher.NewCBCEncrypter(block, iv)

	// Load in file to encrypt.
	data, err := ioutil.ReadFile(*in)
	if err != nil {
		panic(err)
	}

	// Create and append the padding for our file.
	padding := make([]byte, 16-len(data)%16)

	for i, _ := range padding {
		padding[i] = byte(len(padding))
	}

	data = append(data, padding...)

	// Encrypt file.
	mode.CryptBlocks(data, data)

	// Write encrypted file to disk.
	ioutil.WriteFile(*out, append(iv, data...), os.ModePerm)

	fmt.Println("Done!")
}
