package dashboard

import (
	"fmt"
	"strings"
	"time"

	help "github.com/charmbracelet/bubbles/help"
	key "github.com/charmbracelet/bubbles/key"
	progress "github.com/charmbracelet/bubbles/progress"
	spinner "github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	updateInterval    time.Duration
	keys              keyMap
	help              help.Model
	windowHeight      int
	collecting        bool
	debug             bool
}

type tickMsg time.Time

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
	loaded  bool
	isValid bool
	spinner spinner.Model
}

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

func NewModel(cfg *config.HostConfig, updateInterval time.Duration, debug bool) model {
	hosts := make([]hostBox, len(cfg.Hosts))
	for i, h := range cfg.Hosts {
		s := spinner.New()
		s.Spinner = spinner.Dot
		s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		hosts[i] = hostBox{
			name:    h.Name,
			cpu:     progress.New(progress.WithScaledGradient("#00ff00", "#ff0000")),
			disk:    progress.New(progress.WithScaledGradient("#00ff00", "#ff0000")),
			memory:  progress.New(progress.WithScaledGradient("#00ff00", "#ff0000")),
			spinner: s,
		}
	}
	return model{hosts: hosts, hostsCfg: cfg.Hosts, updateInterval: updateInterval, keys: keys, help: help.New(), debug: debug}
}

func (m model) Init() tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(m.hosts)+1)
	for i := range m.hosts {
		cmds = append(cmds, m.hosts[i].spinner.Tick)
	}
	cmds = append(cmds, tick())
	return tea.Batch(cmds...)
}

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func collectMetricsCmd(hostsCfg []config.Host) tea.Cmd {
	return func() tea.Msg {
		collector := metrics.NewCollector(hostsCfg)
		collected, err := collector.Collect()
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
	cmds := make([]tea.Cmd, len(m.hosts))
	for i := range m.hosts {
		host := &m.hosts[i]
		var cmd tea.Cmd
		host.spinner, cmd = host.spinner.Update(msg)
		cmds[i] = cmd
	}
	switch msg := msg.(type) {
	case tickMsg:
		now := time.Now()
		if !m.collecting && now.Sub(m.lastMetricsUpdate) >= m.updateInterval {
			m.collecting = true
			return m, tea.Batch(append([]tea.Cmd{collectMetricsCmd(m.hostsCfg), tick()}, cmds...)...)
		}
		return m, tea.Batch(append([]tea.Cmd{tick()}, cmds...)...)
	case metricsResultMsg:
		m.lastMetricsUpdate = time.Now()
		m.collecting = false
		for i := range m.hosts {
			host := &m.hosts[i]
			if metrics, ok := msg.collected[host.name]; ok && metrics.Error == "" {
				host.cpuVal = metrics.CPU
				host.diskVal = metrics.Disk
				host.memVal = metrics.Memory
				host.loadErr = metrics.Error
				host.loaded = true
				host.isValid = true
			} else {
				host.loadErr = metrics.Error
				host.loaded = true
				host.isValid = false
			}
		}
		if msg.err != nil {
			m.err = msg.err
		}
		return m, tea.Batch(cmds...)
	}
	return m, tea.Batch(cmds...)
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
		if host.loadErr != "" && !host.isValid && !m.debug {
			continue
		}
		out += fmt.Sprintf("[ %s ]\n", host.name)
		if !host.loaded {
			out += fmt.Sprintf("%s Connecting to hosts\n", host.spinner.View())
			continue
		}
		if host.loadErr != "" {
			out += fmt.Sprintf("\033[31m%s\033[0m\n", host.loadErr)
			continue
		}

		colorPercent := func(val float64) string {
			var color string
			switch {
			case val > 80:
				color = "\033[31m"
			case val > 50:
				color = "\033[33m"
			default:
				color = "\033[32m"
			}
			return fmt.Sprintf("%s%5.1f%%\033[0m", color, val)
		}

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
	helpView := m.help.View(m.keys)
	lines := 0
	for _, c := range out {
		if c == '\n' {
			lines++
		}
	}
	if m.windowHeight > 0 {
		pad := m.windowHeight - lines - len(helpView)/80 - 1
		if pad > 0 {
			out += strings.Repeat("\n", pad)
		}
	}
	out += helpView
	return out
}

func Start(cfg *config.HostConfig, updateInterval time.Duration, debug bool) {
	m := NewModel(cfg, updateInterval, debug)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		fmt.Printf("Error running Bubble Tea program: %v\n", err)
	}
}
