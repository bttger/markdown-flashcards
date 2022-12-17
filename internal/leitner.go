package internal

import (
	"fmt"
	"math/rand"
	"time"
)

// Difficulty constants
const (
	NotRemembered = 0
	Hard          = 0.8
	Okay          = 1
	Easy          = 1.5
)

// boxIntervals are the days between the last review and the next review, and they depend on the box the card is in.
var boxIntervals = []uint{0, 1, 2, 4, 8, 15, 25}

type Card struct {
	Front    string
	Back     string
	Category string
	// Box number starts at 0
	Box uint
	Due time.Time
}

type File struct {
	Path         string
	BoxIntervals []uint
	Cards        []Card
}

type Session struct {
	Sequential, TestMode, ShowCategory bool
	Category                           string
	ChooseCategories                   bool
	// Number of cards to study. If 0, study all cards.
	NumberCards uint
	// Usually a flashcard is due on a particular date. But if the study set would be less than Session.NumberCards,
	// the due date is ignored up to a certain number of days in the future. The cards where the due date was missed
	// are added to the study set anyway.
	FutureDaysDue uint
	File          File
	studyQueue    []*Card
	currentCard   *Card
}

type TestModeResults struct {
	NotRemembered, Hard, Okay, Easy uint
}

// initMetadata Initialize the metadata of a new card.
func (c *Card) initMetadata(category string) {
	c.Category = category
	c.resetMetadata()
}

