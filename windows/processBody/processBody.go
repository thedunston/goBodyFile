package processBody

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/airbus-cert/bodyfile"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

var colorDisabled = color.New(color.FgWhite).SprintFunc()

func GetInput() *os.File {
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		return os.Stdin
	}

	if flag.NArg() == 0 {
		flag.Usage()
	}

	filename := flag.Arg(0)

	if filename == "-" {
		return os.Stdin
	}

	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open %s: %s", filename, err)
		os.Exit(1)
	}

	return f
}

func ProcessBody(f *os.File, flagProcess *bool, strict *bool, flagColor *bool, filter *string) {

	body := bodyfile.NewReader(f)
	if *filter != "" {
		err := body.AddFilter(*filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not add filter: %s", err)
			os.Exit(2)
		}
	}

	body.Strict = *strict

	if _, err := body.Slurp(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not read all the content: %s", err)
		os.Exit(3)
	}

	prev := time.Now()
	for {
		tsEntry, err := body.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while iterating: %s", err)
			os.Exit(4)
		}

		mChar := entryType(tsEntry, tsEntry.Entry.ModificationTime, "m")
		aChar := entryType(tsEntry, tsEntry.Entry.AccessTime, "a")
		cChar := entryType(tsEntry, tsEntry.Entry.ChangeTime, "c")
		bChar := entryType(tsEntry, tsEntry.Entry.CreationTime, "b")

		macbLine := fmt.Sprintf("%s%s%s%s", mChar, aChar, cChar, bChar)

		date := tsEntry.Time.Format("2006-01-02")
		if date == prev.Format("2006-01-02") {
			date = colorDisabled(date)
		}

		hour := fmt.Sprintf("%02d", tsEntry.Time.Hour())
		min := fmt.Sprintf(":%02d", tsEntry.Time.Minute())
		sec := fmt.Sprintf(":%02d:", tsEntry.Time.Second())
		if tsEntry.Time.Hour() == prev.Hour() {
			hour = colorDisabled(hour)

			if tsEntry.Time.Minute() == prev.Minute() {
				min = colorDisabled(min)

				if tsEntry.Time.Second() == prev.Second() {
					sec = colorDisabled(sec)
				}
			}
		}

		fmt.Fprintf(color.Output, "%s %s%s%s %s %s\n", date, hour, min, sec, macbLine, tsEntry.Entry.Name)
		prev = tsEntry.Time
	}
}

func entryType(entry *bodyfile.TimeStampedEntry, check time.Time, c string) string {
	if entry.Time.Equal(check) {
		return c
	}
	return "."
}
