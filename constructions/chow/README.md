Chow's White-Box AES
--------------------

Currently unfinished:

- Generating matrices with blocks of full rank.

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
implementations take about 170ns/op.  White-Boxing an AES call seems makes it 2 to 3 orders of magnitude slower, at
most.
