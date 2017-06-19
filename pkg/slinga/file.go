package slinga

import (
	log "github.com/Sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"fmt"
)

// copyFileContents copies the contents of the file named src to the file named
// by dst. Dst will be overwritten if already exists. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFile(src, dst string) (err error) {
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

// Recursively copies a directory tree
// Source directory must exist, destination directory must not exist
func copyDirectory(srcDir string, dstDir string) (err error) {
	// get properties of source directory
	srcStat, err := os.Stat(srcDir)
	if err != nil {
		return err
	}
	if !srcStat.IsDir() {
		return fmt.Errorf("Source is not a directory")
	}

	// ensure destination directory does not already exist
	_, err = os.Open(dstDir)
	if !os.IsNotExist(err) {
		return fmt.Errorf("Destination directory already exists")
	}

	// create destination dir
	err = os.MkdirAll(dstDir, srcStat.Mode())
	if err != nil {
		return err
	}

	// get all entries in a directory
	entries, err := ioutil.ReadDir(srcDir)
	for _, entry := range entries {
		sfp := filepath.Join(srcDir, entry.Name())
		dfp := filepath.Join(dstDir, entry.Name())
		if entry.IsDir() {
			err = copyDirectory(sfp, dfp)
			if err != nil {
				return err
			}
		} else {
			// perform copy
			err = copyFile(sfp, dfp)
			if err != nil {
				return err
			}
		}

	}
	return nil
}


// deleteFile deletes a file
func deleteFile(src string) (err error) {
	return os.Remove(src)
}

// deleteDirectoryContents removes all contents of a directory
func deleteDirectoryContents(dir string) error {
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

func writeTempFile(prefix string, content string) *os.File {
	tmpFile, err := ioutil.TempFile("", "aptomi-"+prefix)
	if err != nil {
		debug.WithFields(log.Fields{
			"prefix": prefix,
			"error":  err,
		}).Fatal("Failed to create temp file")
	}

	_, err = tmpFile.Write([]byte(content))
	if err != nil {
		debug.WithFields(log.Fields{
			"file":    tmpFile.Name(),
			"content": content,
			"error":   err,
		}).Fatal("Failed to write to temp file")
	}

	return tmpFile
}

// Ensures that only one file matches the list of files
func ensureSingleFile(files []string) (string, error) {
	if len(files) <= 0 {
		return "", fmt.Errorf("No files found")
	}
	if len(files) > 1 {
		return "", fmt.Errorf("More than one file found")
	}
	return files[0], nil
}
