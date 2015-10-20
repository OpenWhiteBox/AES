Chow's White-Box AES
--------------------

Chow is an implementation of the white-box AES scheme described by S. Chow et al. in the paper *White-Box
Cryptography and an AES Implementation*.

Broadly, it works by taking a normal AES key and stretching it out and adding randomness to computations that eventually
cancel out and allow someone to compute a correct AES encryption of a plaintext without knowing or being able to find
the key. We can also modify the white-box key such that the function isn't exactly ct = AES(pt), but a masked or
encoded function like ct' = Q(AES(P(pt))), where Q and P are randomly chosen bijections from a family.

We start by generating a white-boxed key:

```go
opts := IndependentMasks{RandomMask, RandomMask} // Random input and output masks.
constr, input, output := chow.GenerateEncryptionKeys(key, seed, opts) // key is the AES key, seed is the seed for the RNG.
```

Which we can use to encrypt data just like a normal AES cipher:

```go
constr.Encrypt(dst, src)
```

AES white-boxes are asymmetric, meaning you have to choose whether to generate encryption or decryption keys because
encryption keys can't be used for decryption and vice versa.  Above, we showed encryption; decryption is similar:

```go
opts := IndependentMasks{RandomMask, RandomMask}
constr, input, output := chow.GenerateDecryptionKeys(key, seed, opts)
..

constr.Decrypt(dst, src)
```

There are two types of mask: `RandomMask` and `IdentityMask`. `RandomMask` is a random linear bijection, `IdentityMask`
is the identity bijection.

There are three types of ways to combine masks with the `Encrypt` function: `IndependentMasks`, `SameMasks`, and
`MatchingMasks`:

- `IndependentMasks{IdentityMask, RandomMask}` - The function's input is unmasked, but the output has a random mask.
- `SameMasks(IdentityMask)` - The function is completely unmasked.
- `MatchingMasks{}` - We choose a random mask for the input and use it's inverse as the output mask.

The `constr` output of `GenerateEncryptionKeys` is compatible with `cipher.Block` so it can automatically be used with
any cipher mode Golang supports.

To mask input or recover output from a masked white-box, you multiply the input/output bit vector by the inverse of the
`input` or `output` matrix returned by `GenerateEncryptionKeys`.

```go
import (
  "github.com/OpenWhiteBox/AES/primitives/matrix"
)

// ...

inputInv, ok := inputMask.Invert() // ok should always be true, so it can be discarded.

input := []byte{...}
copy(input, inputInv.Mul(matrix.Row(input)))

// input is sent off to a server running the white-box to be encrypted/processed.
```

## Performance

Benchmarks are run with `go test -bench Benchmark.*`. One table lookup, on my computer at least *(2.4 GHz Intel Core i5,
8 GB 1600 MHz DDR3)*, takes 10ns/op. ShiftRows has no table lookups, but takes 20ns/op.


### Lookup Multiplicity

- ExpandWord has 4 table lookups
- SquashWords has 3*4*2=24 table lookups
- ExpandBlock has 16 table lookups
- SquashBlocks has 15*16*2=480 table lookups


### Function Multiplicity

- ExpandWord and SquashWord are called 9*2*(16/4)=72 times each in a call to Encrypt.
- ExpandBlock and SquashBlocks are called 2 times each.
- ShiftRows is called 10 times.


### Total Predicted Time

```
   EB       SBs        EW         SWs       SR
2*16*10 + 2*480*10 + 72*4*10 + 72*24*10 + 10*20 = 30,280ns/op
```

For each summand, number of table lookups times 10ns is an under-estimate of the actual time taken to compute the
function (according to benchmarks). The Encrypt benchmark takes about 47,000ns/op. The rest may be accounted for by
expenses associated with memory allocation and function calls.


### Context

With hardware implementations of AES, an Encrypt call can take as little at 30ns/op.  Heavily optimized software
implementations take about 170ns/op.  White-Boxing an AES call seems to make it 2 to 3 orders of magnitude slower, at
most.


## To Do

- Generate matrices with blocks of full rank.
