package scraper

import (
	"image"
)

type ColumnData struct {
	Title []string
	Width []int
}

type Scraper interface {
	FindBand() ([][]string, ColumnData, []string)
	GetAlbumList(string) ([][]string, ColumnData, []string, map[string]string)
	GetAlbum(string) ([][]string, ColumnData, map[string]string, image.Image)
	GetStyleColor() string
}

type Metallum struct {
	Search string
}

type RateYourMusic struct {
	Search  string
	Credits bool
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
