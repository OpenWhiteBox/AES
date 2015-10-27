White-Box File Encryption Demo
------------------------------

First, generate a white-boxed AES key:

```
go run generate-key.go -out key.txt
```
> ```
> $ go run generate-key.go -out key.txt
> Key: 33e3425b4313bdb50ef398b16d7dd0e5
> $
> ```

It will output the white-boxed key to the file `key.txt` and the real AES key to
the terminal.

Then, we can encrypt files with only the white-boxed key:

```
go run encrypt.go -key key.txt -in secrets.txt -out secrets.txt.enc
```
> ```
> $ go run encrypt.go -key key.txt -in secrets.txt -out secrets.txt.enc
> IV: bd4464faaa996293b3f49d3b19359820
> Done!
> $
> ```

Which, given the IV, can be decrypted with the master key or with another
white-box built for decryption:

```
openssl enc -d -aes-128-cbc -K 33e3425b4313bdb50ef398b16d7dd0e5 -iv bd4464faaa996293b3f49d3b19359820 -in secrets.txt.enc
```
> ```
> $ openssl enc -d -aes-128-cbc -K 33e3425b4313bdb50ef398b16d7dd0e5 \
> > -iv bd4464faaa996293b3f49d3b19359820 -in secrets.txt.enc
> Hello, World!
> $
> ```
