package cli

import (
	"fmt"
	"image"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/qeesung/image2ascii/convert"
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

func PrintRows(rowsString [][]string, columnsString []string, widths []int) int {
	columns := createColumns(columnsString, widths)
	rows := createRows(rowsString)
	height := 14
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

	p := tea.NewProgram(model{t, false})
	m, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	if m, ok := m.(model); ok {
		if !m.exit {
			return m.table.Cursor()
		}
	} else {
		fmt.Println("Error in table")
		os.Exit(1)
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

func PrintMetadata(metadata map[string]string, color string) {
	w, _, e := term.GetSize(0)
	if e != nil {
		panic(e)
	}
	max_key_length := 0
	for k := range metadata {
		if len(k) > max_key_length {
			max_key_length = len(k)
		}
	}
	max_key_length += 4
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	value_space := w - (max_key_length + 1)
	for key, value := range metadata {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(value) > value_space {
			var value2 string
			firstI := 0
			lastI := value_space
			part := lastI
			spacingS := strings.Repeat(" ", max_key_length-len(key))
			spacing := spacingS

			for i := 0; i < len(value); i += part {
				value2 += spacingS + value[firstI:lastI] + "\n" + spacing
				spacingS = strings.Repeat(" ", len(key)+1)
				if ((len(value) - lastI) < part) || ((len(value) - lastI) == part) {
					break
				}
				firstI = lastI
				lastI += part
			}
			value2 += spacingS + value[lastI:]
			fmt.Println(style.Render(key), value2)
		} else {
			fmt.Println(style.Render(key)+strings.Repeat(" ", max_key_length-len(key)), value)
		}
	}
}
