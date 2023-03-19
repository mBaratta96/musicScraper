package metallum

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/gocolly/colly"
)

type SearchResponse struct {
	Error                string     `json:"error"`
	ITotalRecords        int        `json:"iTotalRecords"`
	ITotalDisplayRecords int        `json:"iTotalDisplayRecords"`
	SEcho                int        `json:"sEcho"`
	AaData               [][]string `json:"aaData"`
}

type Scraper interface {
	FindBand() ([]table.Row, []table.Column, []string)
}

type Metallum struct {
	Search string
}

type RateYourMusic struct {
	Search string
}

func getMetadata(h *colly.HTMLElement) ([]string, []string) {
	key_style := lipgloss.NewStyle().Width(32).Foreground(lipgloss.Color("#b57614"))
	keys, values := []string{}, []string{}
	h.ForEach("dt", func(_ int, h *colly.HTMLElement) {
		keys = append(keys, key_style.Render(h.Text))
	})
	h.ForEach("dd", func(_ int, h *colly.HTMLElement) {
		values = append(values, strings.Replace(h.Text, "\n", "", -1))
	})
	return keys, values
}

func CreateRows(link string) ([]table.Row, []table.Column, []string, []string, []string) {
	c := colly.NewCollector()

	c.OnHTML("#band_disco a[href*='all']", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	rows := make([]table.Row, 0)
	columns := []table.Column{
		{Title: "Name", Width: 64},
		{Title: "Type", Width: 16},
		{Title: "Year", Width: 4},
		{Title: "Review", Width: 8},
	}
	album_links := make([]string, 0)
	c.OnHTML("table.display.discog tbody tr", func(h *colly.HTMLElement) {
		var row [4]string
		h.ForEach(".album,.demo,.other,td a[href]", func(i int, h *colly.HTMLElement) {
			row[i] = h.Text
			if i == 0 {
				album_links = append(album_links, h.Attr("href"))
			}
		})
		rows = append(rows, table.Row{row[0], row[1], row[2], row[3]})
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})
	keys, values := []string{}, []string{}
	c.OnHTML("dl.float_right,dl.float_left", func(h *colly.HTMLElement) {
		keys, values = getMetadata(h)
	})
	c.Visit(link)
	return rows, columns, keys, values, album_links
}

func (m *Metallum) FindBand() ([]table.Row, []table.Column, []string) {
	c := colly.NewCollector()

	rows := make([]table.Row, 0)
	links := make([]string, 0)
	columns := []table.Column{
		{Title: "Band Name", Width: 64},
		{Title: "Genre", Width: 64},
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
						fmt.Println("Error on response")
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

	c.Visit(fmt.Sprintf("https://www.metal-archives.com/search/ajax-band-search/?field=name&query=%s", m.Search))
	return rows, columns, links
}

const (
	domain string = "https://rateyourmusic.com"
)

func (r *RateYourMusic) FindBand() ([]table.Row, []table.Column, []string) {
	c := colly.NewCollector()
	rows := make([]table.Row, 0)
	columns := []table.Column{
		{Title: "Band name", Width: 64},
		{Title: "Genere", Width: 64},
		{Title: "Country", Width: 32},
	}
	links := []string{}
	c.OnHTML("table tr.infobox", func(h *colly.HTMLElement) {
		band_link := domain + h.ChildAttr("td:not(.page_search_img_cell) a.searchpage", "href")
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

	c.Visit(fmt.Sprintf(domain+"/search?searchterm=%s&searchtype=a", strings.Replace(r.Search, " ", "%20", -1)))
	return rows, columns, links
}

func GetAlbum(album_link string) ([]table.Row, []table.Column, []string, []string, image.Image) {
	c := colly.NewCollector()
	rows := make([]table.Row, 0)
	columns := []table.Column{
		{Title: "N.", Width: 4},
		{Title: "Title", Width: 64},
		{Title: "Duration", Width: 8},
		{Title: "Lyric", Width: 16},
	}
	c.OnHTML("div#album_tabs_tracklist tr.even, div#album_tabs_tracklist tr.odd", func(h *colly.HTMLElement) {
		var row [4]string
		h.ForEach("td", func(i int, h *colly.HTMLElement) {
			row[i] = h.Text
		})
		rows = append(rows, table.Row{row[0], row[1], row[2], row[3]})
	})
	var img image.Image

	c.OnHTML("a#cover.image", func(h *colly.HTMLElement) {
		image_src := h.ChildAttr("img", "src")
		h.Request.Visit(image_src)
	})

	c.OnResponse(func(r *colly.Response) {
		if r.Headers.Get("content-type") == "image/jpeg" {
			var err error
			img, _, err = image.Decode(bytes.NewReader(r.Body))
			if err != nil {
				fmt.Println(err)
			}
		}
	})
	keys, values := []string{}, []string{}
	c.OnHTML("dl.float_right,dl.float_left", func(h *colly.HTMLElement) {
		keys, values = getMetadata(h)
	})
	c.Visit(album_link)
	return rows, columns, keys, values, img
}
