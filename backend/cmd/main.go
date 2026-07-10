package main

import (
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

	entry.Start(reseed)
}
