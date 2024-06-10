package main

import (
	"flag"
	"fmt"
	"os"
	"simpyl/repl"
)

func main() {
	file := flag.String("file", "", "")
	flag.Parse()

	if *file != "" {
		repl.StartInterpreter(*file)
	} else {
		fmt.Print("Welcome to the Simpyl programming language!\n")
		repl.StartInteractive(os.Stdin, os.Stdout)
	}
}
