package main

import (
	"15-puzzle/internal/puzzle"
	"fmt"
	"os"
)

func main() {
	if err := puzzle.Init(func(p *puzzle.Controller) { p.SetActive(true) }); err != nil {
		fmt.Fprintf(os.Stderr, "start failed: %v", err)
		os.Exit(1)
	}
}
