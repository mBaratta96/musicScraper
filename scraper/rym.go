package scraper

import (
	"bytes"
	"fmt"
	"image"
	"strconv"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	"github.com/gocolly/colly"
	"golang.org/x/exp/slices"
)

const (
	DOMAIN        string = "https://rateyourmusic.com"
	RYMSTYLECOLOR string = "#427b58"
)

var (
	rBandColumnTitles      = [3]string{"Band Name", "Genre", "Country"}
	rBandColumnWidths      = [3]int{64, 64, 32}
	rAlbumlistColumnTitles = [8]string{"Rec.", "Title", "Year", "Reviews", "Ratings", "Average", "Type", "Vote"}
	rAlbumlistColumnWidths = [8]int{4, 64, 4, 7, 7, 7, 12, 5}
	rAlbumColumnTitles     = [3]string{"N.", "Title", "Duration"}
	rAlbumColumnWidths     = [3]int{4, 64, 8}
)

type RateYourMusic struct {
	Link    string
	Ratings RYMRatingSlice
	Credits bool
}

type AlbumTable struct {
	Query   string
	Section string
}

func getVote(divId string, ratings RYMRatingSlice) string {
	if len(ratings.Ids) == 0 || len(ratings.Ratings) == 0 {
		return ""
	}
	splitted := strings.Split(divId, "_")
	isListened := false
	var vote int
	if release_id, err := strconv.Atoi(splitted[len(splitted)-1]); err == nil {
		if slices.Contains(ratings.Ids, release_id) {
			isListened = true
			vote = ratings.Ratings[slices.Index(ratings.Ids, release_id)]
		}
	} else {
		panic(err)
	}
	gradedVote := ""
	if isListened {
		gradedVote = fmt.Sprintf("%.1f", float32(vote)/2)
	}
	return gradedVote
}

func getAlbumListDiscography(
	data *ScrapedData,
	link string,
	tableQuery string,
	albumTables []AlbumTable,
	hasBio bool,
	userRatings RYMRatingSlice,
) {
	c := colly.NewCollector()

	c.OnHTML("div#column_container_right div.section_artist_image > a > div", func(h *colly.HTMLElement) {
		data.Metadata["Top Album"] = h.Text
	})
	if hasBio {
		c.OnHTML(
			"div#column_container_right div.section_artist_biography > span.rendered_text",
			func(h *colly.HTMLElement) {
				data.Metadata["Biography"] = strings.ReplaceAll(h.Text, "\n", " ")
			})
	}

	c.OnHTML(tableQuery, func(h *colly.HTMLElement) {
		for _, albumTable := range albumTables {
			h.ForEach(albumTable.Query, func(_ int, h *colly.HTMLElement) {
				gradedVote := getVote(h.Attr("id"), userRatings)
				title := h.ChildText("div.disco_info a.album")
				year := h.ChildText("div.disco_info span[class*='disco_year']")
				reviews := h.ChildText("div.disco_reviews")
				ratings := h.ChildText("div.disco_ratings")
				average := h.ChildText("div.disco_avg_rating")
				recommended := ""
				if h.ChildAttr("div.disco_info b.disco_mainline_recommended", "title") == "Recommended" {
					recommended = ""
				}
				data.Rows = append(
					data.Rows,
					[]string{recommended, title, year, reviews, ratings, average, albumTable.Section, gradedVote},
				)
				data.Links = append(data.Links, DOMAIN+h.ChildAttr("div.disco_info > a", "href"))
			})
		}
	})

	c.Visit(link)
}

func (r *RateYourMusic) FindBand(data *ScrapedData) ([]int, []string) {
	c := colly.NewCollector()

	c.OnHTML("table tr.infobox", func(h *colly.HTMLElement) {
		band_link := DOMAIN + h.ChildAttr("td:not(.page_search_img_cell) a.searchpage", "href")
		data.Links = append(data.Links, band_link)
		band_name := h.ChildText("td:not(.page_search_img_cell) a.searchpage")
		genres := make([]string, 0)
		h.ForEach("a.smallgreen", func(_ int, h *colly.HTMLElement) {
			genres = append(genres, h.Text)
		})
		country := h.ChildAttr("span.ui_flag", "title")
		data.Rows = append(data.Rows, []string{band_name, strings.Join(genres, "/"), country})
	})

	c.Visit(fmt.Sprintf(DOMAIN+"/search?searchterm=%s&searchtype=a", strings.Replace(r.Link, " ", "%20", -1)))
	return rBandColumnWidths[:], rBandColumnTitles[:]
}

