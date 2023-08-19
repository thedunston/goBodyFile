package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"

	"goBody/processBody"

	// Importing windows package
	"goBody/createBodyWin"
)

/*
Gathers information from the file's inode and write it to a file.
*/

func main() {

	// Be sure -body or -process was passed as os.Args[1].
	if len(os.Args) < 2 {

		fmt.Println("Please provide -body or -process as the first argument.")
		os.Exit(1)
	}

	// Check if the first flag value is "body"
	if os.Args[1] == "-body" {

		// Variable for the irectory that will be searched.
		var rootDir string
		var outputFile string
		var body bool
		var sid bool

		// Pass the name of the directory via the commanldine.
		flag.StringVar(&rootDir, "directory", "", "Directory containing the files to collect metadata.")
		flag.StringVar(&outputFile, "output", "", "Output file name")
		flag.BoolVar(&body, "body", false, "Create Body file")
		flag.BoolVar(&sid, "SID", false, "(Optional). If set this will display the SID for the user and group, the default is to show only the UID and GID.)")

		flag.Parse()

		// Check if the root value is empty.
		if rootDir == "" {

			fmt.Println("Directory not provided. Use -directory and provide the directory name.")
			return

		}

		// Check if the outputFIle value is empty.
		if outputFile == "" {

			fmt.Println("Output file not provided. Use -output and provide the file's name.")
			return

		}

		createBodyWin.CreateBodyfileWin(rootDir, outputFile, sid)

	} else if os.Args[1] == "-process" {

		// Create a flag that checks to see if the user wants to process the body file.
		var flagProcess = flag.Bool("process", false, "Process the body file.")
		var strict = flag.Bool("strict", false, "Only show the entries maching the date restrictions")
		var flagColor = flag.Bool("color", false, "Enable color output")
		var filter = flag.String("filter", "", "Event filter, like \"hour > 14\"")

		flag.Parse()

		flag.Usage = func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
			fmt.Fprintf(flag.CommandLine.Output(), "\t%s [options] MFT.txt\n\n", os.Args[0])
			flag.PrintDefaults()
			os.Exit(1)
		}
		f := processBody.GetInput()

		if !*flagColor || !isatty.IsTerminal(os.Stdout.Fd()) {
			color.NoColor = true // disables colorized output
		}
		// Check that an option was passed.
		flag.Parse()

		processBody.ProcessBody(f, flagProcess, strict, flagColor, filter)

	} else {

		fmt.Println("Please select the '-body' or '-process' option.")

	}

}
