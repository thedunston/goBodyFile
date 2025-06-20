package common

import (
	"fmt"
	"os"
	"strings"
)

func CheckFileExists(outputFile string) {

	// Check if the file already exists.
	if _, err := os.Stat(outputFile); err == nil {

		// File exists, prompt to overwrite
		fmt.Printf("File '%s' already exists. Do you want to overwrite it? (y/n): ", outputFile)

		var overwrite string

		fmt.Scanln(&overwrite)

		// Remove spaces from output.
		overwrite = strings.TrimSpace(overwrite)

		// If y or Y isn't selected, exit the program.
		if !(overwrite == "y" || overwrite == "Y") {

			fmt.Println("Exiting the program.")
			os.Exit(1)

		}

	}

}

func CheckDirectoryExists(rootDir string) {
	// Check if the directory exists.
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {

		fmt.Printf("The directory doesn't exist! => %s\n", rootDir)
		return

	}

}
