package dashboard

import (
	"fmt"
	"time"

	progress "github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"

	"monclissh/internal/config"
	"monclissh/internal/metrics"
)

type HostMetrics struct {
	Hostname string
	CPU      float64
	Disk     float64
	Memory   float64
	Error    string
}

type Dashboard struct {
	Hosts []HostMetrics
}

type model struct {
	hosts []hostBox
	err   error
}

type tickMsg time.Time

type metricsMsg struct{}

type hostBox struct {
	name    string
	cpu     progress.Model
	disk    progress.Model
	memory  progress.Model
	cpuVal  float64
	diskVal float64
	memVal  float64
	loadErr string
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
		return
	}
	for i := range d.Hosts {
		host := d.Hosts[i]
		if metrics, ok := collectedMetrics[host.Hostname]; ok {
			d.Hosts[i].CPU = metrics.CPU
			d.Hosts[i].Disk = metrics.Disk
			d.Hosts[i].Memory = metrics.Memory
			d.Hosts[i].Error = metrics.Error
		}
	}
}

func NewModel(cfg *config.HostConfig) model {
	hosts := make([]hostBox, len(cfg.Hosts))
	for i, h := range cfg.Hosts {
		hosts[i] = hostBox{
			name:   h.Name,
			cpu:    progress.New(progress.WithScaledGradient("#00ff00", "#ff0000")),
			disk:   progress.New(progress.WithScaledGradient("#00ff00", "#ff0000")),
			memory: progress.New(progress.WithScaledGradient("#00ff00", "#ff0000")),
		}
	}
	return model{hosts: hosts}
}

func (m model) Init() tea.Cmd {
	return tick()
}

func tick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		return m, tea.Batch(tea.Cmd(func() tea.Msg { return metricsMsg{} }), tick())
	case metricsMsg:
		// This will be set in Start()
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}
	if len(m.hosts) == 0 {
		return "No hosts configured.\n"
	}
	var out string
	for i, host := range m.hosts {
		if i > 0 {
			out += "\n"
		}
		out += fmt.Sprintf("[ %s ]\n", host.name)
		if host.loadErr != "" {
			out += fmt.Sprintf("[red]Error: %s[white]\n", host.loadErr)
			continue
		}
		out += fmt.Sprintf("CPU:    %s %5.1f%%\n", host.cpu.ViewAs(host.cpuVal/100), host.cpuVal)
		out += fmt.Sprintf("Disk:   %s %5.1f%%\n", host.disk.ViewAs(host.diskVal/100), host.diskVal)
		out += fmt.Sprintf("Memory: %s %5.1f%%\n", host.memory.ViewAs(host.memVal/100), host.memVal)
	}
	out += "\nPress q to quit."
	return out
}

func Start(cfg *config.HostConfig) {
	collector := metrics.NewCollector(cfg.Hosts)
	m := NewModel(cfg)

	go func() {
		for {
			time.Sleep(200 * time.Millisecond)
			collectedMetrics, err := collector.Collect()
			for i := range m.hosts {
				host := &m.hosts[i]
				if metrics, ok := collectedMetrics[host.name]; ok {
					host.cpuVal = metrics.CPU
					host.diskVal = metrics.Disk
					host.memVal = metrics.Memory
					host.loadErr = metrics.Error
				} else {
					host.loadErr = "No data"
				}
			}
			if err != nil {
				m.err = err
			}
		}
	}()

	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Printf("Error running Bubble Tea program: %v\n", err)
	}
}
