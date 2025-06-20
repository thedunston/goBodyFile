//go:build windows

package createBody

/*
Returns the body file format based on the Sleuthkit 3.x format: https://wiki.sleuthkit.org/index.php?title=Body_file

"The body file is an intermediate file when creating a timeline of file activity. It is a pipe ("|") delimited text file that contains one line for each file (or other even type, such as a log or registry key). The fls, ils, and mac-robber tools all output this data format. The mactime tool reads this file and sorts the contents (therefore the format is sometimes referred to as the "mactime format").

The body file format in TSK 3.0+ is different from the format used in TSK 1.X and 2.X.

The 3.X output has the following fields:

MD5|name|inode|mode_as_string|UID|GID|size|atime|mtime|ctime|crtime
The times are reported in UNIX time format. Lines that start with '#' are ignored and treated as comments. In mactime, many of theses fields are optional. Its only requirement is that at least one of the time values is non-zero. The non-time values are simply printed as is. Other tools that read this file format may have different requirements.

The 2.X output has the following fields:

	MD5 | path/name | device | inode | mode_as_value | mode_as_string | num_of_links
	| UID | GID | rdev | size | atime | mtime | ctime | block_size | num_of_blocks"
*/

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

// Error log file handle.
var errorLog *os.File

/*
Initialize error logging.
*/
func initErrorLog(outputFile string) error {
	errorLogPath := outputFile + ".errors.log"
	var err error
	errorLog, err = os.OpenFile(errorLogPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to create error log file: %v", err)
	}
	return nil
}

/*
Close error log file.
*/
func closeErrorLog() {
	if errorLog != nil {
		errorLog.Close()
	}
}

/*
Log error to error log file.
*/
func logError(filename string, err error) {
	if errorLog != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		errorLog.WriteString(fmt.Sprintf("[%s] %s: %v\n", timestamp, filename, err))
	}
}

/*
Returns the files time to the UNIX Epoch.
*/
func filetimeToTime(ft windows.Filetime) time.Time {

	return time.Unix(0, ft.Nanoseconds())

}

/*
Returns error for file SID info.
*/
func checkSIDErr(err error) (string, string, error) {

	if err != nil {

		return "", "", err

	}

	return "", "", nil

}

/*
Check for errors when processing files.
*/
func checkProcessErr(err error) {

	if err != nil {
		fmt.Println(err)
	}

}

/*
Returns the file's user and group SID.
*/
func getSIDs(filename string, theSID bool) (string, string, error) {

	// Get the owner SID.
	ownerInfo, err := windows.GetNamedSecurityInfo(filename, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		return "", "", err
	}

	var uidSID string
	var groupSID string

	// Get the User SID.
	tmpUidSID, _, err := ownerInfo.Owner()
	if err != nil {
		return "", "", err
	}

	// Get the group security information.
	groupInfo, err := windows.GetNamedSecurityInfo(filename, windows.SE_FILE_OBJECT, windows.GROUP_SECURITY_INFORMATION)
	if err != nil {
		return "", "", err
	}

	// Get the Group SID.
	tmpGroupSID, _, err := groupInfo.Group()
	if err != nil {
		return "", "", err
	}

	// If theSID is true,
	if theSID {

		// split the SIDs based on the hypher and get the last value.
		x := strings.Split(tmpUidSID.String(), "-")
		uidSID = x[len(x)-1]

		y := strings.Split(tmpGroupSID.String(), "-")
		groupSID = y[len(y)-1]

	} else {

		uidSID = tmpUidSID.String()
		groupSID = tmpGroupSID.String()

	}

	return uidSID, groupSID, nil
}

/*
Return the file's permissions in octal format.
*/
func modeOctal(perm os.FileMode) string {

	return fmt.Sprintf("%04o", perm&07777)

}

