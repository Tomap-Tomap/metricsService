package main

import (
	"fmt"
	"os"
)

func main() {
	os.Exit(0)        // want "use os.Exit"
	fmt.Print("test") // want
}

func osExitCheckFunc() {
	os.Exit(0)        // want
	fmt.Print("test") // want
}
