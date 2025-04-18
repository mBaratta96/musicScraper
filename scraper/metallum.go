package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"path"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/wk8/go-ordered-map/v2"
)

type searchResponse struct {
	AaData [][]string `json:"aaData"`
}

const METALLUMSTYLECOLOR string = "#b57614"

var (
	mBandColTitles          = [3]string{"Band Name", "Genre", "Country"}
	mBandColWidths          = [3]int{64, 64, 32}
	mAlbumlistColTitles     = [4]string{"Name", "Type", "Year", "Reviews"}
	mAlbumlistColWidths     = [4]int{64, 16, 4, 8}
	mAlbumColTitles         = [4]string{"N.", "Title", "Duration", "Lyric"}
	mAlbumColWidths         = [4]int{4, 64, 8, 16}
	mReviewColTitles        = [4]string{"Title", "Rating", "User", "Date"}
	mReviewColWidths        = [4]int{32, 7, 32, 32}
	mSimilarArtistColTitles = [4]string{"Name", "Country", "Genre", "Score"}
	mSimilarArtistColWidths = [4]int{64, 32, 64, 4}
)

type Metallum struct {
	Link      string
	Cookies   map[string]string
	UserAgent string
}

// Metadata contains info (country of origin, genre, themes...) for band page and label for albums.
// For reference inspect:
// https://www.metal-archives.com/bands/Emperor/30
// https://www.metal-archives.com/albums/Emperor/Anthems_to_the_Welkin_at_Dusk/92
func getMetadata(h *colly.HTMLElement, metadata *orderedmap.OrderedMap[string, string]) {
	keys, values := []string{}, []string{}
	h.ForEach("dt", func(_ int, h *colly.HTMLElement) {
		keys = append(keys, h.Text)
	})
	h.ForEach("dd", func(_ int, h *colly.HTMLElement) {
		values = append(values, strings.Replace(h.Text, "\n", "", -1))
	})
	for i, k := range keys {
		metadata.Set(k, values[i])
	}
}

// Metallum search page renders the result of a query from a JSON payload.
// https://www.metal-archives.com/search?searchString=emperor&type=band_name
func (m *Metallum) SearchBand(data *ScrapedData) ([]int, []string) {
	c := createCrawler(0, m.Cookies, m.UserAgent)
	data.Links = make([]string, 0)

	c.OnError(func(res *colly.Response, err error) {
		if res.StatusCode == 403 {
			config, _ := ReadUserConfiguration()
			newCookies, newUserAgent := GetCloudflareCookies(config.FlaresolverrUrl, "http://www.rateyourmusic.com")
			m.UserAgent = newUserAgent
			cacheCookiesAndUser := map[string]string{"user_agent": newUserAgent}
			for k, v := range newCookies {
				cacheCookiesAndUser[k] = v
				m.Cookies[k] = v
			}
			if config.SaveCookies {
				cookieFilePath := GetCookieFilePath("rym")
				SaveCookie(cacheCookiesAndUser, cookieFilePath)
			}
		}
	})

	c.OnResponse(func(r *colly.Response) {
		var response searchResponse
		if err := json.Unmarshal(r.Body, &response); err != nil {
			fmt.Println("Can not unmarshal JSON")
		}
		for _, el := range response.AaData {
			// Search results are contained in the first element of the JSON array as a HTML string.
			// We parse it and get the data.
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(el[0]))
			if err != nil {
				fmt.Println("Error on response")
			}
			bandLink := doc.Find("a").First()
			band := bandLink.Text()
			data.Links = append(data.Links, bandLink.AttrOr("href", ""))
			genre := el[1]
			country := el[2]
			data.Rows = append(data.Rows, []string{band, genre, country})
		}
	})
	c.Visit(fmt.Sprintf("http://www.metal-archives.com/search/ajax-band-search/?field=name&query=%s", m.Link))
	return mBandColWidths[:], mAlbumColTitles[:]
}

// https://www.metal-archives.com/bands/Emperor/30
func (m *Metallum) AlbumList(data *ScrapedData) ([]int, []string) {
	c := createCrawler(0, m.Cookies, m.UserAgent)
	data.Links = make([]string, 0)
	data.Metadata = orderedmap.New[string, string]()

	// Get link to table with all albums
	c.OnHTML("#band_disco a[href*='all']", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})
	// Scrape the table
	c.OnHTML("table.display.discog tbody tr", func(h *colly.HTMLElement) {
		var row [4]string
		h.ForEach(".album,.demo,.other,td a[href]", func(i int, h *colly.HTMLElement) {
			row[i] = h.Text
			if i == 0 {
				data.Links = append(data.Links, h.Attr("href"))
			}
		})
		data.Rows = append(data.Rows, row[:])
	})
	c.OnHTML("dl.float_right,dl.float_left", func(h *colly.HTMLElement) {
		getMetadata(h, data.Metadata)
	})

	c.Visit(m.Link)
	return mAlbumlistColWidths[:], mAlbumlistColTitles[:]
}