/*
Returns the file the metadata for each file.
*/
func processFile(filename string, outputFile *os.File, theSID bool) error {

	// Fetch SIDs
	uidSID, groupSID, err := getSIDs(filename, theSID)
	if err != nil {
		logError(filename, fmt.Errorf("failed to get SIDs: %v", err))
		return err
	}

	// Check if the sid and gid have a hyphen.
	if strings.Contains(uidSID, "-") || strings.Contains(groupSID, "-") {

		// Get the last value.
		uidSID = strings.Split(uidSID, "-")[len(strings.Split(uidSID, "-"))-1]
		groupSID = strings.Split(groupSID, "-")[len(strings.Split(groupSID, "-"))-1]

	}

	// Fetch file's information using os package.
	fileInfo, err := os.Stat(filename)
	if err != nil {
		logError(filename, fmt.Errorf("failed to stat file: %v", err))
		return err
	}

	// Opens the file with no special privileges, don't lock the file (*FILE_SHARE*), and open the file even it is already open.
	getFileInfo, err := windows.CreateFile(&windows.StringToUTF16(filename)[0], 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		logError(filename, fmt.Errorf("failed to open file: %v", err))
		return err
	}

	defer windows.CloseHandle(getFileInfo)

	// Gets the files last access time, last write time, and creation time.
	var theFile windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(getFileInfo, &theFile); err != nil {
		logError(filename, fmt.Errorf("failed to get file information: %v", err))
		return err
	}

	// File mode in octal.
	fmode := modeOctal(fileInfo.Mode())

	// File's inode.
	inode := uint64(theFile.FileIndexHigh)<<32 + uint64(theFile.FileIndexLow)

	// FIle size.
	size := fileInfo.Size()

	// Last access time.
	atime := filetimeToTime(theFile.LastAccessTime).Unix()

	// Last Modification time.
	mtime := filetimeToTime(theFile.LastWriteTime).Unix()

	// Windows doesn't have a ctime like Unix so last write time used here.
	ctime := filetimeToTime(theFile.LastWriteTime).Unix()

	// Creation time.
	crtime := filetimeToTime(theFile.CreationTime).Unix()

	// Write the body file format to the output file.
	_, err = outputFile.WriteString(fmt.Sprintf("0|%s|%d|%s|%s|%s|%d|%d|%d|%d|%d\n", filename, inode, fmode, uidSID, groupSID, size, atime, mtime, ctime, crtime))
	if err != nil {
		logError(filename, fmt.Errorf("failed to write to output file: %v", err))
		return err
	}

	return nil
}

func CreateBody(rootDir string, outputFile string, theSID bool) {

	// Initialize error logging.
	if err := initErrorLog(outputFile); err != nil {
		fmt.Printf("Warning: Could not create error log: %v\n", err)
	}
	defer closeErrorLog()

	// If s is not set to true, set it to false.
	if !theSID {
		theSID = false
	}

	// Check that the directory exists.
	if _, err := os.Stat(rootDir); err != nil {
		fmt.Println("The directory does not exist.")
		os.Exit(1)
	}

	// Check that the output file doesn't exist.
	if _, err := os.Stat(outputFile); err == nil {
		fmt.Println("The output file already exists. Do you want to delete it? (y/n)")

		var answer string
		fmt.Scanln(&answer)

		// If so, then prompt y or n to delete it.
		if answer == "y" {

			os.Remove(outputFile)

			// Otherwise, exit.
		} else {

			fmt.Println("The output file was not deleted.")

			os.Exit(1)

		}

	}

	// Walk through the directory and each of its subdirectories.
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		// If there's an error accessing the directory/file, log it and continue.
		if err != nil {
			logError(path, fmt.Errorf("failed to access path: %v", err))
			return nil
		}

		// Only process files.
		if !info.IsDir() {

			// Open the output file in append mode.
			file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				logError(path, fmt.Errorf("failed to open output file: %v", err))
				return nil
			}
			defer file.Close()

			// Get the file's metadata.
			if err := processFile(path, file, theSID); err != nil {
				// Error already logged in processFile, just continue.
				return nil
			}

		}

		return nil

	})

	if err != nil {
		fmt.Printf("Error during directory walk: %v\n", err)
	}

}
