package rym

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/gocolly/colly"
)

const domain string = "https://rateyourmusic.com"

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
		fmt.Println(h.Text)
		band_link := h.ChildAttr("td:not(.page_search_img_cell) a.searchpage", "href")
		links = append(links, band_link)
		band_name := h.ChildText("td:not(.page_search_img_cell) a.searchpage")
		genres := make([]string, 0)
		h.ForEach("a.smallgreen", func(_ int, h *colly.HTMLElement) {
			genres = append(genres, h.Text)
		})
		country := h.ChildAttr("span.ui_flag", "title")
		rows = append(rows, table.Row{band_name, strings.Join(genres, "/"), country})
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r.StatusCode, "request was", r.Request.Headers, "\nError:", err)
	})

	c.Visit(fmt.Sprintf(domain+"/search?searchterm=%s&searchtype=a", strings.Replace(artist, " ", "%20", -1)))
	return rows, columns, links
}

func GetAlbumList(link string) ([]table.Row, []table.Column, []string) {
	c := colly.NewCollector()

	c.OnHTML("div#column_container_right div.section_artist_image > a > div", func(h *colly.HTMLElement) {
		fmt.Printf("Top Album: %s\n", h.Text)
	})
	c.OnHTML("div#column_container_right div.section_artist_biography > span.rendered_text", func(h *colly.HTMLElement) {
		fmt.Printf(h.Text)
	})
	rows := make([]table.Row, 0)
	columns := []table.Column{
		{Title: "Title", Width: 32},
		{Title: "Year", Width: 32},
		{Title: "Reviews", Width: 32},
		{Title: "Ratings", Width: 32},
		{Title: "Average", Width: 32},
	}
	links := []string{}
	c.OnHTML("div#column_container_left div#discography div.disco_release", func(h *colly.HTMLElement) {
		title := h.ChildText("div.disco_info a.album")
		year := h.ChildText("div.disco_info span[class*='disco_year']")
		reviews := h.ChildText("div.disco_reviews")
		ratings := h.ChildText("div.disco_ratings")
		average := h.ChildText("div.disco_avg_rating")
		rows = append(rows, table.Row{title, year, reviews, ratings, average})
		links = append(links, h.ChildAttr("div.disco_info > a", "href"))
	})

	c.Visit(domain + link)
	return rows, columns, links
}

func GetAlbum(link string) {
	c := colly.NewCollector()

	c.OnHTML("table.album_info > tbody > tr", func(h *colly.HTMLElement) {
		key := h.ChildText("th")
		value := strings.Join(strings.Fields(strings.Replace(h.ChildText("td"), "\n", "", -1)), " ")
		if key != "Share" {
			fmt.Printf("%s: %s\n", key, value)
		}
	})
	c.Visit(domain + link)
}
