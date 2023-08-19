//go:build windows

package createBodyWin

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
		panic(err)
	}

}

/*
Returns the file's user and group SID.
*/
func getSIDs(filename string, theSID bool) (string, string, error) {

	// Get the owner SID
	ownerInfo, err := windows.GetNamedSecurityInfo(filename, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION)
	checkSIDErr(err)

	var uidSID string
	var groupSID string

	// Get the User SID.
	tmpUidSID, _, err := ownerInfo.Owner()
	checkSIDErr(err)

	// Get the group security information.
	groupInfo, err := windows.GetNamedSecurityInfo(filename, windows.SE_FILE_OBJECT, windows.GROUP_SECURITY_INFORMATION)
	checkSIDErr(err)

	// Get the Group SID.
	tmpGroupSID, _, err := groupInfo.Group()
	checkSIDErr(err)

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
func processFile(filename string, outputFile *os.File, theSID bool) {

	// Fetch SIDs
	uidSID, groupSID, err := getSIDs(filename, theSID)
	checkProcessErr(err)

	// Fetch file's information using os package
	fileInfo, err := os.Stat(filename)
	checkProcessErr(err)

	// Opens the file with no special privileges, don't lock the file (*FILE_SHARE*), and open the file even it is already open.
	getFileInfo, err := windows.CreateFile(&windows.StringToUTF16(filename)[0], 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	checkProcessErr(err)

	defer windows.CloseHandle(getFileInfo)

	// Gets the files last access time, last write time, and creation time.
	var theFile windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(getFileInfo, &theFile); err != nil {
		panic(err)
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
	outputFile.WriteString(fmt.Sprintf("0|%s|%d|%s|%s|%s|%d|%d|%d|%d|%d\n", filename, inode, fmode, uidSID, groupSID, size, atime, mtime, ctime, crtime))

}

func CreateBodyfileWin(rootDir string, outputFile string, theSID bool) {

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
		checkProcessErr(err)

		// Only process files.
		if !info.IsDir() {

			// Open the output file in append mode.
			file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				fmt.Println("Error opening file:", err)

			}
			defer file.Close()

			// Get the file's metadata.
			processFile(path, file, theSID)

		}

		return nil

	})

	checkProcessErr(err)

}