// resetMetadata Resets the metadata of a card. This means setting the box to 0 and the due date to today.
func (c *Card) resetMetadata() {
	c.Box = 0
	y, m, d := time.Now().Date()
	c.Due = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// setMetadata Set the metadata of a card from the file.
func (c *Card) setMetadata(box uint, due time.Time, category string) {
	c.Box = box
	c.Due = due
	c.Category = category
}

// Start Starts the study session.
func (s *Session) Start() {
	testModeResults := TestModeResults{}

	s.assembleStudyQueue()
	if len(s.studyQueue) == 0 {
		fmt.Print("\nLooks like you don't have anything to study today.\n\n")
		fmt.Println("If you want to learn cards that are scheduled for the next")
		fmt.Print("few days, use the --future-days-due flag.\n\n")
		s.PrintNextDueDate()
		return
	}

	// Start the study session.
	for len(s.studyQueue) > 0 {
		ScrollDownFake()
		card, difficulty := s.flashNextCard()

		// If in test mode, don't update the metadata.
		if s.TestMode {
			switch difficulty {
			case NotRemembered:
				testModeResults.NotRemembered++
			case Hard:
				testModeResults.Hard++
			case Okay:
				testModeResults.Okay++
			case Easy:
				testModeResults.Easy++
			}
		} else {
			s.updateCard(card, difficulty)
			err := s.WriteFile()
			check(err)
		}
	}

	// Output a user hint about (next) session.
	ClearConsole()
	if s.TestMode {
		fmt.Printf("Not remembered:\t%d\n", testModeResults.NotRemembered)
		fmt.Printf("Hard:\t\t%d\n", testModeResults.Hard)
		fmt.Printf("Okay:\t\t%d\n", testModeResults.Okay)
		fmt.Printf("Easy:\t\t%d\n", testModeResults.Easy)
	}
	fmt.Println("You're done with your session!")
	s.PrintNextDueDate()
}

func (s *Session) PrintNextDueDate() {
	nextSession, err := FindClosestDate(s.File.Cards)
	if err != nil {
		fmt.Println("Please note: You still have cards due to today.")
	} else {
		fmt.Println("Next due date:", nextSession.Format("2006-01-02"))
	}
}

// isDue Checks if a card is due. Returns two values: the first is true if the card is due, the second is true if the
// card is due within the next maxFutureDaysDue days.
func (s *Session) isDue(c Card) (due, nearDue bool) {
	y, m, d := time.Now().Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	if today.After(c.Due) || today.Equal(c.Due) {
		due = true
	} else if nearDay := c.Due.AddDate(0, 0, -int(s.FutureDaysDue)); today.After(nearDay) || today.Equal(nearDay) {
		nearDue = true
	}
	return due, nearDue
}

// assembleStudyQueue Assembles the cards that need to be studied according to their due date, the number of cards
// the user wants to study, and the category. Shuffles the cards if the user doesn't want to study them sequentially.
func (s *Session) assembleStudyQueue() {
	nearDueQueue := make([]*Card, 0)

	for i := 0; i < len(s.File.Cards); i++ {
		c := &s.File.Cards[i]
		if s.NumberCards == 0 && s.Category == "" {
			// Study all cards.
			s.studyQueue = append(s.studyQueue, c)
			continue
		}
		if CompareCategory(c.Category, s.Category) {
			// If a category is specified, only study cards of that category.
			due, nearDue := s.isDue(*c)
			if due || s.TestMode {
				s.studyQueue = append(s.studyQueue, c)
			}
			if nearDue {
				nearDueQueue = append(nearDueQueue, c)
			}
		}
		if s.NumberCards > 0 && uint(len(s.studyQueue)) == s.NumberCards {
			// Break the loop when the number of cards to study is reached.
			break
		}
	}

	// If the study set would be less than s.NumberCards, add cards that are due in the near future.
	if s.NumberCards > 0 && uint(len(s.studyQueue)) < s.NumberCards {
		for _, c := range nearDueQueue {
			s.studyQueue = append(s.studyQueue, c)
			if uint(len(s.studyQueue)) == s.NumberCards {
				break
			}
		}
	}

	// Update the number of cards to be able to track the progress.
	s.NumberCards = uint(len(s.studyQueue))

	if !s.Sequential {
		// The shuffling must happen after the queue has been assembled because of the edge case when the user
		// adds more cards to his markdown file than what he wants to study per session. If the shuffling would happen
		// before the queue is assembled, the user would not be able to study the same set of cards and there would be
		// no learning effect of the spaced repetition.
		rand.Shuffle(len(s.studyQueue), func(i, j int) {
			s.studyQueue[i], s.studyQueue[j] = s.studyQueue[j], s.studyQueue[i]
		})
	}
}

// flashNextCard Shows a card's front side. The card is picked from the study queue.
// Waits for the user to press a key to signal how difficult the card was to remember.
func (s *Session) flashNextCard() (c *Card, difficulty float32) {
	ClearConsole()
	fmt.Printf("--- Cards left for today: %d / %d", len(s.studyQueue), s.NumberCards)

	// Dequeue the next card.
	c = s.studyQueue[0]
	s.studyQueue = s.studyQueue[1:]

	if s.ShowCategory {
		fmt.Printf("   (%s)", c.Category)
	}
	fmt.Printf(" ---")

	fmt.Printf("\n\n%s\n\n", c.Front)

	fmt.Print("--> Press enter to show the back side.")
	ReadEnterInput()
	fmt.Printf("\n%s\n\n", c.Back)

	fmt.Println("--> How difficult was it to remember?")
	fmt.Printf("--> (1) Not remembered, (2) Hard, (3) Okay, (4) Easy: ")
	d := ReadNumberInput(1, 4)
	switch d {
	case 1:
		difficulty = NotRemembered
	case 2:
		difficulty = Hard
	case 3:
		difficulty = Okay
	case 4:
		difficulty = Easy
	}
	return c, difficulty
}

// updateCard Updates the card's metadata according to the user's input. This method may change the box and due date.
// It may also add it back to the study queue if the answer was not remembered.
func (s *Session) updateCard(c *Card, difficulty float32) {
	if difficulty == NotRemembered {
		// Move the card to the first box and add it to the study queue.
		c.resetMetadata()
		s.studyQueue = append(s.studyQueue, c)
	} else {
		// Move the card to the next box but only if it is not in the last box and the answer was not Hard.
		if difficulty != Hard && c.Box < uint(len(s.File.BoxIntervals))-1 {
			c.Box++
		}
		y, m, d := time.Now().Date()
		daysInFuture := int(float32(s.File.BoxIntervals[c.Box]) * difficulty)
		if daysInFuture == 0 {
			// Since the int conversion floors the number, make sure the card is due at least one day in the future.
			daysInFuture = 1
		}
		c.Due = time.Date(y, m, d, 0, 0, 0, 0, time.UTC).AddDate(0, 0, daysInFuture)
	}
}
