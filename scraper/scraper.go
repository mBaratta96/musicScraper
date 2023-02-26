package scraper

import (
	"fmt"

	"github.com/gocolly/colly"
)

type Table struct {
	Name   string
	Type   string
	Year   string
	Review string
}

func PrintRows() []Table {
	c := colly.NewCollector()

	c.OnHTML("#band_disco a[href*='all']", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	rows := make([]Table, 0)
	c.OnHTML("table.display.discog tbody", func(h *colly.HTMLElement) {
		h.ForEach("tr", func(i int, h *colly.HTMLElement) {
			row := Table{}
			h.ForEach(".album,.demo,.other,td a[href]", func(i int, h *colly.HTMLElement) {
				switch i {
				case 0:
					row.Name = h.Text
				case 1:
					row.Type = h.Text
				case 2:
					row.Year = h.Text
				case 3:
					row.Review = h.Text
				}
			})
			rows = append(rows, row)
		})
	})

	c.Visit("https://www.metal-archives.com/bands/Panphage/")
	return rows
}
