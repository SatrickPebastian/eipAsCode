package terraform

import (
    "bytes" // Add this import to use bytes.Buffer
    "os/exec"
	"diploy/parser"
)

func CheckTerraformInstalled() bool {
	cmd := exec.Command("terraform", "version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func runTerraformCommand(args ...string) (string, error) {
    cmd := exec.Command("terraform", args...)
    var out bytes.Buffer
    var stderr bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &stderr
    err := cmd.Run()
    if err != nil {
        return stderr.String(), err
    }
    return out.String(), nil
}

func ApplyTerraform(config parser.Config) (string, error) {
    if _, err := runTerraformCommand("init"); err != nil {
        return "", err
    }
    return runTerraformCommand("apply", "-auto-approve")
}

func DestroyTerraform(config parser.Config) (string, error) {
    if _, err := runTerraformCommand("init"); err != nil {
        return "", err
    }
    return runTerraformCommand("destroy", "-auto-approve")
}
