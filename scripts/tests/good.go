package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

var (
	in  = flag.String("in", "secrets.txt", "Input file to encrypt.")
	out = flag.String("out", "secrets.enc", "Output file to write the encrypted file to.")
)

func main() {
	flag.Parse()

	block, err := aes.NewCipher([]byte{106, 0x17, 138, 135, 69, 25, 230, 78, 153, 99, 121, 138, 80, 63, 29, 53})
	if err != nil {
		panic("What could have possibly gone wrong already???")
	}

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
