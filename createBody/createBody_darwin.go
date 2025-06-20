//go:build darwin

package createBody

import (
	"fmt"
	"gobodyfile/common"
	"os"
	"path/filepath"
	"syscall"
	"time"
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
statFDir is used to get the file system information.
*/
func statFDir(toStat string, output string) error {

	// Lstat is used to not follow symlinks and to get data on the symlink.
	theFile, err := os.Lstat(toStat)

	if err != nil {
		logError(toStat, fmt.Errorf("failed to stat file: %v", err))
		return err
	}

	// Get the file's inode data.
	stat, ok := theFile.Sys().(*syscall.Stat_t)
	if !ok {
		logError(toStat, fmt.Errorf("failed to get file system info"))
		return fmt.Errorf("failed to get file system info")
	}

	// Inode data.

	// File's mode.
	mode := theFile.Mode()

	// Inode number.
	inode := stat.Ino

	// User ID.
	uid := stat.Uid

	// Group ID.
	gid := stat.Gid

	// Get the file size.
	fsize := theFile.Size()

	// Get the file's atime.
	atime := time.Unix(stat.Atimespec.Sec, stat.Atimespec.Nsec)

	// Get the file's modification time.
	mtime := theFile.ModTime()

	// Get the file's creation time or on Unix, get the modification time for the inode.
	// ctime => Windows creation time of the file "birth of the file"
	// Unix ctime is when the attributes of the file changed.
	ctime := time.Unix(stat.Ctimespec.Sec, stat.Ctimespec.Nsec)
	crtime := ctime

	// Create the file and open it.
	file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

	if err != nil {
		logError(toStat, fmt.Errorf("failed to open output file: %v", err))
		return err
	}

	defer file.Close()

	// Write the results to a file.
	_, err = fmt.Fprintf(file, "0|%s|%d|%d|%d|%d|%d|%d|%d|%d|%d\n", toStat, inode, mode, uid, gid, fsize, atime.Unix(), mtime.Unix(), ctime.Unix(), crtime.Unix())
	if err != nil {
		logError(toStat, fmt.Errorf("failed to write to output file: %v", err))
		return err
	}

	return nil
}

func CreateBody(rootDir string, outputFile string, sid bool) {

	// Initialize error logging.
	if err := initErrorLog(outputFile); err != nil {
		fmt.Printf("Warning: Could not create error log: %v\n", err)
	}
	defer closeErrorLog()

	var fileList, dirList []string

	// Check if the output file exists.
	common.CheckFileExists(outputFile)

	// Check if the directory exists.
	common.CheckDirectoryExists(rootDir)

	// Crawl through the directory.
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			logError(path, fmt.Errorf("failed to access path: %v", err))
			return nil
		}

		// Add directories to the directory slice.
		if info.IsDir() {

			dirList = append(dirList, path)

		} else {

			// Add files to the file slice.
			fileList = append(fileList, path)
		}

		return nil

	})

	if err != nil {

		fmt.Printf("Unable to crawl through the directory: %s\n", err)
		return

	}

	// Loop through each file and print the results.
	for _, file := range fileList {

		if err := statFDir(file, outputFile); err != nil {
			// Error already logged in statFDir, just continue
			continue
		}

	}

	// Loop through each directory and print the results.
	for _, dir := range dirList {

		if err := statFDir(dir, outputFile); err != nil {
			// Error already logged in statFDir, just continue.
			continue
		}

	}

}
