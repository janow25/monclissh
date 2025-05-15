package metrics

import (
    "fmt"
    "os/exec"
    "strings"
)

// GetCPUUsage retrieves the current CPU usage percentage from the remote host.
func GetCPUUsage(host string) (string, error) {
    cmd := exec.Command("ssh", host, "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1}'")
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to get CPU usage from host %s: %w", host, err)
    }
    return strings.TrimSpace(string(output)), nil
}