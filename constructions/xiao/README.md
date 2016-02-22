Xiao-Lai's White-Box AES
------------------------

Xiao is an implementation of the white-box AES scheme described by Xiao and Lai in the paper *A secure implementation of
white-box AES*.

The interface here is exactly the same as in Chow et al.'s construction--see the documentation there.


## Performance

Benchmarks are run with `go test -bench Benchmark.*`.   Key generation takes about 43 seconds and encryption takes
200,000ns/op.
