package main

import (
	"fmt"

	"github.com/gocolly/colly"
)

type table struct {
	Name   string
	Type   string
	Year   string
	Review string
}

func main() {
	c := colly.NewCollector()

	c.OnHTML("#band_disco a[href*='all']", func(e *colly.HTMLElement) {
		err := e.Request.Visit(e.Attr("href"))
		if err != nil {
			fmt.Println(err)
		}
	})

	c.OnHTML("table.display.discog tbody", func(h *colly.HTMLElement) {
		rows := make([]table, 0)
		h.ForEach("tr", func(i int, h *colly.HTMLElement) {
			row := table{}
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
		fmt.Println(rows)
	})

	err := c.Visit("https://www.metal-archives.com/bands/Panphage/")
	if err != nil {
		fmt.Println(err)
	}
}
