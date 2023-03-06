package scraper

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/gocolly/colly"
)

type Table struct {
	Name   string
	Type   string
	Year   string
	Review string
}

type SearchResponse struct {
	Error                string     `json:"error"`
	ITotalRecords        int        `json:"iTotalRecords"`
	ITotalDisplayRecords int        `json:"iTotalDisplayRecords"`
	SEcho                int        `json:"sEcho"`
	AaData               [][]string `json:"aaData"`
}

func PrintRows() ([]table.Row, []table.Column) {
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

	c.Visit("https://www.metal-archives.com/bands/Panphage/")
	return rows, columns
}

func FindBand(band string) {
	c := colly.NewCollector()

	c.OnResponse(func(r *colly.Response) {
		fmt.Println(string(r.Body))
		var response SearchResponse
		if err := json.Unmarshal(r.Body, &response); err != nil {
			fmt.Println("Can not unmarshal JSON")
		}
		fmt.Println(response.AaData)
	})

	c.Visit(fmt.Sprintf("https://www.metal-archives.com/search/ajax-band-search/?field=name&query=%s", band))
}