func (r *RateYourMusic) GetAlbumList(data *ScrapedData) ([]int, []string) {
	var albumTables []AlbumTable
	var tableQuery string
	var hasBio bool
	var visitLink string

	if r.Credits {
		albumTables = []AlbumTable{{Query: "div.disco_search_results > div.disco_release", Section: "Credits"}}
		tableQuery = "div#column_container_left div.release_credits"
		hasBio = false
		visitLink = r.Link + "/credits"
	} else {
		albumTables = []AlbumTable{
			{Query: "div#disco_type_s > div.disco_release", Section: "Album"},
			{Query: "div#disco_type_l > div.disco_release", Section: "Live Album"},
			{Query: "div#disco_type_e > div.disco_release", Section: "EP"},
			{Query: "div#disco_type_a > div.disco_release", Section: "Appears On"},
			{Query: "div#disco_type_c > div.disco_release", Section: "Compilation"},
		}
		tableQuery = "div#column_container_left div#discography"
		hasBio = true
		visitLink = r.Link
	}
	getAlbumListDiscography(data, visitLink, tableQuery, albumTables, hasBio, r.Ratings)
	return rAlbumlistColumnWidths[:], rAlbumlistColumnTitles[:]
}

func (r *RateYourMusic) GetAlbum(data *ScrapedData) ([]int, []string) {
	c := colly.NewCollector()

	c.OnHTML("div#column_container_left div.page_release_art_frame", func(h *colly.HTMLElement) {
		image_url := h.ChildAttr("img", "src")
		h.Request.Visit("https:" + image_url)
	})

	c.OnResponse(func(r *colly.Response) {
		if r.Headers.Get("content-type") == "image/jpg" || r.Headers.Get("content-type") == "image/png" {
			var err error
			data.Image, _, err = image.Decode(bytes.NewReader(r.Body))
			if err != nil {
				fmt.Println(err)
			}
		}
	})

	c.OnHTML("table.album_info > tbody > tr", func(h *colly.HTMLElement) {
		key := h.ChildText("th")
		value := strings.Join(strings.Fields(strings.Replace(h.ChildText("td"), "\n", "", -1)), " ")
		if key != "Share" {
			data.Metadata[key] = value
		}
	})
	c.OnHTML("div.album_title > input.album_shortcut", func(h *colly.HTMLElement) {
		if len(r.Ratings.Ids) > 0 && len(r.Ratings.Ratings) > 0 {
			albumId := h.Attr("value")
			if id, err := strconv.Atoi(albumId[6 : len(albumId)-1]); err == nil {
				if slices.Contains(r.Ratings.Ids, id) {
					vote := r.Ratings.Ratings[slices.Index(r.Ratings.Ids, id)]
					data.Metadata["Vote"] = fmt.Sprintf("%.1f", float32(vote)/2)
				}
			}
		}
	})

	c.OnHTML("div#column_container_left div.section_tracklisting ul#tracks", func(h *colly.HTMLElement) {
		h.ForEach("li.track", func(_ int, h *colly.HTMLElement) {
			if len(h.ChildText("span.tracklist_total")) > 0 {
				value := strings.Fields(h.ChildText("span.tracklist_total"))
				data.Metadata["Total Length"] = value[len(value)-1]
			} else {
				number := h.ChildText("span.tracklist_num")
				title := h.ChildText("span[itemprop=name] span.rendered_text")
				duration := h.ChildText("span.tracklist_duration")
				data.Rows = append(data.Rows, []string{number, title, duration})
			}
		})
	})

	c.Visit(r.Link)
	return rAlbumColumnWidths[:], rAlbumColumnTitles[:]
}

func (r *RateYourMusic) GetStyleColor() string {
	return RYMSTYLECOLOR
}

func (r *RateYourMusic) SetLink(link string) {
	r.Link = link
}
