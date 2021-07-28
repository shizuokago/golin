# Where is the command package?

golin command(golin.go[package main]) is in the v2 directory.

## Why?

If golin.go In this directory,,,

```
$ go install github.com/shizuokago/golin/v2/_cmd/golin@latest
```

when you do this,,,

```
but does not contain package github.com/shizuokago/golin/v2/_cmd/golin
```

you get this error. I am avoiding this.

I do not know if this is a formal response.

