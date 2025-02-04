package main

import "fmt"

var Branch string
var Version string

func ProgramName() string {
	if Branch != "main" {
		return fmt.Sprintf("ascii-draw %s (%s)", Version, Branch)
	} else {
		return fmt.Sprintf("ascii-draw %s", Version)
	}
}
