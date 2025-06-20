package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"gobodyfile/createBody"
	"gobodyfile/processBody"
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

	// Check if the first flag value is "body."
	if os.Args[1] == "-body" {

		// Variable for the directory that will be searched.
		var rootDir string
		var outputFile string
		var body bool
		var sid bool

		// Pass the name of the directory via the commandline.
		flag.StringVar(&rootDir, "directory", "", "Directory containing the files to collect metadata.")
		flag.StringVar(&outputFile, "output", "", "Output file name")
		flag.BoolVar(&body, "body", false, "Create Body file")
		flag.BoolVar(&sid, "sid", false, "(Optional) Display the SID. Default will return the UID and GID.")

		flag.Parse()

		// Check if the root value is empty.
		if rootDir == "" {

			fmt.Println("Directory not provided. Use -directory and provide the directory name.")
			return

		}

		// Check if the outputFile value is empty.
		if outputFile == "" {

			fmt.Println("Output file not provided. Use -output and provide the file's name.")
			return

		}

		// Create the body file.
		createBody.CreateBody(rootDir, outputFile, sid)

	} else if os.Args[1] == "-process" {

		// Create flags for processing the body file.
		var flagProcess = flag.Bool("process", false, "Process the body file.")
		var strict = flag.Bool("strict", false, "Only show the entries matching the date restrictions")
		var filter = flag.String("filter", "", "Event filter (e.g., \"hour > 12\", \"day == 19\", \"weekday == \\\"Monday\\\"\")")
		var modifiedFilter = flag.String("modified", "", "Filter on modification time only (e.g., \"date > \\\"2025-06-17\\\"\")")
		var accessFilter = flag.String("access", "", "Filter on access time only (e.g., \"date > \\\"2025-06-17\\\"\")")
		var ctimeFilter = flag.String("ctime", "", "Filter on change time only (e.g., \"date > \\\"2025-06-17\\\"\")")

		flag.Usage = func() {
			usage := fmt.Sprintf(`Usage of %s:
	%s [options] bodyfile.txt

Filter examples:
  -filter "hour > 12"     (files modified after noon)
  -filter "hour < 6"      (files modified before 6 AM)
  -filter "day == 19"     (files modified on the 19th)
  -filter "weekday == \"Monday\"" (files modified on Monday)
  -filter "hour >= 9 && hour <= 17" (files modified 9 AM to 5 PM)
  -filter "date > "%s"" (files modified in last hour)
  -filter "date > "2025-06-19 13:47:35"" (files modified after specific time)
  -filter "date > "2025-06-19"" (files modified after specific date)
  -filter "date > "2025/06/19 13:47:35"" (slash format also supported)

Timestamp-specific filters (only show entries for the specified timestamp type):
  -modified "date < "2025-06-17"" (filter on modification time only)
  -access "date > "2025-06-19"" (filter on access time only)
  -ctime "date == "2025-06-16"" (filter on change time only)

Note: -filter checks ALL timestamp types (access, modification, change, creation).
      -modified, -access, -ctime check ONLY the specified timestamp type.
      Use -strict to show only matching timestamps instead of all timestamps for matching files.

IMPORTANT: Both YYYY-MM-DD and YYYY/MM/DD formats are supported for date filters.
For relative time, use 'date > "YYYY-MM-DD HH:MM:SS"' instead

`, os.Args[0], os.Args[0], time.Now().Add(-1*time.Hour).Format("2006-01-02 15:04:05"))

			fmt.Fprintf(flag.CommandLine.Output(), usage)
			flag.PrintDefaults()
			os.Exit(1)
		}

		flag.Parse()

		// Get the input file from the command line.
		f := processBody.GetInput()

		// Process the body file.
		processBody.ProcessBody(f, flagProcess, strict, filter, modifiedFilter, accessFilter, ctimeFilter)

	} else {

		fmt.Println("Please select the '-body' or '-process' option.")

	}

}
