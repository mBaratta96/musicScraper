package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type SearchResponse struct {
	Error                string     `json:"error"`
	ITotalRecords        int        `json:"iTotalRecords"`
	ITotalDisplayRecords int        `json:"iTotalDisplayRecords"`
	SEcho                int        `json:"sEcho"`
	AaData               [][]string `json:"aaData"`
}

const METALLUMSTYLECOLOR string = "#b57614"

var (
	mBandColumnTitles      = [3]string{"Band Name", "Genre", "Country"}
	mBandColumnWidths      = [3]int{64, 64, 32}
	mAlbumlistColumnTitles = [4]string{"Name", "Type", "Year", "Country"}
	mAlbumlistColumnWidths = [4]int{64, 16, 4, 8}
	mAlbumColumnTitles     = [4]string{"N.", "Title", "Duration", "Lyric"}
	mAlbumColumnWidths     = [4]int{4, 64, 8, 16}
)

func getMetadata(h *colly.HTMLElement, m map[string]string) {
	keys, values := []string{}, []string{}
	h.ForEach("dt", func(_ int, h *colly.HTMLElement) {
		keys = append(keys, h.Text)
	})
	h.ForEach("dd", func(_ int, h *colly.HTMLElement) {
		values = append(values, strings.Replace(h.Text, "\n", "", -1))
	})
	// metadata := make(map[string]string)
	for i, k := range keys {
		// metadata[k] = values[i]
		m[k] = values[i]
	}
	return
}

func (m Metallum) FindBand() ([][]string, ColumnData, []string) {
	c := colly.NewCollector()

	rows := make([][]string, 0)
	links := make([]string, 0)

	c.OnResponse(func(r *colly.Response) {
		var response SearchResponse
		if err := json.Unmarshal(r.Body, &response); err != nil {
			fmt.Println("Can not unmarshal JSON")
		}

		for _, el := range response.AaData {
			var row [3]string
			for i, node := range el {
				switch i {
				case 0:
					doc, err := goquery.NewDocumentFromReader(strings.NewReader(node))
					if err != nil {
						fmt.Println("Error on response")
					}
					band := doc.Find("a").First()
					row[0] = band.Text()
					links = append(links, band.AttrOr("href", ""))
				case 1:
					row[1] = node
				case 2:
					row[2] = node
				}
			}
			rows = append(rows, []string{row[0], row[1], row[2]})
		}
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	columns := ColumnData{
		Title: mBandColumnTitles[:],
		Width: mBandColumnWidths[:],
	}
	c.Visit(fmt.Sprintf("https://www.metal-archives.com/search/ajax-band-search/?field=name&query=%s", m.Search))
	return rows, columns, links
}

func (m Metallum) GetAlbumList(link string) ([][]string, ColumnData, []string, map[string]string) {
	c := colly.NewCollector()

	c.OnHTML("#band_disco a[href*='all']", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	rows := make([][]string, 0)
	album_links := make([]string, 0)
	c.OnHTML("table.display.discog tbody tr", func(h *colly.HTMLElement) {
		var row [4]string
		h.ForEach(".album,.demo,.other,td a[href]", func(i int, h *colly.HTMLElement) {
			row[i] = h.Text
			if i == 0 {
				album_links = append(album_links, h.Attr("href"))
			}
		})
		rows = append(rows, []string{row[0], row[1], row[2], row[3]})
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})
	metadata := make(map[string]string)
	c.OnHTML("dl.float_right,dl.float_left", func(h *colly.HTMLElement) {
		getMetadata(h, metadata)
	})
	c.Visit(link)
	columns := ColumnData{
		Title: mAlbumlistColumnTitles[:],
		Width: mAlbumlistColumnWidths[:],
	}
	return rows, columns, album_links, metadata
}

func (m Metallum) GetAlbum(album_link string) ([][]string, ColumnData, map[string]string, image.Image) {
	c := colly.NewCollector()
	rows := make([][]string, 0)

	c.OnHTML("div#album_tabs_tracklist tr.even, div#album_tabs_tracklist tr.odd", func(h *colly.HTMLElement) {
		var row [4]string
		h.ForEach("td", func(i int, h *colly.HTMLElement) {
			row[i] = h.Text
		})
		rows = append(rows, []string{row[0], row[1], row[2], row[3]})
	})
	var img image.Image

	c.OnHTML("a#cover.image", func(h *colly.HTMLElement) {
		image_src := h.ChildAttr("img", "src")
		h.Request.Visit(image_src)
	})

	c.OnResponse(func(r *colly.Response) {
		if r.Headers.Get("content-type") == "image/jpeg" {
			var err error
			img, _, err = image.Decode(bytes.NewReader(r.Body))
			if err != nil {
				fmt.Println(err)
			}
		}
	})

	metadata := make(map[string]string)
	c.OnHTML("dl.float_right,dl.float_left", func(h *colly.HTMLElement) {
		getMetadata(h, metadata)
	})
	c.Visit(album_link)
	columns := ColumnData{
		Title: mAlbumColumnTitles[:],
		Width: mAlbumColumnWidths[:],
	}
	return rows, columns, metadata, img
}

func (m Metallum) GetStyleColor() string {
	return METALLUMSTYLECOLOR
}
