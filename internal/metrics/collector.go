package metrics

import (
    "fmt"
    "monclissh/internal/config"
    "monclissh/internal/ssh"
    "strconv"
    "strings"
    "sync"
)	

type Metrics struct {
		Error  string
		CPU    float64
    Disk   float64
    Memory float64
}

type Collector struct {
    hosts []config.Host
    mu    sync.Mutex
}

func NewCollector(hosts []config.Host) *Collector {
    return &Collector{hosts: hosts}
}

func (c *Collector) Collect() (map[string]Metrics, error) {
    metrics := make(map[string]Metrics)
    var wg sync.WaitGroup

    for _, host := range c.hosts {
        wg.Add(1)
        go func(h config.Host) {
            defer wg.Done()

            cpu, err := collectCPUFromConfig(h)
            if err != nil {
                // fmt.Printf("Error collecting CPU for host %s: %v\n", h.Name, err)
								c.mu.Lock()
								metrics[h.Name] = Metrics{
									Error:  err.Error(),
									CPU:    0,
									Disk:   0,
									Memory: 0,
								}
								c.mu.Unlock()
                return
            }
            disk, err := collectDiskFromConfig(h)
            if err != nil {
                fmt.Printf("Error collecting Disk for host %s: %v\n", h.Name, err)
                return
            }
            memory, err := collectMemoryFromConfig(h)
            if err != nil {
                fmt.Printf("Error collecting Memory for host %s: %v\n", h.Name, err)
                return
            }

            c.mu.Lock()
            metrics[h.Name] = Metrics{
								Error:  "",
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

func collectCPUFromConfig(hostConfig config.Host) (float64, error) {
    client, err := ssh.NewSSHClientFromConfig(hostConfig)
    if err != nil {
        return 0, err
    }
    defer client.Close()

    output, err := client.ExecuteCommand("top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1}'")
    if err != nil {
        return 0, err
    }

    usage, err := strconv.ParseFloat(strings.TrimSpace(output), 64)
    if err != nil {
        return 0, err
    }

    return usage, nil
}

func collectDiskFromConfig(hostConfig config.Host) (float64, error) {
    client, err := ssh.NewSSHClientFromConfig(hostConfig)
    if err != nil {
        return 0, err
    }
    defer client.Close()

    output, err := client.ExecuteCommand("df / | tail -1 | awk '{print $5}' | sed 's/%//'")
    if err != nil {
        return 0, err
    }

    usage, err := strconv.ParseFloat(strings.TrimSpace(output), 64)
    if err != nil {
        return 0, err
    }

    return usage, nil
}

func collectMemoryFromConfig(hostConfig config.Host) (float64, error) {
    client, err := ssh.NewSSHClientFromConfig(hostConfig)
    if err != nil {
        return 0, err
    }
    defer client.Close()

    output, err := client.ExecuteCommand("free | grep Mem | awk '{print $3/$2 * 100.0}'")
    if err != nil {
        return 0, err
    }

    usage, err := strconv.ParseFloat(strings.TrimSpace(output), 64)
    if err != nil {
        return 0, err
    }

    return usage, nil
}
