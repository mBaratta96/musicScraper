package scraper

import (
	"image"
	"reflect"
)

type ColumnData struct {
	Title []string
	Width []int
}

type ScrapedData struct {
	Rows     [][]string
	Columns  ColumnData
	Links    []string
	Metadata map[string]string
	Image    image.Image
}

type Scraper interface {
	FindBand(*ScrapedData) ([]int, []string)
	GetAlbumList(*ScrapedData) ([]int, []string)
	GetAlbum(*ScrapedData) ([]int, []string)
	GetStyleColor() string
}

type TableConstructor func(*ScrapedData) ([]int, []string)

func ScrapeData(method TableConstructor) ScrapedData {
	data := ScrapedData{
		Rows:     make([][]string, 0),
		Columns:  ColumnData{},
		Links:    make([]string, 0),
		Metadata: make(map[string]string),
		Image:    nil,
	}

	colWidths, colTitles := method(&data)
	data.Columns = createColumns(colWidths, colTitles, data.Rows)
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

func createColumns(widths []int, titles []string, rows [][]string) ColumnData {
	return ColumnData{
		Title: titles[:],
		Width: computeColumnWidth(widths, titles, rows),
	}
}

// func SetLink(s interface{}, link string) {
// 	stype := reflect.ValueOf(s).Elem()
// 	tmp := reflect.New(stype.Elem().Type()).Elem()
// 	tmp.Set(stype.Elem())
// 	tmp.FieldByName("Link").SetString(link)
// 	stype.Set(tmp)
// }

func SetLink(s interface{}, link string) {
	stype := reflect.ValueOf(s).Elem()
	stype.Elem().FieldByName("Link").Set(reflect.ValueOf(&link))
}
