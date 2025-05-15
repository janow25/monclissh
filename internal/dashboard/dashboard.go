package dashboard

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/guptarohit/asciigraph"
	"github.com/rivo/tview"

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

func (d *Dashboard) DisplayWithPieCharts() {
	for _, host := range d.Hosts {
		fmt.Printf("Host: %s\n", host.Hostname)

		// CPU Pie Chart
		cpuGraph := asciigraph.Plot([]float64{host.CPU, 100 - host.CPU}, asciigraph.Width(20), asciigraph.Height(10), asciigraph.Caption("CPU Usage"))
		fmt.Println(cpuGraph)

		// Disk Pie Chart
		diskGraph := asciigraph.Plot([]float64{host.Disk, 100 - host.Disk}, asciigraph.Width(20), asciigraph.Height(10), asciigraph.Caption("Disk Usage"))
		fmt.Println(diskGraph)

		// Memory Pie Chart
		memoryGraph := asciigraph.Plot([]float64{host.Memory, 100 - host.Memory}, asciigraph.Width(20), asciigraph.Height(10), asciigraph.Caption("Memory Usage"))
		fmt.Println(memoryGraph)

		fmt.Println("--------------------------------")
	}
}

func Start(cfg *config.HostConfig) {
	collector := metrics.NewCollector(cfg.Hosts)
	dashboard := NewDashboard(cfg.Hosts)

	app := tview.NewApplication()
	grid := tview.NewGrid().SetRows(0).SetColumns(0).SetBorders(true)

	updateGrid := func() {
		dashboard.UpdateMetrics(collector)
		app.QueueUpdateDraw(func() {
			grid.Clear()

			newPrimitive := func(title, text string) tview.Primitive {
				return tview.NewFrame(nil).
					SetBorders(0, 0, 0, 0, 0, 0).
					AddText(text, true, tview.AlignCenter, tcell.ColorWhite)
			}

			for i, host := range dashboard.Hosts {
				text := fmt.Sprintf("[yellow]CPU:[white] %.2f%%\n[yellow] Disk:[white] %.2f%%\n[yellow] Memory:[white] %.2f%%", host.CPU, host.Disk, host.Memory)
				hostBox := newPrimitive(host.Hostname, text)

				grid.AddItem(hostBox, i/2, i%2, 1, 1, 0, 0, false)
			}
		})
	}

	go func() {
		for {
			updateGrid()
			time.Sleep(2 * time.Second)
		}
	}()

	if err := app.SetRoot(grid, true).Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
	}
}
