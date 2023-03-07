package rym

import (
	"fmt"

	"github.com/gocolly/colly"
)

func SearchArtist(artist string) {
	c := colly.NewCollector()

	c.OnHTML("table tr.infobox", func(h *colly.HTMLElement) {
		fmt.Println(h.ChildAttr("td:not(.page_search_img_cell) a.searchpage", "href"))
	})

	c.Visit(fmt.Sprintf("https://rateyourmusic.com/search?searchterm=%s&searchtype=a", artist))
}
