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
	Search string
}
