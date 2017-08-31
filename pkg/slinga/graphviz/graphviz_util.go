package graphviz

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"os/exec"
)

func OpenImage(image image.Image) {
	tmpFile, err := ioutil.TempFile("", "aptomi-graphviz-debug")
	defer tmpFile.Close()

	if err != nil {
		panic("Failed to create temp file")
	}

	err = png.Encode(tmpFile, image)
	if err != nil {
		panic("Failed to write to temp file")
	}

	// Call open on the image
	{
		cmd := "open"
		args := []string{tmpFile.Name()}
		command := exec.Command(cmd, args...)
		var outb, errb bytes.Buffer
		command.Stdout = &outb
		command.Stderr = &errb
		if err := command.Run(); err != nil || len(errb.String()) > 0 {
			panic(fmt.Sprintf("Unable to open '%s': %s", tmpFile.Name(), err.Error()))
		}
	}
}
