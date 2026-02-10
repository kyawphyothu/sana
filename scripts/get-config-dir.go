package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(filepath.Join(configDir, "sana"))
}
