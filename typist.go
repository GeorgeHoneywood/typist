package main

// very useful control code reference
// my printer is in epson emulation mode, so most of the ESC/P commands should work
// https://www.whitefiles.org/dta/pgs/cc03c_prntr_cds.pdf

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
)

func reader(char chan byte) {
	buf := make([]byte, 1)
	for {
		i, err := os.Stdin.Read(buf)
		fmt.Printf("read %d bytes\r\n", i)
		if err != nil {
			panic(err)
		}

		char <- buf[0]
	}
}

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	lp, err := os.Create("/dev/usb/lp0")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := lp.Close(); err != nil {
			panic(err)
		}
	}()

	scrolledUp := false

	pipe := make(chan byte)
	go reader(pipe)

	// ESC/P typewriter mode not emulated :(
	// fo.Write([]byte{27, 105, 1})

	// the idea here is that we can't read characters immediately behind the
	// printer head, so we need to scroll down to read them. we can't do this
	// immediately, so we wait for a timeout
readLoop:
	for {
		select {
		case char := <-pipe:
			if char == 3 {
				fmt.Printf("got ctrl+c\r\n")
				break readLoop
			}

			if char == 13 {
				fmt.Printf("got enter\r\n")
				// line feed
				lp.Write([]byte{10})
				continue
			}

			fmt.Printf("writing %d\r\n", char)

			if scrolledUp {
				// if scrolled up, scroll back down
				fmt.Print("scrolling back down\r\n")

				// 27 106 nn: Perform nn/216 inch reverse line feed
				lp.Write([]byte{27, 106, 216})
				lp.Write([]byte{27, 106, 216})
				scrolledUp = false
			}

			// write a character
			_, err = lp.Write([]byte{char})
			if err != nil {
				panic(err)
			}

		case <-time.After(time.Second * 2):
			if scrolledUp {
				break
			}
			fmt.Print("timed out, scrolling up\r\n")

			// 27 74 nn: Perform nn/216 inch line feed
			lp.Write([]byte{27, 74, 216})
			lp.Write([]byte{27, 74, 216})
			scrolledUp = true
		}

	}
}
