package dashboard

import (
	"fmt"
	"strings"
	"time"

	help "github.com/charmbracelet/bubbles/help"
	key "github.com/charmbracelet/bubbles/key"
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
	hosts             []hostBox
	hostsCfg          []config.Host
	err               error
	lastMetricsUpdate time.Time
	keys              keyMap
	help              help.Model
	windowHeight      int
}

type tickMsg time.Time

// Define a message to carry metrics results
type metricsResultMsg struct {
	collected map[string]HostMetrics
	err       error
}

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

// KeyMap for help
// keyMap defines a set of keybindings. To work for help it must satisfy key.Map.
type keyMap struct {
	Quit key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit}
}
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Quit}}
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
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
	return model{hosts: hosts, hostsCfg: cfg.Hosts, keys: keys, help: help.New()}
}

func (m model) Init() tea.Cmd {
	return tick()
}

func tick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Async command to collect metrics
func collectMetricsCmd(hostsCfg []config.Host) tea.Cmd {
	return func() tea.Msg {
		collector := metrics.NewCollector(hostsCfg)
		collected, err := collector.Collect()
		// Convert collected to map[string]HostMetrics if needed
		result := make(map[string]HostMetrics)
		for k, v := range collected {
			result[k] = HostMetrics{
				Hostname: k,
				CPU:      v.CPU,
				Disk:     v.Disk,
				Memory:   v.Memory,
				Error:    v.Error,
			}
		}
		return metricsResultMsg{collected: result, err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, m.keys.Quit) {
			return m, tea.Quit
		}
	}
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.windowHeight = ws.Height
	}
	switch msg := msg.(type) {
	case tickMsg:
		now := time.Now()
		if now.Sub(m.lastMetricsUpdate) >= 2*time.Second {
			m.lastMetricsUpdate = now
			return m, tea.Batch(collectMetricsCmd(m.hostsCfg), tick())
		}
		return m, tick()
	case metricsResultMsg:
		for i := range m.hosts {
			host := &m.hosts[i]
			if metrics, ok := msg.collected[host.name]; ok {
				host.cpuVal = metrics.CPU
				host.diskVal = metrics.Disk
				host.memVal = metrics.Memory
				host.loadErr = metrics.Error
			} else {
				host.loadErr = "No data"
			}
		}
		if msg.err != nil {
			m.err = msg.err
		}
		return m, nil
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
			out += fmt.Sprintf("\033[31m%s\033[0m\n", host.loadErr)
			continue
		}

		// Helper to color percentage text
		colorPercent := func(val float64) string {
			var color string
			switch {
			case val > 80:
				color = "\033[31m" // red
			case val > 50:
				color = "\033[33m" // yellow
			default:
				color = "\033[32m" // green
			}
			return fmt.Sprintf("%s%5.1f%%\033[0m", color, val)
		}

		// Always show full gradient for background, fill is always green
		barOpts := []progress.Option{
			progress.WithGradient("#00ff00", "#ff0000"),
			progress.WithoutPercentage(),
		}
		cpuBar := progress.New(barOpts...)
		diskBar := progress.New(barOpts...)
		memBar := progress.New(barOpts...)

		out += fmt.Sprintf("CPU:    %s %s\n", cpuBar.ViewAs(host.cpuVal/100), colorPercent(host.cpuVal))
		out += fmt.Sprintf("Disk:   %s %s\n", diskBar.ViewAs(host.diskVal/100), colorPercent(host.diskVal))
		out += fmt.Sprintf("Memory: %s %s\n", memBar.ViewAs(host.memVal/100), colorPercent(host.memVal))
	}
	// Pad with newlines to push help to the bottom
	helpView := m.help.View(m.keys)
	lines := 0
	for _, c := range out {
		if c == '\n' {
			lines++
		}
	}
	if m.windowHeight > 0 {
		pad := m.windowHeight - lines - len(helpView)/80 - 1 // crude line estimate
		if pad > 0 {
			out += strings.Repeat("\n", pad)
		}
	}
	out += helpView
	return out
}

func Start(cfg *config.HostConfig) {
	m := NewModel(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		fmt.Printf("Error running Bubble Tea program: %v\n", err)
	}
}
