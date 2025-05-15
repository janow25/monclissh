package metrics

import (
    "fmt"
    "os/exec"
    "strings"
)

// DiskUsage represents the disk usage metrics.
type DiskUsage struct {
    Filesystem string
    Size       string
    Used       string
    Available  string
    UsePercent string
}

// GetDiskUsage retrieves the disk usage metrics from the specified host.
func GetDiskUsage(host string) ([]DiskUsage, error) {
    cmd := exec.Command("ssh", host, "df -h")
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to execute command on host %s: %w", host, err)
    }

    return parseDiskUsage(string(output)), nil
}

// parseDiskUsage parses the output of the df command into a slice of DiskUsage.
func parseDiskUsage(output string) []DiskUsage {
    var usages []DiskUsage
    lines := strings.Split(output, "\n")

    for _, line := range lines[1:] {
        if line == "" {
            continue
        }
        fields := strings.Fields(line)
        if len(fields) >= 6 {
            usages = append(usages, DiskUsage{
                Filesystem: fields[0],
                Size:       fields[1],
                Used:       fields[2],
                Available:  fields[3],
                UsePercent: fields[4],
            })
        }
    }

    return usages
}