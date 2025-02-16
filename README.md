# typist

This is a tiny program to make my Oki Microline dot matrix printer work like an electric typewriter -- [written up on my blog](https://george.honeywood.org.uk/blog/typewriter-five-ways/).

The main function is automatic scrolling up/down -- so that you can see what you've typed. Without this program whatever you've written on the current line will be hidden behind the print head.
Typist also handles some terminal escape sequences, backspace and the arrow keys, translating them into sensible print head movements.

To build:

```bash
$ go build typist.go
```

To use as an interative typewriter:

```bash
$ ./typist
```

To use as a teletype:

```bash
TERM=lp /bin/sh -i 2>&1 | ./typist
```
