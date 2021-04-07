package main

import (
	"fmt"
	"math/rand"
)

var moods = []string{
	"blissful",
	"determined",
	"devious",
	"excited",
	"ecstatic",
	"gleeful",
	"happy",
	"surprised",
}

var animals = []string{
	"aardvark",
	"bear",
	"capybara",
	"doge",
	"elephant",
	"flamingo",
	"giraffe",
	"hippo",
}

func FunnyAnimal() string {
	return fmt.Sprintf("%s-%s", moods[rand.Int()%len(moods)], animals[rand.Int()%len(animals)])
}
