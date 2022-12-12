package internal

import "time"

type Card struct {
	Front    string
	Back     string
	Category string
	// 1: next day
	// 2: e.g. 3 days later
	// 3: e.g. 7 days later
	// 4: e.g. 15 days later
	// 5: e.g. 30 days later
	Box uint
	Due string
}

type File struct {
	Path         string
	BoxIntervals []int
	Cards        []Card
}

type Session struct {
	Sequential, TestMode, ShowCategory bool
	Category                           string
	ListCategories                     bool
	NumberCards                        uint
	File                               File
}

func (c *Card) initMetadata(category string) {
	c.Box = 1
	c.Due = time.Now().Format("2006-01-02")
	c.Category = category
}

func (c *Card) setMetadata(box uint, due, category string) {
	c.Box = box
	c.Due = due
	c.Category = category
}
