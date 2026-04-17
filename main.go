package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gm-toolkit <tool>")
		fmt.Println("Tools: faction")
		os.Exit(1)
	}
}
