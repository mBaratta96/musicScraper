
package main

import "fmt"
import "github.com/gocolly/colly"
import "strings"

type table struct {
    Name string
    Type string
    Year string
    Review string
}

func main() {
    c := colly.NewCollector()

	// Find and visit all links
	c.OnHTML("#band_disco a[href]", func(e *colly.HTMLElement) {
        if strings.Contains(e.Attr("href"), "all") {
            e.Request.Visit(e.Attr("href"))
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

	c.Visit("https://www.metal-archives.com/bands/Panphage/")
}
