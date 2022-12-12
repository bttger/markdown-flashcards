package main

import (
	"encoding/json"
	"fmt"
	"github.com/bttger/markdown-flashcards/internal"
	"os"
	"strconv"
)

const defaultNumberCards = 20

var boxIntervals = []int{1, 3, 7, 15, 30}

func main() {
	args := os.Args[1:]
	session := internal.Session{}
	session.NumberCards = defaultNumberCards

	readOptArg := false
	for i, arg := range args {
		switch arg {
		case "-h", "--help":
			fmt.Println("Usage: mdfc [options] [file]")
			fmt.Println("\nOptions:")
			fmt.Println("\n\t-h, --help")
			fmt.Println("\t\tShow this help message and exit.")
			fmt.Println("\n\t-s, --sequential")
			fmt.Println("\t\tShow flashcards in sequential order as in the markdown file. The default behavior is to")
			fmt.Println("\t\tshow flashcards in random order.")
			fmt.Println("\n\t-c, --category <category>")
			fmt.Println("\t\tShow only flashcards of the specified category. If no category is specified, you can")
			fmt.Println("\t\tinteractively choose one.")
			fmt.Println("\t\tIt is possible to specify the category by a chapter number or the categories first word,")
			fmt.Println("\t\te.g. \"2.1\" for \"2.1 Regular Expressions\".")
			fmt.Println("\n\t-o, --show-category")
			fmt.Println("\t\tShow the category of each flashcard.")
			fmt.Println("\n\t-t, --test <number_flashcards>")
			fmt.Println("\t\tTest yourself in test mode with random flashcards. If no number is specified, all")
			fmt.Println("\t\tflashcards will be tested.")
			fmt.Println("\n\t-n, --number <number_flashcards>")
			fmt.Println("\t\tLearn n cards during the session. If no number is specified, it will fall back to the")
			fmt.Println("\t\tnumber specified in the YAML front matter of the markdown file. Defaults to 20.")
			return
		case "-s", "--sequential":
			session.Sequential = true
		case "-o", "--show-category":
			session.ShowCategory = true
		case "-c", "--category":
			session.ListCategories = true
			readOptArg = true
		case "-t", "--test":
			session.TestMode = true
			readOptArg = true
		case "-n", "--number":
			readOptArg = true
		default:
			if readOptArg && i != len(args)-1 {
				switch args[i-1] {
				case "-c", "--category":
					session.ListCategories = false
					session.Category = arg
				case "-n", "--number", "-t", "--test":
					n, err := strconv.Atoi(arg)
					if err != nil {
						fmt.Println("Invalid number of flashcards specified.")
						return
					}
					session.NumberCards = uint(n)
				}
				readOptArg = false
			} else {
				session.File = internal.File{Path: arg, BoxIntervals: boxIntervals}
			}
		}
	}

	if len(session.File.Path) == 0 {
		fmt.Println("No file specified.")
		return
	}

	err := session.ReadFile()
	if err != nil {
		panic(err)
	}
	out, err := json.MarshalIndent(session, "", "  ")
	fmt.Println(string(out))
	// TODO check if we need to interactively choose a category
}
