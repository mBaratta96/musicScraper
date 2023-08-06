package scraper

import (
	"image"

	"github.com/wk8/go-ordered-map/v2"
)

type ColumnData struct {
	Title []string
	Width []int
}

type ScrapedData struct {
	Rows    [][]string
	Columns ColumnData
	// optionals
	Links    []string
	Metadata *orderedmap.OrderedMap[string, string]
	Image    image.Image
}

type Scraper interface {
	SearchBand(*ScrapedData) ([]int, []string)
	AlbumList(*ScrapedData) ([]int, []string)
	Album(*ScrapedData) ([]int, []string)
	ReviewsList(*ScrapedData) ([]int, []string)
	Credits() *orderedmap.OrderedMap[string, string]
	StyleColor() string
	SetLink(string)
	ListChoices() []string
	AdditionalFunctions() map[string]interface{}
}

var listMenuDefaultChoices = []string{"Go back", "Show credits", "Show reviews"}

type TableConstructor func(*ScrapedData) ([]int, []string)

func ScrapeData(method TableConstructor) ScrapedData {
	data := ScrapedData{}
	data.Rows = make([][]string, 0)
	colWidths, colTitles := method(&data)
	data.Columns = ColumnData{
		Title: colTitles[:],
		Width: computeColumnWidth(colWidths, colTitles, data.Rows),
	}
	return data
}

func computeColumnWidth(maxWidth []int, colTitles []string, rows [][]string) []int {
	widths := []int{}
	padding := 2
	for i, w := range maxWidth {
		maxLength := 0
		for _, row := range rows {
			cell := row[i]
			if len(cell) > maxLength {
				maxLength = len(cell)
			}
		}
		maxLength += padding
		switch {
		case maxLength > w:
			widths = append(widths, w)
		case maxLength < len(colTitles[i]):
			widths = append(widths, len(colTitles[i]))
		default:
			widths = append(widths, maxLength)
		}
	}
	return widths
}
