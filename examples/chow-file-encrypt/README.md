White-Box File Encryption Demo
------------------------------

First, generate a white-boxed AES key:

```
go run generate-key.go -out key.txt
```

It will output the white-boxed key to the file `key.txt` and the real AES key to the terminal.

Then, we can encrypt files with the white-boxed key:

```
go run encrypt.go -key key.txt -in secrets.txt -out secrets.txt.enc
```