// https://www.metal-archives.com/albums/Emperor/Anthems_to_the_Welkin_at_Dusk/92
func (m *Metallum) Album(data *ScrapedData) ([]int, []string) {
	c := createCrawler(0, m.Cookies, m.UserAgent)
	data.Links = make([]string, 0)
	data.Metadata = orderedmap.New[string, string]()

	c.OnHTML("div#album_tabs_tracklist tr.even, div#album_tabs_tracklist tr.odd", func(h *colly.HTMLElement) {
		var row [4]string
		h.ForEach("td", func(i int, h *colly.HTMLElement) {
			row[i] = h.Text
		})
		data.Rows = append(data.Rows, row[:])
	})
	// Get band id (useful for checking similar bands later)
	c.OnHTML("h2.band_name > a", func(h *colly.HTMLElement) {
		data.Metadata.Set("ID", path.Base(h.Attr("href")))
	})
	c.OnHTML("a#cover.image", func(h *colly.HTMLElement) {
		image_src := h.ChildAttr("img", "src")
		h.Request.Visit(image_src)
	})
	c.OnResponse(func(r *colly.Response) {
		if r.Headers.Get("content-type") == "image/jpeg" {
			var err error
			data.Image, _, err = image.Decode(bytes.NewReader(r.Body))
			if err != nil {
				fmt.Println(err)
			}
		}
	})
	c.OnHTML("dl.float_right,dl.float_left", func(h *colly.HTMLElement) {
		getMetadata(h, data.Metadata)
	})

	c.Visit(m.Link)
	return mAlbumColWidths[:], mAlbumColTitles[:]
}

func (m *Metallum) StyleColor() string {
	return METALLUMSTYLECOLOR
}

func (m *Metallum) SetLink(link string) {
	m.Link = link
}

// https://www.metal-archives.com/albums/Emperor/Anthems_to_the_Welkin_at_Dusk/92
func (m *Metallum) ReviewsList(data *ScrapedData) ([]int, []string) {
	c := createCrawler(0, m.Cookies, m.UserAgent)
	data.Links = make([]string, 0)

	c.OnHTML("div#album_tabs_reviews tr.even, div#album_tabs_reviews tr.odd", func(h *colly.HTMLElement) {
		var row [4]string
		i := 0
		h.ForEach("td", func(_ int, h *colly.HTMLElement) {
			if len(h.Attr("nowrap")) == 0 {
				row[i] = h.Text
				i += 1
			} else {
				h.Request.Visit(h.ChildAttrs("a", "href")[0])
			}
		})
		data.Rows = append(data.Rows, row[:])
	})
	c.OnHTML("div.reviewBox", func(h *colly.HTMLElement) {
		review := h.ChildText("h3.reviewTitle") + "\n"
		review += h.ChildText("div:not([attr_all])") + "\n"
		review += h.ChildText("div.reviewContent")
		data.Links = append(data.Links, review)
	})

	c.Visit(m.Link)
	return mReviewColWidths[:], mReviewColTitles[:]
}

// https://www.metal-archives.com/albums/Emperor/Anthems_to_the_Welkin_at_Dusk/92
func (m *Metallum) Credits() *orderedmap.OrderedMap[string, string] {
	c := createCrawler(0, m.Cookies, m.UserAgent)
	credits := orderedmap.New[string, string]()

	c.OnHTML("div#album_members_lineup table.lineupTable > tbody > tr.lineupRow", func(h *colly.HTMLElement) {
		artist := strings.ReplaceAll(h.ChildText("td:has(a)"), "\n", " ") // If it has a RIP remove new line
		credit := h.ChildText("td:not(:has(a))")
		credits.Set(artist, credit)
	})

	c.Visit(m.Link)
	return credits
}

func (m *Metallum) similarArtists(data *ScrapedData) ([]int, []string) {
	c := createCrawler(0, m.Cookies, m.UserAgent)
	data.Links = make([]string, 0)

	c.OnHTML("table#artist_list tbody tr[id*='recRow']", func(h *colly.HTMLElement) {
		var row [4]string
		h.ForEach("td", func(i int, h *colly.HTMLElement) {
			row[i] = h.Text
			if i == 0 {
				data.Links = append(data.Links, h.ChildAttr("a", "href"))
			}
		})
		data.Rows = append(data.Rows, row[:])
	})
	// This makes len(data.Rows) = len(data.Links) + 1 (see app.go)
	c.OnScraped(func(_ *colly.Response) {
		data.Rows = append(data.Rows, []string{"Go back to choices", "", "", ""})
	})

	c.Visit(m.Link)
	return mSimilarArtistColWidths[:], mSimilarArtistColTitles[:]
}

func (m *Metallum) ListChoices() []string {
	return append(listMenuDefaultChoices, "Get similar artists")
}

func (m *Metallum) AdditionalFunctions() map[string]interface{} {
	return map[string]interface{}{"Get similar artists": m.similarArtists}
}
