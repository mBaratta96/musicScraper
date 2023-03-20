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
	"github.com/charmbracelet/lipgloss"
	"github.com/gocolly/colly"
)

type SearchResponse struct {
	Error                string     `json:"error"`
	ITotalRecords        int        `json:"iTotalRecords"`
	ITotalDisplayRecords int        `json:"iTotalDisplayRecords"`
	SEcho                int        `json:"sEcho"`
	AaData               [][]string `json:"aaData"`
}

type Metallum struct {
	Search string
}

func getMetadata(h *colly.HTMLElement) map[string]string {
	key_style := lipgloss.NewStyle().Width(32).Foreground(lipgloss.Color("#b57614"))
	keys, values := []string{}, []string{}
	h.ForEach("dt", func(_ int, h *colly.HTMLElement) {
		keys = append(keys, key_style.Render(h.Text))
	})
	h.ForEach("dd", func(_ int, h *colly.HTMLElement) {
		values = append(values, strings.Replace(h.Text, "\n", "", -1))
	})
	metadata := map[string]string{}
	for i, k := range keys {
		metadata[k] = values[i]
	}
	return metadata
}

func (m *Metallum) GetAlbumList(link string) ([][]string, []string, map[string]string) {
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
	var metadata map[string]string
	c.OnHTML("dl.float_right,dl.float_left", func(h *colly.HTMLElement) {
		metadata = getMetadata(h)
	})
	c.Visit(link)
	return rows, album_links, metadata
}

func (m *Metallum) FindBand() ([][]string, []string) {
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

	c.Visit(fmt.Sprintf("https://www.metal-archives.com/search/ajax-band-search/?field=name&query=%s", m.Search))
	return rows, links
}

func (m *Metallum) GetAlbum(album_link string) ([][]string, map[string]string, image.Image) {
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
	var metadata map[string]string
	c.OnHTML("dl.float_right,dl.float_left", func(h *colly.HTMLElement) {
		metadata = getMetadata(h)
	})
	c.Visit(album_link)
	return rows, metadata, img
}
