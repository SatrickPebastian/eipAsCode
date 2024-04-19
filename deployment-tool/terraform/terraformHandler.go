package terraform

import (
    //"fmt"
    //"io/ioutil"
    //"strings"
    "os"
    "os/exec"
)

func CheckTerraformInstalled() bool {
	cmd := exec.Command("terraform", "version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func ApplyTerraform() error {
	cmd := exec.Command("terraform", "init")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("terraform", "apply", "-auto-approve")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
