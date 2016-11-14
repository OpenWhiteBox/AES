Example
-------

For instructions on how to install Golang, [start here](https://golang.org/).

White-box cryptography tries to find implementations of keyed block ciphers,
such that the key is hard to extract. Algorithms for generating instances of a
white-box construction have a public output and a private output. The public
output is this special implementation of the block cipher, and the private
output is key and possibly some auxiliary information to help decode the
instance's output. To generate an instance, run:

``` bash
$ # (Replace 000...000 with any hex-encoded 128-bit key.)
$ go run generate_key.go -key 0123456789abcdeffedcba9876543210
$
```

This script generates a white-box instance and writes the public output to
`constr.txt` and the private output to `constr.key`. This particular
construction necessarily puts random affine transformations on the input and
output of the block cipher in `constr.txt`. These transformations are saved in
`constr.key`.

To encrypt a block of data with the cipher, run:

``` bash
$ # (Again, any 128-bit hex-encoded string.)
$ go run encrypt.go -block 000000000000000000000000deadbeef
dfc967b77a809c926075441565cbc3e3
$
```

and to decrypt it,

``` bash
$ go run decrypt.go -block dfc967b77a809c926075441565cbc3e3
000000000000000000000000deadbeef
$
```

The script `encrypt.go` only accesses `constr.txt` and applies the white-box
instance to its input. However, `decrypt.go` only accesses `constr.key`, and
undoes the affine transformations from `encrypt.go` in addition to standard AES
decryption of the input block.

Note that both scripts are deterministic, whereas `generate_key.go` is not, and
that different white-box instances may give different encryptions of the same
data even though they're built with the same key.

``` bash
$ go run generate_key.go -key 0123456789abcdeffedcba9876543210
$ go run encrypt.go -block 000000000000000000000000deadbeef
33cb7aeb14db2329ffebfd003d3fd076
$ go run decrypt.go -block 33cb7aeb14db2329ffebfd003d3fd076
000000000000000000000000deadbeef

$ go run generate_key.go -key 0123456789abcdeffedcba9876543210
$ go run encrypt.go -block 000000000000000000000000deadbeef
5d088966051465354ac0de72c33849f6
$ go run decrypt.go -block 5d088966051465354ac0de72c33849f6
000000000000000000000000deadbeef
$
```
