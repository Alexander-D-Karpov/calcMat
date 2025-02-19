package main

import (
	"calcMat/ui"
	"log"
)

func main() {
	p := ui.NewProgram()
	if err := p.Start(); err != nil {
		log.Fatal("Error running program:", err)
	}
}
