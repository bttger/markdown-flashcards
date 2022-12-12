package internal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ClearConsole Moves the cursor to the home position (0,0) and erases everything from cursor to end of screen.
func ClearConsole() {
	fmt.Print("\033[H\033[0J")
}

// ScrollDown Scrolls down until the cursor is at the top of the screen.
// https://stackoverflow.com/questions/67212319/ansi-escape-code-csi-6n-always-returns-column-1
// https://pkg.go.dev/github.com/pkg/term/termios#Tcsetattr
// https://en.wikipedia.org/wiki/ANSI_escape_code
// TODO flags are missing
func ScrollDown() {
	var input string
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("\033[6n")
	for scanner.Scan() {
		t := scanner.Text()
		if t == "R" {
			break
		}
		input += t
	}
	row := strings.Split(input, "[")[1]
	row = strings.Split(row, ";")[0]
	col := strings.Split(input, ";")[1]
	col = strings.Split(col, "R")[0]
	fmt.Println(row, col)
	fmt.Print("\033[4S")
}
