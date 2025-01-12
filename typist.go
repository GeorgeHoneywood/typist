package main

// typist.go
// reads characters from stdin and writes them to a printer
// supports basic cursor movement and line feed

// usage:
// * to run as a typewriter:
//    go run typist.go
// * or to use like a tty/teletype:
//    go build typist.go
//    TERM=dumb /bin/sh -i 2>&1 | ./typist

// very useful control code reference
// my printer is in epson emulation mode, so most of the ESC/P commands should work
// https://whitefiles.org/dta/pgs/c03c_prntr_cds.pdf

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
)

const lineFeedValue = 36

type printer struct {
	// file descriptor for the printer
	fd *os.File

	// whether we've scrolled up or not
	scrolledUp bool

	// channel for reading characters
	pipe chan byte
}

func (p *printer) readChars() {
	buf := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buf)
		fmt.Printf("read %d bytes, char \"%q\", dec \"%d\", hex \"%x\"\r\n", n, buf[0], buf[0], buf[0])
		if err != nil {
			panic(err)
		}

		p.pipe <- buf[0]
	}
}

func (p *printer) handleInput() {
	// build up escape sequences
	escapes := [2]byte{}

	// the idea here is that we can't read characters immediately behind the
	// printer head, so we need to scroll down to read them. we can't do this
	// immediately, so we wait for a timeout
readLoop:
	for {
		select {
		case char := <-p.pipe:
			fmt.Printf("processing, char \"%q\", dec \"%d\", hex \"%x\"\r\n", char, char, char)
			fmt.Printf("escapes: \"%v\"\r\n", escapes)

			if char == 3 {
				fmt.Printf("got ctrl+c\r\n")
				break readLoop
			}

			if char == 13 {
				fmt.Printf("got enter\r\n")

				// line feed by lineFeedValue/216 inches
				p.fd.Write([]byte{27, 74, lineFeedValue})

				// carriage return
				p.fd.Write([]byte{13})

				continue
			}

			if char == 27 {
				fmt.Printf("got first escape sequence char %d\r\n", char)
				escapes = [2]byte{char, 0}

				continue
			} else if escapes[0] == 27 && escapes[1] == 0 {
				fmt.Printf("got second escape sequence char %d\r\n", char)
				escapes = [2]byte{27, char}

				continue
			} else if escapes[0] == 27 && escapes[1] != 0 {
				fmt.Printf("got third escape sequence char %d\r\n", char)

				switch char {
				case 65:
					fmt.Printf("cursor up\r\n")
					p.fd.Write([]byte{27, 106, lineFeedValue})
				case 66:
					fmt.Printf("cursor down\r\n")
					p.fd.Write([]byte{27, 74, lineFeedValue})
				case 68:
					fmt.Printf("cursor left\r\n")
					p.fd.Write([]byte{8})
				case 67:
					fmt.Printf("cursor right\r\n")
					p.fd.Write([]byte{32})
				default:
					fmt.Printf("unknown escape sequence: %v\r\n", escapes)
				}
				escapes = [2]byte{0, 0}
				continue
			}

			if p.scrolledUp {
				// if scrolled up, scroll back down
				fmt.Print("scrolling back down\r\n")
				p.scroll(false)
			}

			// write a non-special character
			fmt.Printf("writing char \"%q\", dec \"%d\", hex \"%x\"\r\n", char, char, char)
			_, err := p.fd.Write([]byte{char})
			if err != nil {
				panic(err)
			}

		case <-time.After(time.Second * 2):
			if p.scrolledUp {
				break
			}
			fmt.Print("timed out, scrolling up\r\n")

			p.scroll(true)
		}
	}
}

func (p *printer) scroll(up bool) {
	if !up {
		p.scrolledUp = false

		// 27 106 nn: Perform nn/216 inch reverse line feed
		p.fd.Write([]byte{27, 106, 216})
		p.fd.Write([]byte{27, 106, 133})
	} else {
		p.scrolledUp = true

		// 27 74 nn: Perform nn/216 inch line feed
		p.fd.Write([]byte{27, 74, 216})
		p.fd.Write([]byte{27, 74, 133})
	}
}

func main() {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Println("stdin is a terminal, setting raw mode")
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			panic(err)
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)
	} else {
		fmt.Println("stdin is not a terminal")
	}

	lp, err := os.Create("/dev/usb/lp0")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := lp.Close(); err != nil {
			panic(err)
		}
	}()

	printer := printer{
		fd:         lp,
		pipe:       make(chan byte),
		scrolledUp: true,
	}

	// ESC/P typewriter mode not emulated :(
	// fo.Write([]byte{27, 105, 1})

	go printer.readChars()
	printer.handleInput()
}
