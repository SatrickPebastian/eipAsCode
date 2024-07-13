package target

import (
    "bytes"
    "os/exec"
)

// CheckComposeInstalled checks if Docker Compose is installed by attempting to run `docker-compose version`.
func CheckComposeInstalled() bool {
    cmd := exec.Command("docker-compose", "version")
    if err := cmd.Run(); err != nil {
        return false
    }
    return true
}

// RunComposeCommand runs a Docker Compose command with the given arguments and returns the output or any errors.
func RunComposeCommand(args ...string) (string, error) {
    cmd := exec.Command("docker-compose", args...)
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

// ApplyCompose applies the configurations in docker-compose.yaml.
func ApplyCompose() (string, error) {
    return RunComposeCommand("up", "-d") // Detached mode
}

// DestroyCompose stops and removes containers, networks, volumes, and images created by `up`.
func DestroyCompose() (string, error) {
    return RunComposeCommand("down")
}
