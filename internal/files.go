package internal

import (
	"bufio"
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func NewFile(path string) File {
	if path == "" {
		check(errors.New("no file specified"))
	}
	absPath, err := filepath.Abs(path)
	check(err)
	return File{Path: absPath, BoxIntervals: boxIntervals}
}

// ReadFile Reads a markdown file containing flashcards and returns a slice of Card structs.
func (s *Session) ReadFile() error {
	if s.File.Path == "" {
		return errors.New("no file specified")
	}
	f, err := os.Open(s.File.Path)
	check(err)

	scanner := bufio.NewScanner(f)
	var c Card
	category := ""
	readBack := false
	line := 0
	appendCard := func() {
		c.Back = strings.TrimSpace(c.Back)
		if c.Front == "" || c.Back == "" {
			check(errors.New(fmt.Sprint("front or back is empty in line ", line-1)))
		}
		s.File.Cards = append(s.File.Cards, c)
		c = Card{}
	}
	appendNewCard := func() {
		c.initMetadata(category)
		appendCard()
		readBack = false
	}
	for scanner.Scan() {
		line++
		t := scanner.Text()
		switch {
		case strings.HasPrefix(t, "`mdfc;"):
			// metadata
			args := strings.Split(t, ";")
			box, err := strconv.ParseUint(strings.Split(args[1], ":")[1], 10, 64)
			check(err)
			due, err := time.Parse("2006-01-02", strings.Split(args[2], ":")[1])
			check(err)
			c.setMetadata(uint(box), due, category)
			appendCard()
			readBack = false
		case strings.HasPrefix(t, "# "):
			// category
			if readBack {
				// no metadata found for previous card
				appendNewCard()
			}
			category = strings.SplitN(t, " ", 2)[1]
		case strings.HasPrefix(t, "## "):
			// front
			if readBack {
				// no metadata found for previous card
				appendNewCard()
			}
			c.Front = strings.SplitN(t, " ", 2)[1]
			readBack = true
		default:
			// back
			if readBack {
				c.Back += t + "\n"
			}
		}
	}
	if readBack {
		// no metadata found for previous card and EOF reached
		appendNewCard()
	}

	err = scanner.Err()
	check(err)

	if len(s.File.Cards) == 0 {
		return errors.New("no flashcards found in file")
	}

	return f.Close()
}

// WriteFile Writes the file to disk with updated metadata.
func (s *Session) WriteFile() error {
	if s.File.Path == "" {
		return errors.New("no file specified")
	}
	f, err := os.Create(s.File.Path)
	check(err)

	w := bufio.NewWriter(f)
	lastCategory := ""
	for _, c := range s.File.Cards {
		if c.Category != lastCategory {
			_, err := fmt.Fprintf(w, "# %s\n\n", c.Category)
			check(err)
			lastCategory = c.Category
		}
		_, err := fmt.Fprintf(w, "## %s\n\n", c.Front)
		check(err)
		_, err = fmt.Fprintf(w, "%s\n\n", c.Back)
		check(err)
		_, err = fmt.Fprintf(w, "`mdfc;box:%d;due:%s;`\n\n", c.Box, c.Due.Format("2006-01-02"))
		check(err)
	}

	err = w.Flush()
	check(err)
	return f.Close()
}

// CheckCategory Checks if the session's category is valid, meaning it is present in the File. If the input is empty, it
// returns nil according to the CompareCategory function.
func (s *Session) CheckCategory() error {
	for _, c := range s.File.Cards {
		if CompareCategory(c.Category, s.Category) {
			return nil
		}
	}
	return errors.New("category not found")
}

// ChooseCategory Lets the user choose a category from the file's headings.
func (s *Session) ChooseCategory() {
	fmt.Println("Please select the category you want to study:")
	var categories []string
	for _, c := range s.File.Cards {
		if !slices.Contains(categories, c.Category) {
			categories = append(categories, c.Category)
		}
	}
	for i, c := range categories {
		fmt.Printf("(%d) %s\n", i+1, c)
	}

	fmt.Print("Your choice: ")
	choice := ReadNumberInput(1, len(categories))
	s.Category = categories[choice-1]
}
