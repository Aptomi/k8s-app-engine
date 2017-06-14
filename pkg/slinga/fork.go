package slinga

import (
	"bufio"
	"github.com/golang/glog"
	"os/exec"
)

func runCmd(cmdName string, cmdArgs ...string) error {
	cmd := exec.Command(cmdName, cmdArgs...)
	glog.Infof("Running command '%s' with args: %s", cmdName, cmdArgs)

	cmdStdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		glog.Errorf("Failed running command '%s' (with args: %s): %s", cmdName, cmdArgs, err)
		return err
	}

	cmdStdoutScanner := bufio.NewScanner(cmdStdoutReader)
	go func() {
		for cmdStdoutScanner.Scan() {
			glog.Infof("%s out | %s\n", cmdName, cmdStdoutScanner.Text())
		}
	}()

	cmdStderrReader, err := cmd.StderrPipe()
	if err != nil {
		glog.Errorf("Failed running command '%s' (with args: %s): %s", cmdName, cmdArgs, err)
		return err
	}

	cmdStderrScanner := bufio.NewScanner(cmdStderrReader)
	go func() {
		for cmdStderrScanner.Scan() {
			glog.Infof("%s err | %s\n", cmdName, cmdStderrScanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		glog.Errorf("Failed running command '%s' (with args: %s): %s", cmdName, cmdArgs, err)
		return err
	}

	err = cmd.Wait()
	if err != nil {
		glog.Errorf("Failed running command '%s' (with args: %s): %s", cmdName, cmdArgs, err)
		return err
	}

	return nil
}
