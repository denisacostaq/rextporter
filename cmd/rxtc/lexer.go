package main

import (
	"fmt"

	"github.com/simelo/rextporter/src/rxt/grammar"
)

func main() {
	// grammar.LexTheRxt(&rxt.TokenWriter{}, "LEX")
	fmt.Println(grammar.Parse())
}
