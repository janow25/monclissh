package metrics

import (
    "fmt"
    "sync"
)

type Metrics struct {
    CPU    float64
    Disk   float64
    Memory float64
}

type Collector struct {
    hosts []string
    mu    sync.Mutex
}

func NewCollector(hosts []string) *Collector {
    return &Collector{hosts: hosts}
}

func (c *Collector) Collect() (map[string]Metrics, error) {
    metrics := make(map[string]Metrics)
    var wg sync.WaitGroup

    for _, host := range c.hosts {
        wg.Add(1)
        go func(h string) {
            defer wg.Done()
            cpu, err := collectCPU(h)
            if err != nil {
                fmt.Printf("Error collecting CPU for host %s: %v\n", h, err)
                return
            }
            disk, err := collectDisk(h)
            if err != nil {
                fmt.Printf("Error collecting Disk for host %s: %v\n", h, err)
                return
            }
            memory, err := collectMemory(h)
            if err != nil {
                fmt.Printf("Error collecting Memory for host %s: %v\n", h, err)
                return
            }

            c.mu.Lock()
            metrics[h] = Metrics{
                CPU:    cpu,
                Disk:   disk,
                Memory: memory,
            }
            c.mu.Unlock()
        }(host)
    }

    wg.Wait()
    return metrics, nil
}

// Placeholder functions for collecting metrics
func collectCPU(host string) (float64, error) {
    // Implement CPU collection logic
    return 0.0, nil
}

func collectDisk(host string) (float64, error) {
    // Implement Disk collection logic
    return 0.0, nil
}

func collectMemory(host string) (float64, error) {
    // Implement Memory collection logic
    return 0.0, nil
}