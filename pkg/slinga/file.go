package slinga

import (
	"fmt"
	"io"
	"os"
)

// copyFileContents copies the contents of the file named src to the file named
// by dst. Error will be thrown if the dst file already exists. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFile(src, dst string) (err error) {
	if stat, err := os.Stat(dst); err == nil && !stat.IsDir() {
		return fmt.Errorf("File %s already exists", dst)
	}

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
