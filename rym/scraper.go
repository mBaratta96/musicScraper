package rym

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/gocolly/colly"
)

func SearchArtist(artist string) ([]table.Row, []table.Column, []string) {
	c := colly.NewCollector()
	rows := make([]table.Row, 0)
	columns := []table.Column{
		{Title: "Band name", Width: 32},
		{Title: "Genere", Width: 32},
		{Title: "Country", Width: 32},
	}
	links := []string{}
	c.OnHTML("table tr.infobox", func(h *colly.HTMLElement) {
		band_link := h.ChildAttr("td:not(.page_search_img_cell) a.searchpage", "href")
		links = append(links, band_link)
		band_name := h.ChildText("td:not(.page_search_img_cell) a.searchpage")
		genres := []string{}
		h.ForEach("a.smallgreen", func(_ int, h *colly.HTMLElement) {
			genres = append(genres, h.Text)
		})
		fmt.Println(strings.Join(genres, "/"))
		country := h.ChildAttr("span.ui_flag", "title")
		rows = append(rows, table.Row{band_name, strings.Join(genres, "/"), country})
	})

	c.Visit(fmt.Sprintf("https://rateyourmusic.com/search?searchterm=%s&searchtype=a", artist))
	return rows, columns, links
}
