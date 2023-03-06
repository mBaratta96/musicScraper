package scraper

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/bubbles/table"
	"github.com/gocolly/colly"
)

type SearchResponse struct {
	Error                string     `json:"error"`
	ITotalRecords        int        `json:"iTotalRecords"`
	ITotalDisplayRecords int        `json:"iTotalDisplayRecords"`
	SEcho                int        `json:"sEcho"`
	AaData               [][]string `json:"aaData"`
}

func PrintRows(link string) ([]table.Row, []table.Column) {
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
	c.OnHTML("table.display.discog tbody", func(h *colly.HTMLElement) {
		h.ForEach("tr", func(i int, h *colly.HTMLElement) {
			var row [4]string
			h.ForEach(".album,.demo,.other,td a[href]", func(i int, h *colly.HTMLElement) {
				row[i] = h.Text
			})
			rows = append(rows, table.Row{row[0], row[1], row[2], row[3]})
		})
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.Visit(link)
	return rows, columns
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
						fmt.Println("EROOR")
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
