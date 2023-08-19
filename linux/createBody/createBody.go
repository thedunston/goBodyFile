//go:build !windows

package createBody

import (
	"fmt"
	"goBody/common"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

func statFDir(toStat string, output string) {

	// Lstat is used to not follow symlinks and to get data on the symlink.
	theFile, err := os.Lstat(toStat)

	if err != nil {

		return

	}

	// Get the file's inode data.
	stat, ok := theFile.Sys().(*syscall.Stat_t)
	if !ok {

		return

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
	atime := time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))

	// Get the file's modification time.
	mtime := theFile.ModTime()

	// Get the file's creation time or on Unix, get the modification time for the inode.
	// ctime => Windows creation time of the file "birth of the file"
	// Unix ctime is when the attributes of the file changed.
	ctime := time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
	crtime := ctime

	// Create the file and open it.
	file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

	if err != nil {

		fmt.Println("Error opening file:", err)
		return

	}

	defer file.Close()

	// Write the results to a file.
	fmt.Fprintf(file, "0|%s|%d|%d|%d|%d|%d|%d|%d|%d|%d\n", toStat, inode, mode, uid, gid, fsize, atime.Unix(), mtime.Unix(), ctime.Unix(), crtime.Unix())
}

func CreateBody(rootDir string, outputFile string) {

	var fileList, dirList []string

	// Check if the ouput file exists.
	common.CheckFileExists(outputFile)

	// Check if the directory exists.
	common.CheckDirectoryExists(rootDir)

	// Crawl through the directory.
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {

		if err != nil {

			return err

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

		statFDir(file, outputFile)

	}

	// Loop through each directory and print the results.
	for _, dir := range dirList {

		statFDir(dir, outputFile)

	}

}
