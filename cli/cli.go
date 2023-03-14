package cli

import (
	"fmt"
	"image"
	"os"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/qeesung/image2ascii/convert"
)

type model struct {
	table table.Model
	exit  bool
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
			return m, tea.Quit
		case "q", "ctrl+c":
			m.exit = true
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func PrintRows(rows []table.Row, columns []table.Column) int {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	p := tea.NewProgram(model{t, false})
	m, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	if m, ok := m.(model); ok {
		if m.exit == true {
			os.Exit(1)
		}
		return m.table.Cursor()
	}
	return -1
}

func PrintImage(img image.Image) {
	converter := convert.NewImageConverter()
	convertOptions := convert.DefaultOptions
	fmt.Println(converter.Image2ASCIIString(img, &convertOptions))
}

func PrintLink(link string) {
	fmt.Println("\n" + link + "\n")
}

func PrintMetadata(keys []string, values []string) {
	for i, key := range keys {
		fmt.Println(key + values[i])
	}
}
