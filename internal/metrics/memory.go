package metrics

import (
    "fmt"
    "github.com/shirou/gopsutil/mem"
)

// CollectMemoryMetrics retrieves the RAM usage metrics from the host.
func CollectMemoryMetrics() (string, error) {
    v, err := mem.VirtualMemory()
    if err != nil {
        return "", fmt.Errorf("failed to get memory metrics: %v", err)
    }

    return fmt.Sprintf("Total: %v MB, Free: %v MB, Used: %v MB, Usage: %.2f%%",
        v.Total/1024/1024, v.Free/1024/1024, v.Used/1024/1024, v.UsedPercent), nil
}