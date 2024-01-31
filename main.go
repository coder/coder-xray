package main

import (
	"os"
)

func main() {
	err := root().Execute()
	if err != nil {
		os.Exit(1)
	}
}
