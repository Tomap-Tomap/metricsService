package pkg1

import (
	"fmt"
	"os"
)

func osExitCheckFunc() {
	os.Exit(0) // want "use os.Exit"
	fmt.Print("test")
}
