Katas is a simple CLI tool to track your programming training.

```
# Installation and initialization.
$ go install github.com/go-monk/katas
$ katas -init

# Daily training.
$ katas
Kata         Last done  Done  URL
----         ---------  ----  ---
hello        never      0x    https://github.com/golang/example/tree/master/hello
helloserver  never      0x    https://github.com/golang/example/tree/master/helloserver
outyet       never      0x    https://github.com/golang/example/tree/master/outyet
----                    ----  
3                       0x
$ katas -done hello
```