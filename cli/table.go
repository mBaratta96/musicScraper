package cli

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type tableModel struct {
	table table.Model
	exit  bool
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func (m tableModel) Init() tea.Cmd { return nil }

func (m tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m tableModel) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func createColumns(columnNames []string, widths []int) []table.Column {
	columns := make([]table.Column, 0)
	screenWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))
	totalColumnWidth := 0
	for _, width := range widths {
		totalColumnWidth += width
	}
	maxScreenSize := screenWidth - 2*(len(columnNames)+1)
	for i, width := range widths {
		var w int
		if totalColumnWidth < maxScreenSize {
			w = width
		} else {
			w = width - int(math.Ceil(float64(totalColumnWidth-maxScreenSize)/float64(len(widths))))
		}
		columns = append(columns, table.Column{Title: columnNames[i], Width: w})
	}
	return columns
}

func createRows(rowsString [][]string) []table.Row {
	rows := make([]table.Row, 0)
	for _, row := range rowsString {
		for i, el := range row {
			row[i] = strings.TrimSpace(el)
		}
		rows = append(rows, row)
	}
	return rows
}

func PrintRows(rowsString [][]string, columnsString []string, widths []int) int {
	columns := createColumns(columnsString, widths)
	rows := createRows(rowsString)
	_, screenHeigth, _ := term.GetSize(int(os.Stdout.Fd()))
	var height int
	if screenHeigth/2 < len(rows) {
		height = screenHeigth / 2
	} else {
		height = len(rows)
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
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

	p := tea.NewProgram(tableModel{t, false})
	m, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	if m, ok := m.(tableModel); ok {
		if !m.exit {
			return m.table.Cursor()
		}
	} else {
		fmt.Println("Error in table")
		os.Exit(1)
	}
	return -1
}
