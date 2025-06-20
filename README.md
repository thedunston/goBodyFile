# goBodyFile

**Status: Actively supported.**

goBodyFile collects metadata on files in a body file format that can be parsed by timeliner, mactime, or similar tools and create a forensic timeline.

### Background

I wrote a Python program in 2012 years ago that created a body file on Windows in order to generate a timeline for incident response investigations. Cygwin was used along with the requisite DLLs for it to run.  When I was teaching college, I had students create the program as a practical way of learning Python. I decided to rewrite it in Golang so that a body file can be created across multiple platforms with few dependencies. I would have them compile and use the [timeliner](https://github.com/airbus-cert/timeliner) program to see how to create a timeline for incident response investigations.

While converting the Python program to Golang, I decided to integrate the [Airbus-cert](https://github.com/airbus-cert) [timeliner](https://github.com/airbus-cert/timeliner) program to process the body file.

### Building

The project now uses a unified codebase with build tags for cross-platform support. Simply run:

This will automatically build for your current platform (Linux, Windows, or macOS).

For cross-platform builds, use the provided build script:

```bash
./build.sh
```
This will create binaries for:
- Linux (amd64)
- Windows (amd64)
- macOS (amd64)

### Using the program

https://wiki.sleuthkit.org/index.php?title=Body_file
````
The 3.X output has the following fields:

MD5|name|inode|mode_as_string|UID|GID|size|atime|mtime|ctime|crtime
````

### Run

The program must be run with the "-body" or "-process" option first. That will select which task to run.  The "-body" option will create a body file and the "-process" option will process a body file.

````
>> gobodyfile -h

Please select the '-body' or '-process' option.
````
See the required options by passing -h.

````
>> gobodyfile -body -h

Usage of ./gobodyfile:
  -body
        Create Body file
  -directory string
        Directory containing the files to collect metadata.
  -output string
        Output file name
  -sid
        (Optional) Display the SID. Default will return the UID and GID.
````
NOTE that sid is only available on Windows. Also, the process option will fail if the _sid_ option is selected in the body file because it expects a UID or GID. Without the -sid option on windows, the UID is used from the SID.

Example running on Windows:

````
>> .\gobodyfile.exe -body -directory "c:\Users\Public" -output file.txt

>> type .\file.txt

0|build.sh|5598450|493|1000|1000|50|1692410607|1692410607|1692410612|1692410612
0|common/common.go|5597946|420|1000|1000|812|1692410459|1692388523|1692410459|1692410459
0|createBody/createBody.go|5597949|420|1000|1000|2403|1692410459|1692402724|1692410459|1692410459
0|createBody/goStat|5597948|493|1000|1000|2313742|1692410417|1691797491|1692410459|1692410459
0|go.mod|5597953|420|1000|1000|323|1692410459|1692410459|1692410459|1692410459
0|go.sum|5597955|420|1000|1000|1382|1692410459|1692410459|1692410459|1692410459
0|gobodyfile-0.3.bin|5598313|493|1000|1000|2972696|1692411409|1692410564|1692410564|1692410564
0|main.go|5597961|420|1000|1000|2300|1692410459|1692410459|1692410459|1692410459
0|processBody/processBody.go|5597959|420|1000|1000|2296|1692410465|1692381314|1692410459|1692410459
0|version|5598200|420|1000|1000|4|1692410563|1692410526|1692410526|1692410526
0|.|5595701|2147484141|1000|1000|198|1692411997|1692412029|1692412029|1692412029
0|common|5597945|2147484141|1000|1000|18|1692410465|1692376770|1692410459|1692410459
0|createBody|5597947|2147484141|1000|1000|38|1692410465|1692378468|1692410459|1692410459
0|processBody|5597958|2147484141|1000|1000|28|1692410465|1692376581|1692410459|1692410459
````

The program timeliner can then be used to generate the timeline:

````
>> .\gobodyfile.exe -process file.txt

2023-08-11 23:44:51: m... createBody/goStat
2023-08-18 16:36:21: m... processBody
2023-08-18 16:39:30: m... common
2023-08-18 17:07:48: m... createBody
2023-08-18 17:55:14: m... processBody/processBody.go
2023-08-18 19:55:23: m... common/common.go
2023-08-18 23:52:04: m... createBody/createBody.go
2023-08-19 02:00:17: .a.. createBody/goStat
2023-08-19 02:00:59: ..cb processBody/processBody.go
2023-08-19 02:00:59: macb main.go
2023-08-19 02:00:59: ..cb createBody
2023-08-19 02:00:59: ..cb processBody
2023-08-19 02:00:59: macb go.sum
2023-08-19 02:00:59: ..cb common
2023-08-19 02:00:59: ..cb createBody/goStat
2023-08-19 02:00:59: .acb createBody/createBody.go
2023-08-19 02:00:59: .acb common/common.go
2023-08-19 02:00:59: macb go.mod
2023-08-19 02:01:05: .a.. common
2023-08-19 02:01:05: .a.. createBody
2023-08-19 02:01:05: .a.. processBody
2023-08-19 02:01:05: .a.. processBody/processBody.go
2023-08-19 02:02:06: m.cb version
2023-08-19 02:02:43: .a.. version
2023-08-19 02:02:44: m.cb gobodyfile-0.3.bin
2023-08-19 02:03:27: ma.. build.sh
2023-08-19 02:03:32: ..cb build.sh
2023-08-19 02:16:49: .a.. gobodyfile-0.3.bin
2023-08-19 02:26:37: .a.. .
2023-08-19 02:27:09: m.cb .

````

## Advanced Filtering Examples

goBodyFile supportsfiltering capabilities to aid your timeline analysis. The tool provides both general and timestamp specific filters.

### Filter Types

- **`-filter`**: Checks ALL timestamp types (access, modification, change, creation)
- **`-modified`**: Filters on modification time (mtime) only
- **`-access`**: Filters on access time (atime) only
- **`-ctime`**: Filters on change time (ctime) only
- **`-strict`**: Shows only matching timestamps instead of all timestamps for matching files

### Basic Filter Examples

**Show files modified after noon:**
```bash
./gobodyfile -process -filter "hour > 12" bodyfile.txt
```

**Show files modified on a specific day:**
```bash
./gobodyfile -process -filter "day == 19" bodyfile.txt
```

**Show files modified on Monday:**
```bash
./gobodyfile -process -filter 'weekday == "Monday"' bodyfile.txt
```

**Show files modified during business hours (9 AM to 5 PM):**
```bash
./gobodyfile -process -filter "hour >= 9 && hour <= 17" bodyfile.txt
```

### Date-Based Filtering

**Show files modified after a specific date:**
```bash
./gobodyfile -process -filter 'date > "2025-06-19"' bodyfile.txt
```

**Show files modified after a specific date and time:**
```bash
./gobodyfile -process -filter 'date > "2025-06-19 13:47:35"' bodyfile.txt
```

**Show files modified in the last hour:**
```bash
./gobodyfile -process -filter 'date > "2025-06-19 14:30:00"' bodyfile.txt
```

**Show files modified before a specific date:**
```bash
./gobodyfile -process -filter 'date < "2025-06-17"' bodyfile.txt
```

### Timestamp-Specific Filtering

**Show only files where modification time is before June 17th:**
```bash
./gobodyfile -process -modified 'date < "2025-06-17"' bodyfile.txt
```

**Show only files where access time is after June 19th:**
```bash
./gobodyfile -process -access 'date > "2025-06-19"' bodyfile.txt
```

**Show only files where change time equals June 16th:**
```bash
./gobodyfile -process -ctime 'date == "2025-06-16"' bodyfile.txt
```

### Understanding the Difference

**Regular filter (checks all timestamp types):**
```bash
./gobodyfile -process -filter 'date < "2025-06-17"' bodyfile.txt
```
This shows files where ANY timestamp (access, modification, change, creation) is before June 17th. You might see June 19th entries because their access time is displayed, even though their modification time matched the filter.

**Timestamp-specific filter (checks only specified timestamp type):**
```bash
./gobodyfile -process -modified 'date < "2025-06-17"' bodyfile.txt
```
This shows ONLY files where the modification time is before June 17th. No June 19th entries will appear.

**Strict mode (shows only matching timestamps):**
```bash
./gobodyfile -process -strict -filter 'date < "2025-06-17"' bodyfile.txt
```
This shows only the timestamps that actually matched the filter, not all timestamps for matching files.

### Real-World Investigation Examples

**Find files accessed during a suspected intrusion window:**
```bash
./gobodyfile -process -access 'date > "2025-06-19 13:00:00" && date < "2025-06-19 15:00:00"' bodyfile.txt
```

**Find files modified on the day of a security incident:**
```bash
./gobodyfile -process -modified 'date == "2025-06-19"' bodyfile.txt
```

**Find files that were created or modified after a system compromise:**
```bash
./gobodyfile -process -filter 'date > "2025-06-19 14:30:00"' bodyfile.txt
```

**Find files accessed during non-business hours:**
```bash
./gobodyfile -process -access 'hour < 6 || hour > 18' bodyfile.txt
```

**Find files modified on weekends:**
```bash
./gobodyfile -process -modified 'weekday == "Saturday" || weekday == "Sunday"' bodyfile.txt
```

**Find files with suspicious timing patterns (modified and accessed within minutes):**
```bash
./gobodyfile -process -modified 'date > "2025-06-19"' bodyfile.txt | \
./gobodyfile -process -access 'date > "2025-06-19"' bodyfile.txt
```

### Date Format Support

Both YYYY-MM-DD and YYYY/MM/DD formats are supported:

```bash
# These are equivalent:
./gobodyfile -process -filter 'date > "2025-06-19"' bodyfile.txt
./gobodyfile -process -filter 'date > "2025/06/19"' bodyfile.txt
```

### Combining Filters

You can combine multiple conditions in a single filter:

```bash
# Files modified on weekdays during business hours
./gobodyfile -process -modified 'weekday != "Saturday" && weekday != "Sunday" && hour >= 9 && hour <= 17' bodyfile.txt

# Files accessed after a specific date but before another date
./gobodyfile -process -access 'date > "2025-06-15" && date < "2025-06-20"' bodyfile.txt
```

The [timeliner GitHub repo](https://github.com/airbus-cert/timeliner) has information on using the process expression engine:

### Interpeting the output

[Mactime Output](https://wiki.sleuthkit.org/index.php?title=Mactime_output)
