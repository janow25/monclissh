package metrics

import (
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

	// fmt.Println("Starting metrics collection...")

	for _, host := range c.hosts {
		wg.Add(1)
		go func(h config.Host) {
			defer wg.Done()

			client, err := ssh.NewSSHClientFromConfig(h)
			if err != nil {
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
			defer client.Close()

			cpu, disk, memory, err := collectAllMetrics(client)
			if err != nil {
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

func collectAllMetrics(client *ssh.SSHClient) (float64, float64, float64, error) {
	cpuCmd := "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1}'"
	diskCmd := "df / | tail -1 | awk '{print $5}' | sed 's/%//'"
	memCmd := "free | grep Mem | awk '{print $3/$2 * 100.0}'"

	cpuOut, err := client.ExecuteCommand(cpuCmd)
	if err != nil {
		return 0, 0, 0, err
	}
	diskOut, err := client.ExecuteCommand(diskCmd)
	if err != nil {
		return 0, 0, 0, err
	}
	memOut, err := client.ExecuteCommand(memCmd)
	if err != nil {
		return 0, 0, 0, err
	}

	cpu, err := strconv.ParseFloat(strings.TrimSpace(cpuOut), 64)
	if err != nil {
		return 0, 0, 0, err
	}
	disk, err := strconv.ParseFloat(strings.TrimSpace(diskOut), 64)
	if err != nil {
		return 0, 0, 0, err
	}
	memory, err := strconv.ParseFloat(strings.TrimSpace(memOut), 64)
	if err != nil {
		return 0, 0, 0, err
	}

	return cpu, disk, memory, nil
}
