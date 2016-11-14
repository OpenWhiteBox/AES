OpenWhiteBox AES
----------------

This repository aims to collect implementations of white-box AES constructions and their cryptanalyses. All
documentation is in godocs:
- constructions/
  - [bes/](https://godoc.org/github.com/OpenWhiteBox/AES/constructions/bes) An un-obfuscated, reference BES (Big Encryption System) implementation.
  - [chow/](https://godoc.org/github.com/OpenWhiteBox/AES/constructions/chow) Chow et al.'s white-box AES construction.
  - [full/](https://godoc.org/github.com/OpenWhiteBox/AES/constructions/full) Full construction from paper.
  - [saes/](https://godoc.org/github.com/OpenWhiteBox/AES/constructions/saes) An un-obfuscated, reference AES implementation.
  - [toy/](https://godoc.org/github.com/OpenWhiteBox/AES/constructions/toy) Toy construction from paper.
  - [xiao/](https://godoc.org/github.com/OpenWhiteBox/AES/constructions/xiao) Xiao and Lai's white-box AES construction.
- cryptanalysis/
  - [chow/](https://godoc.org/github.com/OpenWhiteBox/AES/cryptanalysis/chow) Cryptanalysis of Chow et al.'s construction.
  - [toy/](https://godoc.org/github.com/OpenWhiteBox/AES/cryptanalysis/toy) Cryptanalysis of toy construction.
  - [xiao/](https://godoc.org/github.com/OpenWhiteBox/AES/cryptanalysis/xiao) Cryptanalysis of Xiao and Lai's construction.

The "full" construction is the only white-box construction which does not have a corresponding cryptanalysis implemented
(though that doesn't mean it's secure). See example/ for code and instructions on how to use the "full" construction.
