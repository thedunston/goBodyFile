# goBodyFile
Create a body file and timeline for incident response investigations.

### Background

I wrote a python program around 11 years ago that created a body file on Windows in order to generate a timeline for incident response investigations. Cygwin was used along with the requisite DLLs for it to run.  I decided to rewrite it in Golang so that a body file can be created across multiple platforms with few dependencies. When I was teaching college, I had students create the program as a practical way of learning Python.  I would have them compile and use the [timeliner](https://github.com/airbus-cert/timeliner) program to see how to create a timeliner for incident response investigations.

While converting the Python program to Golang, I decided to integrate the [Airbus-cert](https://github.com/airbus-cert) [timeliner](https://github.com/airbus-cert/timeliner) program to process the body file.

### Using the program

Go body file collects metadata on files in a body file format that can be parsed by timeliner, mactime, or similar tools and create a forensic timeline.

https://wiki.sleuthkit.org/index.php?title=Body_file
````
The 3.X output has the following fields:

MD5|name|inode|mode_as_string|UID|GID|size|atime|mtime|ctime|crtime
````

### Run

The program must be run with the "-body" or "-process" option first. That will select which task to run.  The "-body" option will create a body file and the -process option will process a body file.

````
>> goBodyFile.bin -h

Please select the '-body' or '-process' option.
````
See the required options by passing -h.

````
>> goBodyFile.bin -body -h

Usage of ./goBodyFile-0.3.bin:
  -body
        Create Body file
  -directory string
        Directory containing the files to collect metadata.
  -output string
        Output file name
  -sid
        (Optional) Display the SID. Default will return the UID and GID.
````
NOTE that sid is only available on Windows. Also, the process option will fail if the _sid_ option is selected in the body file because it expects a UID or GID.

Example running on Windows:

````
>> .\goBodyFile.exe -body -directory "c:\Users\Public" -output file.txt

>> type .\file.txt

0|go.mod|562949953873706|0666|1001|513|59|1692321674|1692293485|1692293485|1692293467
0|go.sum|562949953874515|0666|1001|513|153|1692321674|1692293485|1692293485|1692293485
0|goStat.go|844424930583991|0666|1001|513|6833|1692321675|1692321660|1692321660|1692293407
````

The program timeliner can then be used to generate the timeline:

````
>> .\goBodyFile-0.3.exe -process file.txt

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
2023-08-19 02:02:44: m.cb goBodyFile-0.3.bin
2023-08-19 02:03:27: ma.. build.sh
2023-08-19 02:03:32: ..cb build.sh
2023-08-19 02:16:49: .a.. goBodyFile-0.3.bin
2023-08-19 02:24:25: .a.. .main.go.swp
2023-08-19 02:24:31: m.cb .main.go.swp
2023-08-19 02:26:37: .a.. .
2023-08-19 02:27:09: m.cb .

````

The [timeliner GitHub repo](https://github.com/airbus-cert/timeliner) has information on using the process expression engine:

### Interpeting the output

[Mactime Output](https://wiki.sleuthkit.org/index.php?title=Mactime_output)
