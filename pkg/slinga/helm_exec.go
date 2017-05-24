package slinga

import (
	"bufio"
	"github.com/golang/glog"
	"os/exec"
	"strings"
)

// HelmCodeExecutor is an executor that uses Helm for deployment of apps on kubernetes
type HelmCodeExecutor struct {
	Code *Code
}

func HelmName(str string) string {
	r := strings.NewReplacer("#", "-", "_", "-")
	return r.Replace(str)
}

// Install for HelmCodeExecutor runs "helm install" for the corresponding helm chart
func (executor HelmCodeExecutor) Install(key string, content map[string]map[string]string) error {
	uid := HelmName(key)

	chartName := content["chart"]["name"]

	// TODO(slukjanov): Replace with marshalling all params to temp file (YAML)
	setValues := ""
	if params, ok := content["params"]; ok {
		for key, value := range params {
			setValues += key + "=" + value + ","
		}
	}

	helmArgs := []string{"install", "--name", uid}
	if len(setValues) > 0 {
		helmArgs = append(helmArgs, "--set", setValues)
	}
	if version, ok := content["chart"]["version"]; ok {
		helmArgs = append(helmArgs, "--version", version)
	}
	if namespace, ok := content["chart"]["namespace"]; ok {
		helmArgs = append(helmArgs, "--namespace", namespace)
	} else {
		helmArgs = append(helmArgs, "--namespace", "aptomi")
	}
	helmArgs = append(helmArgs, chartName)

	return runHelmCmd(helmArgs...)
}

// Update for HelmCodeExecutor runs "helm update" for the corresponding helm chart
func (executor HelmCodeExecutor) Update(key string, labels LabelSet) error {
	// TODO: implement update method
	return nil
}

// Destroy for HelmCodeExecutor runs "helm delete" for the corresponding helm chart
func (executor HelmCodeExecutor) Destroy(key string) error {
	uid := HelmName(key)

	return runHelmCmd("delete", "--purge", uid)
}

func runHelmCmd(helmArgs ...string) error {
	return runCmd("helm", helmArgs...)
}

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
