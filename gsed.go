package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/pkg/term"
)

func getch() []byte {
	t, _ := term.Open("/dev/tty")
	term.RawMode(t)
	bytes := make([]byte, 3)
	numRead, err := t.Read(bytes)
	t.Restore()
	t.Close()
	if err != nil {
		return nil
	}
	return bytes[0:numRead]
}

type Cursor struct {
	x int
	y int
}

var lines = []string{""}

func init() {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 { // data is being piped from stdin
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}

		lines = strings.Split(string(bytes), "\n")
		for _, line := range lines {
			fmt.Fprintln(os.Stderr, line)
		}
		fmt.Fprintf(os.Stderr, "\033[%vF\033[0G", len(lines))
	}
}

func main() {
	cursor := Cursor{0, 0}
	defer func() {
		fmt.Println(strings.Join(lines, "\n"))
	}()
	for {
		c := getch()
		switch {
		case bytes.Equal(c, []byte{3}):
			return
		case bytes.Equal(c, []byte{4}):
			return
		case bytes.Equal(c, []byte{13}): // newline
			if cursor.y < len(lines)-1 {
				lines = append(lines[:cursor.y+1], append([]string{""}, lines[cursor.y+1:]...)...)
			} else {
				lines = append(lines, "")
			}

			// clear line after cursor
			fmt.Fprint(os.Stderr, strings.Repeat(" ", utf8.RuneCountInString(lines[cursor.y])))

			fmt.Fprint(os.Stderr, "\033[1E")
			if cursor.x < len(lines[cursor.y]) {
				lines[cursor.y+1] = lines[cursor.y][cursor.x:]
				lines[cursor.y] = lines[cursor.y][:cursor.x]
			}
			for _, line := range lines[cursor.y+1:] {
				fmt.Fprintln(os.Stderr, line)
			}
			fmt.Fprint(os.Stderr, " \033[1D")

			fmt.Fprint(os.Stderr, "\033[1G")
			if cursor.y < len(lines)-2 {
				fmt.Fprintf(os.Stderr, "\033[%vA", len(lines)-cursor.y-1)
				fmt.Fprint(os.Stderr, lines[cursor.y+1]+strings.Repeat(" ", utf8.RuneCountInString(lines[cursor.y])+1))
			}

			cursor.x = 0
			cursor.y++

			fmt.Fprint(os.Stderr, "\033[1G")
		case bytes.Equal(c, []byte{127}): // backspace
			if cursor.x > 0 {
				lines[cursor.y] = lines[cursor.y][:cursor.x-1] + lines[cursor.y][cursor.x:]
				fmt.Fprintf(os.Stderr, "\r%s", lines[cursor.y]+" ")
				fmt.Fprintf(os.Stderr, "\033[%vG", cursor.x)
				cursor.x--
			} else if cursor.y > 0 {
				oldLen := len(lines[cursor.y-1])
				lines[cursor.y-1] += lines[cursor.y]
				lines = append(lines[:cursor.y], lines[cursor.y+1:]...) // remove line
				fmt.Fprintf(os.Stderr, "\033[%vF\r", len(lines))
				for _, line := range lines {
					fmt.Fprintln(os.Stderr, line)
				}
				fmt.Fprint(os.Stderr, "\033[36m~\033[39m"+strings.Repeat(" ", len(lines[len(lines)-1])))
				fmt.Fprintf(os.Stderr, "\033[%vF\033[%vG", len(lines)-cursor.y+1, oldLen+1)
				cursor.x = oldLen
				cursor.y--
			}
		case bytes.Equal(c, []byte{27, 91, 68}): // left
			if cursor.x > 0 {
				cursor.x--
				fmt.Fprint(os.Stderr, "\033[1D")
			} else {
				if cursor.y > 0 {
					cursor.y--
					cursor.x = len(lines[cursor.y])
					fmt.Fprintf(os.Stderr, "\033[1A\033[%vG", cursor.x+1)
				}
			}
		case bytes.Equal(c, []byte{27, 91, 67}): // right
			if !(cursor.y >= len(lines)-1 && cursor.x > len(lines[cursor.y])-1) {
				cursor.x++
				if cursor.x > len(lines[cursor.y]) {
					if cursor.y < len(lines)-1 {
						cursor.x = 0
						cursor.y++
						fmt.Fprint(os.Stderr, "\033[1E")
					}
				} else {
					fmt.Fprint(os.Stderr, "\033[1C")
				}
			}
		case bytes.Equal(c, []byte{27, 91, 65}): // up
			if cursor.y > 0 {
				cursor.y--
				fmt.Fprint(os.Stderr, "\033[1A")
			}
			if cursor.x > len(lines[cursor.y]) {
				cursor.x = len(lines[cursor.y])
				fmt.Fprintf(os.Stderr, "\033[%vG", cursor.x+1)
			}
		case bytes.Equal(c, []byte{27, 91, 66}): // down
			if cursor.y < len(lines)-1 {
				cursor.y++
				fmt.Fprint(os.Stderr, "\033[1B")
			}
			if cursor.x >= len(lines[cursor.y]) {
				cursor.x = len(lines[cursor.y])
				fmt.Fprintf(os.Stderr, "\033[%vG", cursor.x+1)
			}
		case bytes.Compare(c, []byte{32}) >= 0 && bytes.Compare(c, []byte{127}) <= 0: // printable chars
			lines[cursor.y] = lines[cursor.y][:cursor.x] + string(c) + lines[cursor.y][cursor.x:]
			fmt.Fprint(os.Stderr, lines[cursor.y][cursor.x:])
			cursor.x++
			fmt.Fprintf(os.Stderr, "\033[%vG", cursor.x+1)
		case bytes.Equal(c, []byte{9}):
			fmt.Fprint(os.Stderr, "\t")
		case bytes.Equal(c, []byte{27}):
			fmt.Fprintf(os.Stderr, "\033[%vF\033[1E", cursor.y+1)
			for _, line := range lines {
				fmt.Fprintln(os.Stderr, line+strings.Repeat(" ", utf8.RuneCountInString(line)))
			}
			fmt.Fprint(os.Stderr, "\033[36m~\033[39m"+strings.Repeat(" ", len(lines[len(lines)-1])))
			fmt.Fprintf(os.Stderr, "\033[%vF", len(lines)-cursor.y)
			fmt.Fprintf(os.Stderr, "\033[%vG", cursor.x+1)
		default:
			// fmt.Fprintln(os.Stderr, "")
			// fmt.Fprintln(os.Stderr, c)
		}
	}
}
