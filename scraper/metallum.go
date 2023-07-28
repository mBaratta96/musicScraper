package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"strings"

	_ "image/jpeg"
	_ "image/png"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/wk8/go-ordered-map/v2"
)

type SearchResponse struct {
	Error                string     `json:"error"`
	ITotalRecords        int        `json:"iTotalRecords"`
	ITotalDisplayRecords int        `json:"iTotalDisplayRecords"`
	SEcho                int        `json:"sEcho"`
	AaData               [][]string `json:"aaData"`
}

const METALLUMSTYLECOLOR string = "#b57614"

var (
	mBandColumnTitles      = [3]string{"Band Name", "Genre", "Country"}
	mBandColumnWidths      = [3]int{64, 64, 32}
	mAlbumlistColumnTitles = [4]string{"Name", "Type", "Year", "Reviews"}
	mAlbumlistColumnWidths = [4]int{64, 16, 4, 8}
	mAlbumColumnTitles     = [4]string{"N.", "Title", "Duration", "Lyric"}
	mAlbumColumnWidths     = [4]int{4, 64, 8, 16}
	mReviewColumnTitles    = [4]string{"Title", "Rating", "User", "Date"}
	mReviewColumnWidths    = [4]int{32, 7, 32, 32}
)

type Metallum struct {
	Link string
}

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

func (m *Metallum) SearchBand(data *ScrapedData) ([]int, []string) {
	c := colly.NewCollector()
	data.Links = make([]string, 0)

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
					data.Links = append(data.Links, band.AttrOr("href", ""))
				case 1:
					row[1] = node
				case 2:
					row[2] = node
				}
			}
			data.Rows = append(data.Rows, row[:])
		}
	})

	c.Visit(fmt.Sprintf("https://www.metal-archives.com/search/ajax-band-search/?field=name&query=%s", m.Link))
	return mBandColumnWidths[:], mAlbumColumnTitles[:]
}

func (m *Metallum) AlbumList(data *ScrapedData) ([]int, []string) {
	c := colly.NewCollector()
	data.Links = make([]string, 0)
	data.Metadata = orderedmap.New[string, string]()

	c.OnHTML("#band_disco a[href*='all']", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

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
	return mAlbumlistColumnWidths[:], mAlbumlistColumnTitles[:]
}

func (m *Metallum) Album(data *ScrapedData) ([]int, []string) {
	c := colly.NewCollector()
	data.Links = make([]string, 0)
	data.Metadata = orderedmap.New[string, string]()

	c.OnHTML("div#album_tabs_tracklist tr.even, div#album_tabs_tracklist tr.odd", func(h *colly.HTMLElement) {
		var row [4]string
		h.ForEach("td", func(i int, h *colly.HTMLElement) {
			row[i] = h.Text
		})
		data.Rows = append(data.Rows, row[:])
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
	return mAlbumColumnWidths[:], mAlbumColumnTitles[:]
}

func (m *Metallum) StyleColor() string {
	return METALLUMSTYLECOLOR
}

func (m *Metallum) SetLink(link string) {
	m.Link = link
}

func (m *Metallum) ReviewsList(data *ScrapedData) ([]int, []string) {
	c := colly.NewCollector()
	data.Links = make([]string, 0)

	c.OnHTML("div#album_tabs_reviews tr.even, div#album_tabs_reviews tr.odd", func(h *colly.HTMLElement) {
		var row [4]string
		var link string
		i := 0
		h.ForEach("td", func(_ int, h *colly.HTMLElement) {
			if len(h.Attr("nowrap")) == 0 {
				row[i] = h.Text
				i += 1
			} else {
				link = h.ChildAttrs("a", "href")[0]
				h.Request.Visit(link)
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
	return mReviewColumnWidths[:], mReviewColumnTitles[:]
}

func (m *Metallum) Credits() *orderedmap.OrderedMap[string, string] {
	c := colly.NewCollector()
	credits := orderedmap.New[string, string]()

	c.OnHTML("div#album_members_lineup table.lineupTable > tbody > tr.lineupRow", func(h *colly.HTMLElement) {
		artist := h.ChildText("td:has(a)")
		credit := h.ChildText("td:not(:has(a))")
		credits.Set(artist, credit)
	})

	c.Visit(m.Link)
	return credits
}

func (m *Metallum) ListChoices() []string {
	return listMenuDefaultChoices
}

func (m *Metallum) AdditionalFunctions() map[int]interface{} {
	return map[int]interface{}{}
}
