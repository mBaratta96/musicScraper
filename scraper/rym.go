package scraper

import (
	"bytes"
	"fmt"
	"image"
	"strings"

	_ "image/jpeg"

	"github.com/gocolly/colly"
)

const (
	DOMAIN        string = "https://rateyourmusic.com"
	RYMSTYLECOLOR string = "#427b58"
)

var (
	rBandColumnTitles      = [3]string{"Band Name", "Genre", "Country"}
	rBandColumnWidths      = [3]int{64, 64, 32}
	rAlbumlistColumnTitles = [7]string{"Rec.", "Title", "Year", "Reviews", "Rating", "Average", "Type"}
	rAlbumlistColumnWidths = [7]int{4, 64, 4, 7, 7, 7, 12}
	rAlbumColumnTitles     = [3]string{"N.", "Title", "Duration"}
	rAlbumColumnWidths     = [3]int{4, 64, 8}
)

func (r RateYourMusic) FindBand() ([][]string, ColumnData, []string) {
	c := colly.NewCollector()
	rows := make([][]string, 0)

	links := []string{}
	c.OnHTML("table tr.infobox", func(h *colly.HTMLElement) {
		band_link := DOMAIN + h.ChildAttr("td:not(.page_search_img_cell) a.searchpage", "href")
		links = append(links, band_link)
		band_name := h.ChildText("td:not(.page_search_img_cell) a.searchpage")
		genres := make([]string, 0)
		h.ForEach("a.smallgreen", func(_ int, h *colly.HTMLElement) {
			genres = append(genres, h.Text)
		})
		country := h.ChildAttr("span.ui_flag", "title")
		rows = append(rows, []string{band_name, strings.Join(genres, "/"), country})
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r.StatusCode, "request was", r.Request.Headers, "\nError:", err)
	})

	c.Visit(fmt.Sprintf(DOMAIN+"/search?searchterm=%s&searchtype=a", strings.Replace(r.Search, " ", "%20", -1)))

	columns := ColumnData{
		Title: rBandColumnTitles[:],
		Width: rBandColumnWidths[:],
	}
	return rows, columns, links
}

func addAlbums(h *colly.HTMLElement, query string, section string) ([][]string, []string) {
	links := []string{}
	rows := make([][]string, 0)
	h.ForEach(query, func(_ int, h *colly.HTMLElement) {
		title := h.ChildText("div.disco_info a.album")
		year := h.ChildText("div.disco_info span[class*='disco_year']")
		reviews := h.ChildText("div.disco_reviews")
		ratings := h.ChildText("div.disco_ratings")
		average := h.ChildText("div.disco_avg_rating")
		recommended := ""
		if h.ChildAttr("div.disco_info b.disco_mainline_recommended", "title") == "Recommended" {
			recommended = ""
		}
		rows = append(rows, []string{recommended, title, year, reviews, ratings, average, section})
		links = append(links, DOMAIN+h.ChildAttr("div.disco_info > a", "href"))
	})
	return rows, links
}

type AlbumTable struct {
	Query   string
	Section string
}

func (r RateYourMusic) GetAlbumList(link string) ([][]string, ColumnData, []string, map[string]string) {
	c := colly.NewCollector()

	metadata := make(map[string]string)
	c.OnHTML("div#column_container_right div.section_artist_image > a > div", func(h *colly.HTMLElement) {
		metadata["Top Album"] = h.Text
	})
	c.OnHTML("div#column_container_right div.section_artist_biography > span.rendered_text", func(h *colly.HTMLElement) {
		metadata["Biography"] = h.Text
	})

	rows := make([][]string, 0)
	links := []string{}
	albumTables := []AlbumTable{
		{Query: "div#disco_type_s > div.disco_release", Section: "Album"},
		{Query: "div#disco_type_l > div.disco_release", Section: "Live Album"},
		{Query: "div#disco_type_e > div.disco_release", Section: "EP"},
		{Query: "div#disco_type_a > div.disco_release", Section: "Appears On"},
		{Query: "div#disco_type_c > div.disco_release", Section: "Compilation"},
	}
	c.OnHTML("div#column_container_left div#discography", func(h *colly.HTMLElement) {
		for _, albumTable := range albumTables {
			album_rows, album_links := addAlbums(h, albumTable.Query, albumTable.Section)
			rows = append(rows, album_rows...)
			links = append(links, album_links...)
		}
	})

	c.Visit(link)
	columns := ColumnData{
		Title: rAlbumlistColumnTitles[:],
		Width: rAlbumlistColumnWidths[:],
	}
	return rows, columns, links, metadata
}

func (r RateYourMusic) GetAlbum(link string) ([][]string, ColumnData, map[string]string, image.Image) {
	c := colly.NewCollector()

	c.OnHTML("div#column_container_left div.page_release_art_frame", func(h *colly.HTMLElement) {
		image_url := h.ChildAttr("img", "src")
		h.Request.Visit("https:" + image_url)
	})

	var img image.Image
	c.OnResponse(func(r *colly.Response) {
		if r.Headers.Get("content-type") == "image/jpg" {
			var err error
			img, _, err = image.Decode(bytes.NewReader(r.Body))
			if err != nil {
				fmt.Println(err)
			}
		}
	})

	metadata := make(map[string]string)
	c.OnHTML("table.album_info > tbody > tr", func(h *colly.HTMLElement) {
		key := h.ChildText("th")
		value := strings.Join(strings.Fields(strings.Replace(h.ChildText("td"), "\n", "", -1)), " ")
		if key != "Share" {
			metadata[key] = value
		}
	})
	rows := make([][]string, 0)

	c.OnHTML("div#column_container_left div.section_tracklisting ul#tracks", func(h *colly.HTMLElement) {
		h.ForEach("li.track", func(_ int, h *colly.HTMLElement) {
			if len(h.ChildText("span.tracklist_total")) > 0 {
				value := strings.Fields(h.ChildText("span.tracklist_total"))
				metadata["Total Length"] = value[len(value)-1]
			} else {
				number := h.ChildText("span.tracklist_num")
				title := h.ChildText("span[itemprop=name] span.rendered_text")
				duration := h.ChildText("span.tracklist_duration")
				rows = append(rows, []string{number, title, duration})
			}
		})
	})
	c.Visit(link)
	columns := ColumnData{
		Title: rAlbumColumnTitles[:],
		Width: rAlbumColumnWidths[:],
	}
	return rows, columns, metadata, img
}

func (r RateYourMusic) GetStyleColor() string {
	return RYMSTYLECOLOR
}