package main

import (
	"log"
	"os"

	"social-network/backend/entry"
)

func main() {
	reseed := false
	for _, arg := range os.Args[1:] {
		if arg == "--reseed" {
			reseed = true
		}
	}

	if err := entry.Start(reseed); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
