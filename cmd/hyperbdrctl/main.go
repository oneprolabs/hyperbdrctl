package main

import (
	"fmt"
	"os"

	"hyperbdr-client/internal/commands"
)

func main() {
	if err := commands.Execute(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
