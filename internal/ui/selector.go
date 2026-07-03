package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/deepawasthi/devstack/internal/engine"
	"github.com/deepawasthi/devstack/internal/services"
)

type SelectorResult struct {
	Services []string
	Name     string
}

type selectorModel struct {
	catalog  services.Catalog
	status   engine.RuntimeStatus
	items    []services.Service
	filtered []services.Service
	cursor   int
	selected map[string]bool
	search   textinput.Model
	mode     string
	name     textinput.Model
	done     bool
}

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).MarginBottom(1)
	okStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	warnStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
)

func RunSelector(catalog services.Catalog, status engine.RuntimeStatus) (SelectorResult, error) {
	search := textinput.New()
	search.Placeholder = "search services"
	search.Prompt = "/ "
	name := textinput.New()
	name.Placeholder = "environment name"
	name.SetValue("dev")
	name.Focus()
	model := selectorModel{
		catalog:  catalog,
		status:   status,
		items:    catalog.All(),
		filtered: catalog.All(),
		selected: map[string]bool{},
		search:   search,
		mode:     "select",
		name:     name,
	}
	program := tea.NewProgram(model)
	final, err := program.Run()
	if err != nil {
		return SelectorResult{}, err
	}
	m := final.(selectorModel)
	var ids []string
	for _, service := range m.items {
		if m.selected[service.ID] {
			ids = append(ids, service.ID)
		}
	}
	return SelectorResult{Services: ids, Name: strings.TrimSpace(m.name.Value())}, nil
}

func (m selectorModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case "name":
			switch msg.String() {
			case "enter":
				m.done = true
				return m, tea.Quit
			case "esc":
				m.mode = "select"
				m.search.Blur()
				return m, nil
			}
			m.name, cmd = m.name.Update(msg)
			return m, cmd
		case "search":
			switch msg.String() {
			case "enter", "esc":
				m.mode = "select"
				m.search.Blur()
				m.applySearch()
				return m, nil
			}
			m.search, cmd = m.search.Update(msg)
			m.applySearch()
			return m, cmd
		default:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.filtered)-1 {
					m.cursor++
				}
			case " ":
				if len(m.filtered) > 0 {
					id := m.filtered[m.cursor].ID
					m.selected[id] = !m.selected[id]
				}
			case "/":
				m.mode = "search"
				m.search.Focus()
				return m, textinput.Blink
			case "enter":
				m.mode = "name"
				m.name.Focus()
				return m, textinput.Blink
			}
		}
	}
	return m, cmd
}

func (m *selectorModel) applySearch() {
	query := strings.TrimSpace(m.search.Value())
	m.filtered = m.catalog.Search(query)
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func (m selectorModel) View() string {
	if m.done {
		return ""
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render("DevStack") + "\n")
	b.WriteString("Docker Environment Manager\n\n")
	b.WriteString(check("Docker Installed", m.status.DockerInstalled))
	b.WriteString(check("Docker Engine Running", m.status.EngineRunning))
	b.WriteString(check("Docker Compose Installed", m.status.ComposePresent))
	if m.status.PodmanPresent {
		b.WriteString(check("Podman Detected (future)", true))
	}
	b.WriteString("\n")
	if m.mode == "name" {
		b.WriteString("Environment name\n")
		b.WriteString(m.name.View() + "\n\n")
		b.WriteString(helpStyle.Render("ENTER = Create  ESC = Back") + "\n")
		return b.String()
	}
	if m.mode == "search" {
		b.WriteString(m.search.View() + "\n\n")
	}
	start, end := window(m.cursor, len(m.filtered), 14)
	for i := start; i < end; i++ {
		service := m.filtered[i]
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		box := "[ ]"
		if m.selected[service.ID] {
			box = "[x]"
		}
		line := fmt.Sprintf("%s%s %-24s %s", cursor, box, service.Name, service.Category)
		if i == m.cursor {
			line = selectedStyle.Render(line)
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("SPACE = Select  / = Search  ENTER = Continue  q = Quit") + "\n")
	return b.String()
}

func check(label string, ok bool) string {
	if ok {
		return okStyle.Render("✓ "+label) + "\n"
	}
	return warnStyle.Render("✗ "+label) + "\n"
}

func window(cursor, total, size int) (int, int) {
	if total <= size {
		return 0, total
	}
	start := cursor - size/2
	if start < 0 {
		start = 0
	}
	if start+size > total {
		start = total - size
	}
	return start, start + size
}
