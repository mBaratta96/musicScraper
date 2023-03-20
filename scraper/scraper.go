package scraper

import "image"

type Scraper interface {
	FindBand() ([][]string, []string)
	GetAlbumList(string) ([][]string, []string, map[string]string)
	GetAlbum(string) ([][]string, map[string]string, image.Image)
}
