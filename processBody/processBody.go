package processBody

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/airbus-cert/bodyfile"
	"github.com/mattn/go-isatty"
)

/*
parseHumanDate converts human-readable date formats to Unix timestamp.
*/
func parseHumanDate(dateStr string) (int64, error) {

	// Remove quotes if present.
	dateStr = strings.Trim(dateStr, `"'`)

	// Only allow dash format.
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Unix(), nil
		}
	}
	return 0, fmt.Errorf("unable to parse date: %s (use YYYY-MM-DD or YYYY-MM-DD HH:MM:SS)", dateStr)
}

/* processFilter converts human-readable date formats in filter expressions.
 */
func processFilter(filter string) (string, error) {

	// Convert slashes to dashes for user convenience.
	filter = strings.ReplaceAll(filter, "/", "-")

	// Match dash format only.
	re := regexp.MustCompile(`date(\s*)([<>=!]+)\s*["']?([0-9]{4}-[0-9]{1,2}-[0-9]{1,2}(?:\s+[0-9]{1,2}:[0-9]{1,2}(?::[0-9]{1,2})?)?)["']?`)
	matches := re.FindAllStringSubmatchIndex(filter, -1)
	if len(matches) == 0 {
		return filter, nil
	}
	result := filter

	// Replace from the end to avoid messing up indices.
	for i := len(matches) - 1; i >= 0; i-- {
		m := matches[i]
		space := filter[m[2]:m[3]]
		operator := filter[m[4]:m[5]]
		dateStr := filter[m[6]:m[7]]
		timestamp, err := parseHumanDate(dateStr)
		if err != nil {
			return "", fmt.Errorf("invalid date format in filter: %s", err)
		}
		replacement := fmt.Sprintf("date%s%s %d", space, operator, timestamp)
		result = result[:m[0]] + replacement + result[m[1]:]
	}
	return result, nil
}

/* GetInput checks if the input is from a terminal and returns the input file.
 */
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

func showFilterHelp(filter string) {
	helpText := fmt.Sprintf(`
No results found for filter: %s

Filter examples:
  hour > 12     (files modified after noon)
  hour < 6      (files modified before 6 AM)
  hour == 13    (files modified at 1 PM)
  day == 19     (files modified on the 19th)
  weekday == "Monday" (files modified on Monday)
  date > "2025-06-19 13:47:35" (files modified after specific date/time)
  date > "2025-06-19" (files modified after specific date)
  date > "2025/06/19 13:47:35" (slash format also supported)

Time ranges:
  hour < 1      = midnight to 1 AM (NOT 'less than 1 hour ago')
  hour > 12     = 1 PM to midnight
  hour >= 9 && hour <= 17 = 9 AM to 5 PM

IMPORTANT: 'hour < 1' means 'between midnight and 1 AM', not 'less than 1 hour ago'
For relative time (hours ago), use: date > "%s" (files modified in last hour)

Use -strict flag to show only matching timestamps instead of all timestamps for matching files.
`, filter, time.Now().Add(-1*time.Hour).Format("2006-01-02 15:04:05"))

	fmt.Fprint(os.Stderr, helpText)
}

/* ProcessBody processes the body file.
 */
func ProcessBody(f *os.File, flagProcess *bool, strict *bool, filter *string, modifiedFilter *string, accessFilter *string, ctimeFilter *string) {

	body := bodyfile.NewReader(f)

	// Determine which timestamp types to filter on.
	var timestampTypes []string
	var finalFilter string

	// Process the filters.
	switch {
	case *modifiedFilter != "":
		processedModified, err := processFilter(*modifiedFilter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Modified filter error: %s", err)
			os.Exit(2)
		}
		timestampTypes = append(timestampTypes, "modified")
		finalFilter = processedModified
	case *accessFilter != "":
		processedAccess, err := processFilter(*accessFilter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Access filter error: %s", err)
			os.Exit(2)
		}
		timestampTypes = append(timestampTypes, "access")
		finalFilter = processedAccess
	case *ctimeFilter != "":
		processedCtime, err := processFilter(*ctimeFilter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ctime filter error: %s", err)
			os.Exit(2)
		}
		timestampTypes = append(timestampTypes, "ctime")
		finalFilter = processedCtime
	case *filter != "":
		// Process human-readable dates in the filter
		processedFilter, err := processFilter(*filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Filter error: %s", err)
			os.Exit(2)
		}
		finalFilter = processedFilter
	}

	if finalFilter != "" {
		err := body.AddFilter(finalFilter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not add filter: %s", err)
			os.Exit(2)
		}
	}

	body.Strict = *strict

	count, err := body.Slurp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read all the content: %s", err)
		os.Exit(3)
	}

	// If no results found and a filter was used, show helpful message.
	if count == 0 && (*filter != "" || *modifiedFilter != "" || *accessFilter != "" || *ctimeFilter != "") {
		showFilterHelp(*filter)
		return
	}

	// Iterate through the body file.
	for {
		tsEntry, err := body.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while iterating: %s", err)
			os.Exit(4)
		}

		// Apply timestamp-specific filtering if needed
		if len(timestampTypes) > 0 {
			shouldShow := false
			for _, tsType := range timestampTypes {
				switch tsType {
				case "modified":
					if tsEntry.Time.Equal(tsEntry.Entry.ModificationTime) {
						shouldShow = true
					}
				case "access":
					if tsEntry.Time.Equal(tsEntry.Entry.AccessTime) {
						shouldShow = true
					}
				case "ctime":
					if tsEntry.Time.Equal(tsEntry.Entry.ChangeTime) {
						shouldShow = true
					}
				}
			}
			if !shouldShow {
				continue
			}
		}

		// Get the entry type.
		mChar := entryType(tsEntry, tsEntry.Entry.ModificationTime, "m")
		aChar := entryType(tsEntry, tsEntry.Entry.AccessTime, "a")
		cChar := entryType(tsEntry, tsEntry.Entry.ChangeTime, "c")
		bChar := entryType(tsEntry, tsEntry.Entry.CreationTime, "b")

		// Get the MACB line.
		macbLine := fmt.Sprintf("%s%s%s%s", mChar, aChar, cChar, bChar)

		// Get the date, hour, minute, and second.
		date := tsEntry.Time.Format("2006-01-02")
		hour := fmt.Sprintf("%02d", tsEntry.Time.Hour())
		min := fmt.Sprintf(":%02d", tsEntry.Time.Minute())
		sec := fmt.Sprintf(":%02d:", tsEntry.Time.Second())

		// Print the entry.
		fmt.Printf("%s %s%s%s %s %s\n", date, hour, min, sec, macbLine, tsEntry.Entry.Name)
	}
}

/* entryType returns the entry type.
 */
func entryType(entry *bodyfile.TimeStampedEntry, check time.Time, c string) string {
	if entry.Time.Equal(check) {
		return c
	}
	return "."
}
