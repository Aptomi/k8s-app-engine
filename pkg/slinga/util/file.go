package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// CopyFile copies the contents of the file named src to the file named
// by dst. Dst will be overwritten if already exists. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

// CopyDirectory recursively copies a directory tree
// Source directory must exist, destination directory must not exist
func CopyDirectory(srcDir string, dstDir string) (err error) {
	// get properties of source directory
	srcStat, err := os.Stat(srcDir)
	if err != nil {
		return err
	}
	if !srcStat.IsDir() {
		return fmt.Errorf("Source is not a directory")
	}

	/*
		// ensure destination directory does not already exist
		_, err = os.Open(dstDir)
		if !os.IsNotExist(err) {
			return fmt.Errorf("Destination directory already exists")
		}
	*/

	// create destination dir
	err = os.MkdirAll(dstDir, srcStat.Mode())
	if err != nil {
		return err
	}

	// get all entries in a directory
	entries, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sfp := filepath.Join(srcDir, entry.Name())
		dfp := filepath.Join(dstDir, entry.Name())
		if entry.IsDir() {
			err = CopyDirectory(sfp, dfp)
			if err != nil {
				return err
			}
		} else {
			// perform copy
			err = CopyFile(sfp, dfp)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

// DeleteFile deletes a file
func DeleteFile(src string) (err error) {
	return os.Remove(src)
}

// DeleteDirectoryContents removes all contents of a directory
func DeleteDirectoryContents(dir string) error {
	d, err := os.Open(dir)
	if os.IsNotExist(err) {
		// do nothing
		return nil
	}
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

// WriteTempFile creates a temporary file
func WriteTempFile(prefix string, content string) *os.File {
	tmpFile, err := ioutil.TempFile("", "aptomi-"+prefix)
	if err != nil {
		panic("Failed to create temp file")
	}

	_, err = tmpFile.Write([]byte(content))
	if err != nil {
		panic("Failed to write to temp file")
	}

	return tmpFile
}

// EnsureSingleFile ensures that only one file matches the list of files
func EnsureSingleFile(files []string) (string, error) {
	if len(files) <= 0 {
		return "", fmt.Errorf("No files found")
	}
	if len(files) > 1 {
		return "", fmt.Errorf("More than one file found")
	}
	return files[0], nil
}
