package cli

import (
	"math"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type model struct {
	table table.Model
	exit  bool
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func CallClear() {
	clear := make(map[string]func())
	clear["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			panic("Error in Linux clear terminal")
		}
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			panic("Error in Windows clear terminal")
		}
	}
	value, ok := clear[runtime.GOOS]
	if ok {
		value()
	}
}

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
