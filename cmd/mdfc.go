package main

import (
	"fmt"
	"github.com/bttger/markdown-flashcards/internal"
	"os"
	"strconv"
)

func printHelp() {
	fmt.Println("Usage: mdfc [options] [file]")
	fmt.Println("\nOptions:")
	fmt.Println("\n\t-h, --help")
	fmt.Println("\t\tShow this help message and exit.")
	fmt.Println("\n\t-s, --sequential")
	fmt.Println("\t\tShow flashcards in sequential order as in the markdown file. The default behavior is to")
	fmt.Println("\t\tshow flashcards in random order.")
	fmt.Println("\n\t-o, --show-category")
	fmt.Println("\t\tShow the category of each flashcard.")
	fmt.Println("\n\t-c, --category <category>")
	fmt.Println("\t\tShow only flashcards of the specified category. A category is a first-level heading in the")
	fmt.Println("\t\tmarkdown file. A category can be specified by a case-insensitive prefix of the heading.")
	fmt.Println("\t\tIf no category is specified, you can interactively choose one.")
	fmt.Println("\n\t-t, --test <number_flashcards>")
	fmt.Println("\t\tTest yourself in test mode with random flashcards. If no number is specified, all")
	fmt.Println("\t\tflashcards will be shown. Possible to combine with -c, --category.")
	fmt.Println("\n\t-n, --number <number_flashcards>")
	fmt.Println("\t\tLearn n cards during the session. Set it to 0 to study all cards that are due to today.")
	fmt.Println("\t\tDefaults to 20.")
	fmt.Println("\n\t-f, --future-days-due <days>")
	fmt.Println("\t\tUsually a flashcard is due on a particular date. If you want to learn flashcards")
	fmt.Println("\t\tbefore they are due, you can specify the number of days in the future when a flashcard")
	fmt.Println("\t\tshould be due. This might be helpful in the case when you have no cards due for today's")
	fmt.Println("\t\tlearning session. Cards where the due date was missed will be added anyway. Defaults to 0.")
	fmt.Println("\n\t-w, --wrap-lines <line_length>")
	fmt.Println("\t\tWrap lines to a maximum length. Only breaks lines at whitespaces. Defaults to terminal width.")
}

func printDebugHelp(session internal.Session) {
	if os.Getenv("DEBUG") == "true" {
		internal.PrintJSON(session)
	}
}

const defaultNumberCards = 20

func main() {
	args := os.Args[1:]
	session := internal.Session{NumberCards: defaultNumberCards}

	readOptArg := false
	for i, arg := range args {
		switch arg {
		case "-h", "--help":
			printHelp()
			return
		case "-s", "--sequential":
			session.Sequential = true
		case "-o", "--show-category":
			session.ShowCategory = true
		case "-c", "--category":
			session.ChooseCategories = true
			readOptArg = true
		case "-t", "--test":
			session.TestMode = true
			session.NumberCards = 0
			readOptArg = true
		case "-n", "--number":
			session.NumberCards = defaultNumberCards
			readOptArg = true
		case "-f", "--future-days-due":
			session.FutureDaysDue = 0
			readOptArg = true
		case "-w", "--wrap-lines":
			session.WrapLines = 0
			readOptArg = true
		default:
			if readOptArg && i != len(args)-1 {
				switch args[i-1] {
				case "-c", "--category":
					session.ChooseCategories = false
					session.Category = arg
				case "-n", "--number", "-t", "--test":
					n, err := strconv.Atoi(arg)
					if err != nil || n < 0 {
						fmt.Println("Invalid number of flashcards specified.")
						return
					}
					session.NumberCards = uint(n)
				case "-f", "--future-days-due":
					n, err := strconv.Atoi(arg)
					if err != nil || n < 0 {
						fmt.Println("Invalid number of future due days specified.")
						return
					}
					session.FutureDaysDue = uint(n)
				case "-w", "--wrap-lines":
					n, err := strconv.Atoi(arg)
					if err != nil || n < 0 {
						fmt.Println("Invalid number of maximum line length specified.")
						return
					}
					session.WrapLines = uint(n)
				}
				readOptArg = false
			} else {
				session.File = internal.NewFile(arg)
			}
		}
	}

	err := session.OpenFile()
	if err != nil {
		fmt.Printf("%v\n\n", err)
		printHelp()
		return
	}

	err = session.CheckCategory()
	if err != nil {
		fmt.Println("Invalid category specified.")
		return
	}

	printDebugHelp(session)
	if session.ChooseCategories && session.Category == "" {
		session.ChooseCategory()
	}
	session.Start()
	printDebugHelp(session)
}
