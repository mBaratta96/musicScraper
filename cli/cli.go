package cli

import (
	"fmt"
	"image"
	"os"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/qeesung/image2ascii/convert"
	"golang.org/x/term"
)

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

func PrintList() {
	createList()
}

func PrintReview(review string) {
	words := strings.FieldsFunc(review, func(r rune) bool {
		return unicode.IsSpace(r) && r != '\n'
	})
	screenWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))
	totalTextLength := 0
	for _, word := range words {
		switch {
		case strings.Contains(word, "\n"): // if it's word + \n + word
			fmt.Printf("%s ", word)
			totalTextLength = len(strings.Split(word, "\n")[1])
		case len(word)+1+totalTextLength > screenWidth:
			fmt.Printf("\n%s ", word)
			totalTextLength = len(word) + 1
		default:
			fmt.Printf("%s ", word)
			totalTextLength += len(word) + 1

		}
	}
	fmt.Print("\n")
}
