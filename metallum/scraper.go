package metallum

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/bubbles/table"
	"github.com/gocolly/colly"
	"github.com/qeesung/image2ascii/convert"
)

type SearchResponse struct {
	Error                string     `json:"error"`
	ITotalRecords        int        `json:"iTotalRecords"`
	ITotalDisplayRecords int        `json:"iTotalDisplayRecords"`
	SEcho                int        `json:"sEcho"`
	AaData               [][]string `json:"aaData"`
}

func CreateRows(link string) ([]table.Row, []table.Column, []string) {
	c := colly.NewCollector()

	c.OnHTML("#band_disco a[href*='all']", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	rows := make([]table.Row, 0)
	columns := []table.Column{
		{Title: "Name", Width: 32},
		{Title: "Type", Width: 32},
		{Title: "Year", Width: 32},
		{Title: "Review", Width: 32},
	}
	album_links := make([]string, 0)
	c.OnHTML("table.display.discog tbody", func(h *colly.HTMLElement) {
		h.ForEach("tr", func(i int, h *colly.HTMLElement) {
			var row [4]string
			h.ForEach(".album,.demo,.other,td a[href]", func(i int, h *colly.HTMLElement) {
				row[i] = h.Text
				if i == 0 {
					album_links = append(album_links, h.Attr("href"))
				}
			})
			rows = append(rows, table.Row{row[0], row[1], row[2], row[3]})
		})
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.Visit(link)
	return rows, columns, album_links
}

func FindBand(band string) ([]table.Row, []table.Column, []string) {
	c := colly.NewCollector()

	rows := make([]table.Row, 0)
	links := make([]string, 0)
	columns := []table.Column{
		{Title: "Band Name", Width: 32},
		{Title: "Genre", Width: 32},
		{Title: "Country", Width: 32},
	}

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
			rows = append(rows, table.Row{row[0], row[1], row[2]})
		}
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.Visit(fmt.Sprintf("https://www.metal-archives.com/search/ajax-band-search/?field=name&query=%s", band))
	return rows, columns, links
}

func GetAlbum(album_link string) ([]table.Row, []table.Column) {
	c := colly.NewCollector()
	rows := make([]table.Row, 0)
	columns := []table.Column{
		{Title: "N.", Width: 32},
		{Title: "Title", Width: 32},
		{Title: "Duration", Width: 32},
		{Title: "Lyric", Width: 32},
	}
	c.OnHTML("table.display.table_lyrics tbody", func(h *colly.HTMLElement) {
		h.ForEach("tr.even,tr.odd", func(i int, h *colly.HTMLElement) {
			var row [4]string
			h.ForEach("td", func(i int, h *colly.HTMLElement) {
				row[i] = h.Text
			})
			rows = append(rows, table.Row{row[0], row[1], row[2], row[3]})
		})
	})

	c.OnHTML("a#cover.image", func(h *colly.HTMLElement) {
		image_src := h.ChildAttr("img", "src")
		h.Request.Visit(image_src)
	})

	c.OnResponse(func(r *colly.Response) {
		if r.Headers.Get("content-type") == "image/jpeg" {
			img, _, err := image.Decode(bytes.NewReader(r.Body))
			if err != nil {
				fmt.Println(err)
			}
			converter := convert.NewImageConverter()
			convertOptions := convert.DefaultOptions
			convertOptions.FixedWidth = 180
			convertOptions.FixedHeight = 40
			fmt.Print(converter.Image2ASCIIString(img, &convertOptions))
		}
	})
	metadata_keys := make([]string, 0)
	metadata_values := make([]string, 0)
	c.OnHTML("dl.float_right,dl.float_left", func(h *colly.HTMLElement) {
		h.ForEach("dt", func(_ int, h *colly.HTMLElement) {
			metadata_keys = append(metadata_keys, h.Text)
		})
		h.ForEach("dd", func(_ int, h *colly.HTMLElement) {
			metadata_values = append(metadata_values, h.Text)
		})
		for i, key := range metadata_keys {
			fmt.Printf("%s %s\n", key, metadata_values[i])
		}
	})

	c.Visit(album_link)
	return rows, columns
}
