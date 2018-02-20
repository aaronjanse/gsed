package main

import (
	"bytes"
	"fmt"

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

func main() {
	cursor := Cursor{0, 0}
	lines := []string{""}
	for {
		c := getch()
		switch {
		case bytes.Equal(c, []byte{3}):
			return
		case bytes.Equal(c, []byte{13}): // newline
			lines = append(lines, "")
			cursor.x = 0
			cursor.y++
			fmt.Println()
		case bytes.Equal(c, []byte{127}): // backspace
			if cursor.x > 0 {
				lines[cursor.y] = lines[cursor.y][:cursor.x-1] + lines[cursor.y][cursor.x:]
				fmt.Printf("\r%s", lines[cursor.y]+" ")
				fmt.Printf("\033[%vG", cursor.x+1)
			}
			fallthrough
		case bytes.Equal(c, []byte{27, 91, 68}): // left
			if cursor.x > 0 {
				cursor.x--
				fmt.Print("\033[1D")
			} else {
				if cursor.y > 0 {
					cursor.y--
					cursor.x = len(lines[cursor.y])
					fmt.Printf("\033[1A\033[%vG", cursor.x+1)
				}
			}
		case bytes.Equal(c, []byte{27, 91, 67}): // right
			cursor.x++
			if cursor.x > len(lines[cursor.y]) {
				if cursor.y < len(lines)-1 {
					cursor.x = 0
					cursor.y++
					fmt.Print("\033[1E")
				}
			} else {
				fmt.Print("\033[1C")
			}
		case bytes.Equal(c, []byte{27, 91, 67}): // right
			cursor.x++
			if cursor.x > len(lines[cursor.y]) {
				if cursor.y < len(lines)-1 {
					cursor.x = 0
					cursor.y++
					fmt.Print("\033[1E")
				}
			} else {
				fmt.Print("\033[1C")
			}

		case bytes.Equal(c, []byte{27, 91, 65}): // up
			cursor.y--
			if cursor.x > len(lines[cursor.y]) {
				cursor.x = len(lines[cursor.y])
				fmt.Printf("\033[%vG", cursor.x+1)
			}
			fmt.Print("\033[1A")
		case bytes.Equal(c, []byte{27, 91, 66}): // down
			cursor.y++
			if cursor.x > len(lines[cursor.y]) {
				cursor.x = len(lines[cursor.y])
				fmt.Printf("\033[%vG", cursor.x+1)
			}
			fmt.Print("\033[1B")
		case bytes.Compare(c, []byte{32}) >= 0 && bytes.Compare(c, []byte{127}) <= 0: // printable chars
			cursor.x++
			lines[cursor.y] += string(c)
			fmt.Print(string(c))
		case bytes.Equal(c, []byte{9}):
			fmt.Print("\t")
		default:
			// fmt.Println()
			// fmt.Println(c)
		}
	}
}
