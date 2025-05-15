package dashboard

import (
    "fmt"
    "time"

    "monclissh/internal/config"
    "monclissh/internal/metrics"
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

func (d *Dashboard) UpdateMetrics(collector *metrics.Collector) {
    collectedMetrics, err := collector.Collect()
    if err != nil {
        fmt.Printf("Error collecting metrics: %v\n", err)
        return
    }

    for i := range d.Hosts {
        host := d.Hosts[i]
        if metrics, ok := collectedMetrics[host.Hostname]; ok {
            d.Hosts[i].CPU = metrics.CPU
            d.Hosts[i].Disk = metrics.Disk
            d.Hosts[i].Memory = metrics.Memory
        }
    }
}

func (d *Dashboard) Display() {
    for _, host := range d.Hosts {
        fmt.Printf("Host: %s, CPU: %.2f%%, Disk: %.2f%%, Memory: %.2f%%\n",
            host.Hostname, host.CPU, host.Disk, host.Memory)
    }
}

func Start(cfg *config.HostConfig) {
    collector := metrics.NewCollector(cfg.Hosts)
    dashboard := NewDashboard(cfg.Hosts)

    for {
        dashboard.UpdateMetrics(collector)
        dashboard.Display()
        time.Sleep(100 * time.Millisecond)
    }
}