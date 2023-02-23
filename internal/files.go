package internal

import (
	"bufio"
	"errors"
	"fmt"
	gonanoid "github.com/matoous/go-nanoid"
	"golang.org/x/exp/slices"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// getMetadata extracts the metadata (ID, box, due date; embedded in html comment tag) from a line.
func getMetadata(line string) (id, box, due string) {
	re := regexp.MustCompile(`<!--\s*(.{4});(\d);(\d{4}-\d{2}-\d{2})\s*-->`)
	matches := re.FindStringSubmatch(line)
	if len(matches) == 4 {
		return matches[1], matches[2], matches[3]
	}
	return
}

// initializeMetadata initializes the metadata (ID, box, due date; embedded in html comment tag) for a new card.
func initializeMetadata(line string) (updatedLine, id, box, due string) {
	id = gonanoid.MustGenerate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", 4)
	box = "0"
	due = time.Now().Format("2006-01-02")
	updatedLine = fmt.Sprintf("%s <!--%s;%s;%s-->", line, id, box, due)
	return
}

// generateNewId generates a new id for a card and updates the line with the new id.
func generateNewId(line string) (updatedLine, id string) {
	re := regexp.MustCompile(`<!--\s*(.{4});(\d);(\d{4}-\d{2}-\d{2})\s*-->`)
	id = gonanoid.MustGenerate("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", 4)
	updatedLine = re.ReplaceAllString(line, fmt.Sprintf("<!--%s;$2;$3-->", id))
	return
}

// extractQuestion extracts the question from a second-level (or third, etc.) markdown header.
func extractQuestion(line string) string {
	re := regexp.MustCompile(`##\s+(.*)<!--`)
	matches := re.FindStringSubmatch(line)
	if len(matches) == 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// getCardFromLine extracts the card data from a second-level (or third, etc.) markdown header.
func getCardFromLine(line, category string) (card Card) {
	card.Category = category
	id, box, due := getMetadata(line)
	card.Id = id
	boxUint, err := strconv.Atoi(box)
	check(err)
	card.Box = uint(boxUint)
	card.Due, err = time.Parse("2006-01-02", due)
	check(err)
	card.Front = extractQuestion(line)
	return
}

// OpenFile Reads a markdown file containing flashcards and initializes the Session.
func (s *Session) OpenFile(path string) error {
	if path == "" {
		return errors.New("no file specified")
	}
	absPath, err := filepath.Abs(path)
	check(err)
	s.File = File{Path: absPath, BoxIntervals: boxIntervals}

	if s.File.Path == "" {
		return errors.New("no file specified")
	}
	f, err := os.Open(s.File.Path)
	if err != nil {
		return errors.New("file not found")
	}

	ids := make(map[string]bool)
	lines := make([]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "### ") || strings.HasPrefix(line, "#### ") {
			id, _, _ := getMetadata(line)
			if id == "" {
				line, id, _, _ = initializeMetadata(line)
			}
			for ids[id] {
				line, id = generateNewId(line)
			}
			ids[id] = true
		}
		lines = append(lines, line)
	}
	err = f.Close()
	check(err)

	// Update the file with the new metadata
	f, err = os.Create(s.File.Path)
	_, err = f.WriteString(strings.Join(lines, "\n"))
	check(err)
	err = f.Sync()
	check(err)
	err = f.Close()
	check(err)

	// Go through all lines and initialize questions with ID if they don't have one.
	// Also, initialize the File's Cards.
	currentCard := Card{}
	currentCategory := ""
	readBack := false
	appendCard := func() {
		currentCard.Back = strings.TrimSpace(currentCard.Back)
		s.File.Cards = append(s.File.Cards, currentCard)
		currentCard = Card{}
	}

	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, "# "):
			if currentCard.Front != "" && currentCard.Back != "" {
				appendCard()
			}
			currentCategory = strings.TrimSpace(l[2:])
			readBack = false
		case strings.HasPrefix(l, "## "), strings.HasPrefix(l, "### "), strings.HasPrefix(l, "#### "):
			if currentCard.Front != "" && currentCard.Back != "" {
				appendCard()
			}
			currentCard = getCardFromLine(l, currentCategory)
			readBack = true
		default:
			if readBack {
				currentCard.Back += l + "\n"
			}
		}
	}
	// End of file reached, append the last card
	if readBack {
		appendCard()
	}

	if len(s.File.Cards) == 0 {
		return errors.New("no flashcards found in file")
	}

	return nil
}

// updateCardInFile Updates the card's metadata in the file.
func (s *Session) updateCardInFile(c *Card) {
	data, err := os.ReadFile(s.File.Path)
	check(err)
	md := string(data)
	re := regexp.MustCompile(fmt.Sprintf(`<!--\s*%s;\d;\d{4}-\d{2}-\d{2}\s*-->`, c.Id))
	md = re.ReplaceAllString(md, fmt.Sprintf("<!--%s;%d;%s-->", c.Id, c.Box, c.Due.Format("2006-01-02")))
	err = os.WriteFile(s.File.Path, []byte(md), 0644)
	check(err)
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
