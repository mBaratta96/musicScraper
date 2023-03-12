package rym

import (
	"bytes"
	"fmt"
	"image"
	"strings"

	_ "image/jpeg"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/gocolly/colly"
	"github.com/qeesung/image2ascii/convert"
)

const (
	domain string = "https://rateyourmusic.com"
)

func Login(user string, password string) {
	c := colly.NewCollector(colly.AllowURLRevisit(), colly.MaxDepth(5))
	fmt.Println(user)
	fmt.Println(password)
	data := map[string][]byte{
		"user":             []byte(user),
		"password":         []byte(password),
		"remember":         []byte("true"),
		"maintain_session": []byte("true"),
		"action":           []byte("Login"),
		"rym_ajax_req":     []byte("1"),
	}
	// data_rating := map[string][]byte{
	// 	"type":         []byte("l"),
	// 	"assoc_id":     []byte("15114102"),
	// 	"rating":       []byte("8"),
	// 	"action":       []byte("CatalogSetRating"),
	// 	"rym_ajax_req": []byte("1"),
	// }
	c.OnResponse(func(r *colly.Response) {
		fmt.Println(r.StatusCode)
		fmt.Println(r.Headers)
		fmt.Println(c.Cookies(r.Request.URL.String()))
	})
	c.PostMultipart("https://rateyourmusic.com/httprequest/Login", data)
}

func SearchArtist(artist string) ([]table.Row, []table.Column, []string) {
	c := colly.NewCollector()
	rows := make([]table.Row, 0)
	columns := []table.Column{
		{Title: "Band name", Width: 64},
		{Title: "Genere", Width: 64},
		{Title: "Country", Width: 32},
	}
	links := []string{}
	c.OnHTML("table tr.infobox", func(h *colly.HTMLElement) {
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

func addAlbums(h *colly.HTMLElement, query string, section string) ([]table.Row, []string) {
	links := []string{}
	rows := make([]table.Row, 0)
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
		rows = append(rows, table.Row{recommended, title, year, reviews, ratings, average, section})
		links = append(links, h.ChildAttr("div.disco_info > a", "href"))
	})
	return rows, links
}

type AlbumTable struct {
	Query   string
	Section string
}

func GetAlbumList(link string) ([]table.Row, []table.Column, []string) {
	c := colly.NewCollector()

	keyStyle := lipgloss.NewStyle().Width(32).Foreground(lipgloss.Color("#427b58"))
	c.OnHTML("div#column_container_right div.section_artist_image > a > div", func(h *colly.HTMLElement) {
		fmt.Println(keyStyle.Render("Top Album") + h.Text)
	})
	c.OnHTML("div#column_container_right div.section_artist_biography > span.rendered_text", func(h *colly.HTMLElement) {
		fmt.Println(keyStyle.Render("Biography") + h.Text)
	})
	rows := make([]table.Row, 0)
	columns := []table.Column{
		{Title: "Rec.", Width: 4},
		{Title: "Title", Width: 64},
		{Title: "Year", Width: 4},
		{Title: "Reviews", Width: 7},
		{Title: "Ratings", Width: 7},
		{Title: "Average", Width: 7},
		{Title: "Type", Width: 12},
	}
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

	c.Visit(domain + link)
	return rows, columns, links
}

func GetAlbum(link string) ([]table.Row, []table.Column) {
	c := colly.NewCollector()

	c.OnHTML("div#column_container_left div.page_release_art_frame", func(h *colly.HTMLElement) {
		image_url := h.ChildAttr("img", "src")
		h.Request.Visit("https:" + image_url)
	})

	c.OnResponse(func(r *colly.Response) {
		if r.Headers.Get("content-type") == "image/jpg" {
			img, _, err := image.Decode(bytes.NewReader(r.Body))
			if err != nil {
				fmt.Println(err)
			}
			converter := convert.NewImageConverter()
			convertOptions := convert.DefaultOptions
			fmt.Print(converter.Image2ASCIIString(img, &convertOptions))
			fmt.Println(domain + link)
		}
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r.StatusCode, "request was", r.Request.Headers, "\nError:", err)
	})

	keyStyle := lipgloss.NewStyle().Width(32).Foreground(lipgloss.Color("#427b58"))
	c.OnHTML("table.album_info > tbody > tr", func(h *colly.HTMLElement) {
		key := h.ChildText("th")
		value := strings.Join(strings.Fields(strings.Replace(h.ChildText("td"), "\n", "", -1)), " ")
		if key != "Share" {
			fmt.Println(keyStyle.Render(key) + value)
		}
	})
	rows := make([]table.Row, 0)
	columns := []table.Column{
		{Title: "N.", Width: 2},
		{Title: "Title", Width: 64},
		{Title: "Duration", Width: 8},
	}
	c.OnHTML("div#column_container_left div.section_tracklisting ul#tracks", func(h *colly.HTMLElement) {
		h.ForEach("li.track", func(_ int, h *colly.HTMLElement) {
			if len(h.ChildText("span.tracklist_total")) > 0 {
				key := keyStyle.Render("Total length")
				value := strings.Fields(h.ChildText("span.tracklist_total"))
				fmt.Println(key + value[len(value)-1])
			} else {
				number := h.ChildText("span.tracklist_num")
				title := h.ChildText("span[itemprop=name] span.rendered_text")
				duration := h.ChildText("span.tracklist_duration")
				rows = append(rows, table.Row{number, title, duration})
			}
		})
	})
	c.Visit(domain + link)
	return rows, columns
}
