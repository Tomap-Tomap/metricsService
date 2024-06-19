// Package build contains methods for displaying information about build
package build

import (
	"cmp"
	"fmt"
)

// DisplayBuild displays information about the version date and commit of the application in the console
func DisplayBuild(version, date, commit string) (string, string, string) {
	emptyData := "N/A"

	version = cmp.Or(version, emptyData)
	date = cmp.Or(date, emptyData)
	commit = cmp.Or(commit, emptyData)

	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)

	return version, date, commit
}
