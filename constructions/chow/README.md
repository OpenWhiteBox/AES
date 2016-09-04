Chow's White-box AES Construction
---------------------------------

Broadly, Chow's construction works by taking a normal AES key and converting the encryption algorithm into a series of
table lookups. The table lookups are then randomized such that this randomness eventually cancels out and gives a
correct AES encryption of a plaintext, without leaking the key. We can also modify the white-box key such that the
function isn't exactly ct = AES(pt), but a masked or encoded function like ct' = Q(AES(P(pt))), where Q and P are
randomly chosen affine transformations.

We start by generating a white-boxed key:
```go
opts := common.IndependentMasks{common.RandomMask, common.RandomMask} // Random input and output masks.
constr, input, output := chow.GenerateEncryptionKeys(key, seed, opts) // key is the AES key, seed is the seed for the RNG.
```
which we can use to encrypt data, just like a normal cipher:
```go
  constr.Encrypt(dst, src)
```

Chow's white-boxes are asymmetric, meaning you have to choose whether to generate encryption or decryption keys because
encryption keys can't be used for decryption and vice versa. Above we showed encryption; decryption is similar:
```go
opts := common.IndependentMasks{common.RandomMask, common.RandomMask}
constr, input, output := chow.GenerateDecryptionKeys(key, seed, opts)
...
constr.Decrypt(dst, src)
```

There are two types of mask: `common.RandomMask` and `common.IdentityMask`. RandomMask is a random linear transformation
and IdentityMask is the identity transformation.

There are three types of ways to attach masks to the white-box: `common.IndependentMasks`, `common.SameMasks`, and
`common.MatchingMasks`. `IndependentMasks` specifies and chooses the input and output masks independently of each other.
`SameMasks` chooses a mask of the specified type and puts the same one on the input and output. `MatchingMasks` chooses
a random mask for the input and puts the inverse mask on the output.

"White-Box Cryptography and an AES Implementation" by Stanley Chow, Philip Eisen, Harold Johnson, and Paul C. Van
Oorschot, http://link.springer.com/chapter/10.1007%2F3-540-36492-7_17?LI=true

"A Tutorial on White-Box AES" by James A. Muir, https://eprint.iacr.org/2013/104.pdf
