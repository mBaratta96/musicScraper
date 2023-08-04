package cli

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/qeesung/image2ascii/convert"
	"github.com/wk8/go-ordered-map/v2"
	"golang.org/x/term"
)

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

func PrintImage(img image.Image) {
	converter := convert.NewImageConverter()
	convertOptions := convert.DefaultOptions
	fmt.Println(converter.Image2ASCIIString(img, &convertOptions))
}

func PrintLink(link string) {
	fmt.Println("\n" + link + "\n")
}

// use utf8 for non-English alphabets
func PrintMap(color string, metadata *orderedmap.OrderedMap[string, string]) {
	w, _, e := term.GetSize(0)
	if e != nil {
		panic(e)
	}
	maxKeyLength := 0
	for pair := metadata.Oldest(); pair != nil; pair = pair.Next() {
		k := pair.Key
		if utf8.RuneCountInString(k) > maxKeyLength {
			maxKeyLength = utf8.RuneCountInString(k)
		}
	}
	maxKeyLength += 4
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	valueSpace := w - maxKeyLength
	for pair := metadata.Oldest(); pair != nil; pair = pair.Next() {
		key := pair.Key
		value := pair.Value
		if len(value) > valueSpace {
			fmt.Printf("%s%s", style.Render(key), strings.Repeat(" ", maxKeyLength-utf8.RuneCountInString(key)))
			words := strings.Fields(value)
			totalTextLength := 0
			spacing := strings.Repeat(" ", maxKeyLength)
			for _, word := range words {
				if len(word)+1+totalTextLength > valueSpace {
					fmt.Printf("\n%s", spacing)
					totalTextLength = 0
				}
				fmt.Printf("%s ", word)
				totalTextLength += len(word) + 1
			}
			fmt.Print("\n")
		} else {
			fmt.Println(style.Render(key) + strings.Repeat(" ", maxKeyLength-utf8.RuneCountInString(key)) + value)
		}
	}
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
			totalTextLength = len(strings.Split(word, "\n")[1]) + 1
		case len(word)+1+totalTextLength > screenWidth:
			fmt.Print("\n")
			totalTextLength = len(word) + 1
		default:
			totalTextLength += len(word) + 1
		}
		fmt.Printf("%s ", word)
	}
	fmt.Print("\n")
}
