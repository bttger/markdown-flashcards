package internal

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// ReadFile Reads a markdown file containing flashcards and returns a slice of Card structs.
func (session *Session) ReadFile() error {
	absPath, err := filepath.Abs(session.File.Path)
	check(err)
	f, err := os.Open(absPath)
	check(err)

	scanner := bufio.NewScanner(f)
	var c Card
	category := ""
	readBack := false
	line := 0
	appendCard := func() {
		c.Back = strings.TrimSpace(c.Back)
		if c.Front == "" || c.Back == "" {
			check(errors.New(fmt.Sprint("front or back is empty in line", line-1)))
		}
		session.File.Cards = append(session.File.Cards, c)
		c = Card{}
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
			c.setMetadata(uint(box), strings.Split(args[2], ":")[1], category)
			appendCard()
			readBack = false
		case strings.HasPrefix(t, "# "):
			// category
			if readBack {
				// no metadata found for previous card
				c.initMetadata(category)
				appendCard()
				readBack = false
			}
			category = strings.SplitN(t, " ", 2)[1]
		case strings.HasPrefix(t, "## "):
			// front
			if readBack {
				// no metadata found for previous card
				c.initMetadata(category)
				appendCard()
				readBack = false
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

	err = scanner.Err()
	check(err)

	return f.Close()
}
