package internal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ClearConsole Moves the cursor to the home position (0,0) and erases everything from cursor to end of screen.
func ClearConsole() {
	if os.Getenv("DEBUG") != "true" {
		fmt.Print("\033[H\033[0J")
	}
}

// ScrollDownFake scrolls down by printing newlines
func ScrollDownFake() {
	for i := 0; i < 60; i++ {
		fmt.Println()
	}
}

// ScrollDown Scrolls down until the cursor is at the top of the screen.
// https://stackoverflow.com/questions/67212319/ansi-escape-code-csi-6n-always-returns-column-1
// https://en.wikipedia.org/wiki/ANSI_escape_code
// https://pkg.go.dev/github.com/pkg/term/termios#Tcsetattr
// TODO flags are missing
func scrollDown() {
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

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// PrintJSON pretty prints any struct as JSON
func PrintJSON[T any](v T) {
	out, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(out))
}

// ReadNumberInput reads a number from standard input. The number must be within i and j. If it is not, it will retry.
func ReadNumberInput(i, j int) int {
	res := i - 1
	scanner := bufio.NewScanner(os.Stdin)
	for res < i || res > j {
		scanner.Scan()
		in := scanner.Text()
		nr, err := strconv.Atoi(in)
		if err != nil || nr < i || nr > j {
			fmt.Print("Please enter a number: ")
			continue
		}
		res = nr
	}
	return res
}

// ReadEnterInput Blocks until the user enters a newline.
func ReadEnterInput() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
}
