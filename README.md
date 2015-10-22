White-Box Crypto
----------------

This repository aims to collect implementations of white-box AES constructions and their cryptanalyses.

> The challenge that white-box cryptography aims to address is to implement a cryptographic algorithm in software in
> such a way that cryptographic assets remain secure even when subject to white-box attacks.

The project structure is as follows:

- **constructions/** - Each new white-box construction is given it's own subfolder in this directory.
  - **chow/** - The white-box designed by S. Chow et al. in *White-Box Cryptography and an AES Implementation*
  - **saes/** - A completely textbook implementation of AES (aka, Standard AES) with no optimizations.  (Useful for
    playing with alternate representations of AES.)
  - **test/** - Not a construction--directory of AES test vectors.
- **cryptanalysis/** - Contains cryptanalyses of provided white-box constructions.
  - **chow/** - An implementation of the BGE attack on S. Chow et al.'s white-box.
- **examples/** - Code examples for using the white-box constructions.
- **scripts/** - Tools and build scripts.


### References

- http://www.whiteboxcrypto.com/
