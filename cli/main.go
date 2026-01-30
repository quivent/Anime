package main

import (
	"os"

	"github.com/joshkornreich/anime/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		// Error has already been beautifully displayed by our custom handler
		// Just exit with error code
		os.Exit(1)
	}
}
