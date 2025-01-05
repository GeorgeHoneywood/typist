package main

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
)

func main() {
	// open input file
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// open output file
	fo, err := os.Create("/dev/usb/lp0")
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	lastWrite := time.Now()

	buf := make([]byte, 1)
	for {
		time.Since(lastWrite)
		// read a chunk
		i, err := os.Stdin.Read(buf)
		fmt.Printf("read %d bytes\n", i)
		if err != nil {
			panic(err)
		}

		// check for eof
		if i == 0 {
			fmt.Printf("got eof\n")
			break
		}

		if buf[0] == 3 {
			fmt.Printf("got ctrl-c\n")
			break
		}

		if buf[0] == 13 {
			fmt.Printf("got enter\n")
			fo.Write([]byte{10})
			continue
		}

		fmt.Printf("writing %d\n", buf[0])

		// write a chunk
		_, err = fo.Write(buf)
		if err != nil {
			panic(err)
		}

		// fo.Write([]byte{27, 74, 216})
		fo.Write([]byte{27, 93})

	}
}
