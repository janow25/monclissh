package dashboard

import (
    "fmt"
    "time"

    "monclissh/internal/config"
)

type HostMetrics struct {
    Hostname string
    CPU      float64
    Disk     float64
    Memory   float64
}

type Dashboard struct {
    Hosts []HostMetrics
}

func NewDashboard(hosts []config.Host) *Dashboard {
    var metrics []HostMetrics
    for _, host := range hosts {
        metrics = append(metrics, HostMetrics{Hostname: host.Name})
    }
    return &Dashboard{Hosts: metrics}
}

func (d *Dashboard) UpdateMetrics() {
    for i := range d.Hosts {
        // Here you would call the metrics collection functions
        // For example: d.Hosts[i].CPU = collectCPU(d.Hosts[i].Hostname)
        // d.Hosts[i].Disk = collectDisk(d.Hosts[i].Hostname)
        // d.Hosts[i].Memory = collectMemory(d.Hosts[i].Hostname)
    }
}

func (d *Dashboard) Display() {
    for _, host := range d.Hosts {
        fmt.Printf("Host: %s, CPU: %.2f%%, Disk: %.2f%%, Memory: %.2f%%\n",
            host.Hostname, host.CPU, host.Disk, host.Memory)
    }
}

func (d *Dashboard) Start(refreshInterval time.Duration) {
    for {
        d.UpdateMetrics()
        d.Display()
        time.Sleep(refreshInterval)
    }
}

func Start(cfg *config.HostConfig) {
    dashboard := NewDashboard(cfg.Hosts)
    dashboard.Start(5 * time.Second)
}