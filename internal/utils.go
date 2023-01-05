package internal

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/term"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ClearConsole Moves the cursor to the home position (0,0) and erases everything from cursor to end of screen.
func ClearConsole() {
	if os.Getenv("DEBUG") != "true" {
		fmt.Print("\033[H\033[0J")
	}
}

// ScrollDownScreen scrolls down by printing newlines. This can be helpful to prevent overwriting previous console output
// when clearing the console.
func ScrollDownScreen() {
	_, height, err := term.GetSize(int(os.Stdin.Fd()))
	check(err)
	for i := 0; i < height; i++ {
		fmt.Println()
	}
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

// CompareCategory compares the category name to the user input and returns true if the input matches with the
// category according to the following rules:
// - If the input is empty, it will match with any category.
// - The category and input get transformed to lowercase.
// - The input matches the category either if it is equal or if it is a prefix of the category.
func CompareCategory(category, input string) bool {
	if input == "" {
		return true
	}
	category = strings.ToLower(category)
	input = strings.ToLower(input)
	return strings.HasPrefix(category, input)
}

// FindClosestDate finds the closest due date in the future in the given slice of cards.
// If it contains a date that is before or equal to today, it will return an error.
func FindClosestDate(cards []Card) (time.Time, error) {
	y, m, d := time.Now().Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	var closestDate time.Time
	for _, c := range cards {
		if c.Due.Before(today) || c.Due.Equal(today) {
			return time.Time{}, errors.New("found due date in the past")
		}
		if closestDate.IsZero() || c.Due.Before(closestDate) {
			closestDate = c.Due
		}
	}
	return closestDate, nil
}

// WrapLines wraps the given string into lines of the given length.
// It will not break words and thus only breaks at whitespace. It assumes that no word in the given string exceeds the
// requested line length. Lines that start with an indent will be indented by the given indent plus, if the line is
// a list item, the length of the list item prefix.
func WrapLines(s string, lineLength uint) string {
	lineFeedRegex := regexp.MustCompile("\r?\n")
	indentRegex := regexp.MustCompile(`(?m)^[\-+*\d.\s]+`)
	lineBreakRegex := regexp.MustCompile(fmt.Sprintf(`(?m)^.{1,%d}\s`, lineLength))

	var result string
	lines := lineFeedRegex.Split(s, -1)
	for _, line := range lines {
		if len(line) == 0 {
			result += "\n"
		}

		linePrefix := indentRegex.FindString(line)
		lineIndent := len(linePrefix)
		for len(line) > 0 {
			if uint(len(line)) <= lineLength {
				result += line + "\n"
				break
			} else {
				idx := lineBreakRegex.FindStringIndex(line)[1]
				result += line[:idx] + "\n"
				remainder := strings.TrimSpace(line[idx:])
				remainderLen := len([]rune(remainder))
				paddingFmt := fmt.Sprintf("%%%ds", lineIndent+remainderLen)
				line = fmt.Sprintf(paddingFmt, remainder)
			}
		}
	}

	return result
}
