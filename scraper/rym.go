package scraper

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/wk8/go-ordered-map/v2"
)

const (
	DOMAIN        string = "https://rateyourmusic.com"
	RYMSTYLECOLOR string = "#427b58"
	LOGIN         string = "https://rateyourmusic.com/httprequest/Login"
	RATING        string = "https://rateyourmusic.com/httprequest/CatalogSetRating"
	USERDATA      string = "https://rateyourmusic.com/user_albums_export?album_list_id="
)

var (
	rBandColTitles      = [3]string{"Band Name", "Genre", "Country"}
	rBandColWidths      = [3]int{64, 64, 32}
	rAlbumlistColTitles = [8]string{"Rec.", "Title", "Year", "Reviews", "Ratings", "Average", "Type", "Vote"}
	rAlbumlistColWidths = [8]int{4, 64, 4, 7, 7, 7, 12, 5}
	rAlbumColTitles     = [3]string{"N.", "Title", "Duration"}
	rAlbumColWidths     = [3]int{4, 64, 8}
	rReviewColTitles    = [3]string{"User", "Date", "Rating"}
	rReviewColWidths    = [3]int{64, 16, 7}
)

type RateYourMusic struct {
	Delay      int
	Link       string
	Cookies    map[string]string
	GetCredits bool
	Expand     bool
	UserAgent  string
}

type AlbumTable struct {
	section string
	query   string
	t       rune // see AlbumList. t defines the type of the album section ('s' for albums, 'e' for EPs, etc...)
	rows    [][]string
	links   []string
}

type albumQuery struct {
	tableQuery  string
	albumTables []AlbumTable
	hasBio      bool
}

// RYM requires an async crawler with delay limitation. Otherwise your IP will be banned.
func createCrawler(delay int, cookies map[string]string, userAgent string) *colly.Collector {
	var c *colly.Collector
	if delay > 0 {
		c = colly.NewCollector(colly.Async(true), colly.UserAgent(userAgent))
		c.Limit(&colly.LimitRule{
			DomainGlob:  "*",
			Parallelism: 4,
			RandomDelay: time.Duration(delay) * time.Second,
		})
	} else {
		c = colly.NewCollector(colly.UserAgent(userAgent))
	}
	if cookies != nil {
		c.OnRequest(func(r *colly.Request) {
			r.Headers.Set("cookie", createCookieHeader(cookies))
		})
	}
	return c
}

func extractAlbumData(h *colly.HTMLElement, query string, section string, rows *[][]string, links *[]string) {
	h.ForEach(query, func(_ int, h *colly.HTMLElement) {
		rating := h.ChildText("div.disco_expandcat span.disco_cat_inner")
		title := h.ChildText("div.disco_info a.album")
		year := h.ChildText("div.disco_info span[class*='disco_year']")
		reviews := h.ChildText("div.disco_reviews")
		ratings := h.ChildText("div.disco_ratings")
		average := h.ChildText("div.disco_avg_rating")
		recommended := ""
		if h.ChildAttr("div.disco_info b.disco_mainline_recommended", "title") == "Recommended" {
			recommended = ""
		}
		row := []string{recommended, title, year, reviews, ratings, average, section, rating}
		*rows = append(*rows, row)
		*links = append(*links, DOMAIN+h.ChildAttr("div.disco_info > a", "href"))
	})
}

// https://rateyourmusic.com/search?searchterm=velvet%20underground&searchtype=a
func (r *RateYourMusic) SearchBand(data *ScrapedData) ([]int, []string) {
	data.Links = make([]string, 0)
	c := createCrawler(r.Delay, r.Cookies, r.UserAgent)

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

	c.OnError(func(res *colly.Response, err error) {
		if res.StatusCode == 403 {
			config, _ := ReadUserConfiguration()
			newCookies, newUserAgent := GetCloudflareCookies(config.FlaresolverrUrl, "http://www.rateyourmusic.com")
			r.UserAgent = newUserAgent
			for k, v := range newCookies {
				r.Cookies[k] = v
			}
			if config.SaveCookies {
				cacheCookiesAndUser := map[string]string{"user_agent": newUserAgent}
				for k, v := range r.Cookies {
					cacheCookiesAndUser[k] = v
				}
				cookieFilePath := GetCookieFilePath("rym")
				SaveCookie(cacheCookiesAndUser, cookieFilePath)
			}
		}
	})

	c.Visit(fmt.Sprintf(DOMAIN+"/search?searchterm=%s&searchtype=a", url.PathEscape(r.Link)))
	c.Wait()
	return rBandColWidths[:], rBandColTitles[:]
}

// In both cases, we are scraping table containing album data. However, the Jquery strings to retrieve the data
// from change on the type of album we want to find.
// AlbumTable is a struct for extracting album data from a particular section (Album, EP, credit...)
// tableQuery defines the wrapper (also a table) of all the tables of a page.
// hasBio is just to decide if we want to look for the artist's bio (which exists only in the main page).
var credits albumQuery = albumQuery{
	albumTables: []AlbumTable{
		{"Credits", "div.disco_search_results > div.disco_release", 'c', make([][]string, 0), make([]string, 0)},
	},
	tableQuery: "div#column_container_left div.release_credits",
	hasBio:     false,
}

var mainPage albumQuery = albumQuery{
	albumTables: []AlbumTable{
		{"Album", "div#disco_type_s > div.disco_release", 's', make([][]string, 0), make([]string, 0)},
		{"Live Album", "div#disco_type_l > div.disco_release", 'l', make([][]string, 0), make([]string, 0)},
		{"EP", "div#disco_type_e > div.disco_release", 'e', make([][]string, 0), make([]string, 0)},
		{"Appears On", "div#disco_type_a > div.disco_release", 'a', make([][]string, 0), make([]string, 0)},
		{"Compilation", "div#disco_type_c > div.disco_release", 'c', make([][]string, 0), make([]string, 0)},
	},
	tableQuery: "div#column_container_left div#discography",
	hasBio:     true,
}

// https://rateyourmusic.com/artist/the-velvet-underground
// https://rateyourmusic.com/artist/the-velvet-underground/credits/
// r.Expand does what pressing the "showing all" button does. It is applied to all the section of the main artist page.
func (r *RateYourMusic) AlbumList(data *ScrapedData) ([]int, []string) {
	var query albumQuery
	var visitLink string
	data.Links = make([]string, 0)
	data.Metadata = orderedmap.New[string, string]()

	if r.GetCredits {
		query = credits
		visitLink = r.Link + "/credits"
	} else {
		query = mainPage
		visitLink = r.Link
	}
	c := createCrawler(r.Delay, r.Cookies, r.UserAgent)
	c.OnHTML("div#column_container_right div.section_artist_image > a > div", func(h *colly.HTMLElement) {
		data.Metadata.Set("Top Album", h.Text)
	})
	if query.hasBio {
		c.OnHTML(
			"div#column_container_right div.section_artist_biography > span.rendered_text",
			func(h *colly.HTMLElement) {
				data.Metadata.Set("Biography", h.Text)
			})
	}
	if !r.Expand || r.GetCredits {
		c.OnHTML(query.tableQuery, func(h *colly.HTMLElement) {
			for i := range query.albumTables {
				albumTable := &query.albumTables[i]
				extractAlbumData(h, albumTable.query, albumTable.section, &albumTable.rows, &albumTable.links)
			}
		})
	} else {
		expandForm := map[string][]byte{
			"sort":             []byte("release_date.a,title.a"),
			"show_appearances": []byte("false"),
			"action":           []byte("ExpandDiscographySection"),
			"rym_ajax_req":     []byte("1"),
		}
		if token, err := url.PathUnescape(r.Cookies["ulv"]); err == nil {
			expandForm["request_token"] = []byte(token)
		}
		c.OnHTML("div.section_artist_name input.rym_shortcut", func(h *colly.HTMLElement) {
			artistId := h.Attr("value")
			expandForm["artist_id"] = []byte(artistId[7 : len(artistId)-1])
			for _, albumTable := range query.albumTables {
				expandForm["type"] = []byte(string(albumTable.t))
				h.Request.PostMultipart("https://rateyourmusic.com/httprequest/ExpandDiscographySection", expandForm)
			}
		})
		// Body of the response is a Javascript function, which returns a object containing the expanded HTML. We parse
		// the function, and get the substring that contains only the HTML. The HTML is parsed and data is extracted.
		c.OnResponse(func(r *colly.Response) {
			if r.Headers.Get("content-type") == "application/javascript; charset=utf-8" {
				body := string(r.Body)
				newHTML := body[strings.Index(body, "<div") : len(body)-2]
				doc, err := goquery.NewDocumentFromReader(strings.NewReader(newHTML))
				if err != nil {
					fmt.Println("Error on response")
				}
				albumType := body[strings.Index(body, "('")+2] // char after javascript function header
				albumTableIndex := slices.IndexFunc(query.albumTables, func(table AlbumTable) bool {
					return table.t == rune(albumType)
				})
				albumTable := &query.albumTables[albumTableIndex]
				albumSelector := doc.Find(fmt.Sprintf("div#disco_type_%c", albumType))
				album := colly.NewHTMLElementFromSelectionNode(r, albumSelector, albumSelector.Get(0), 0)
				extractAlbumData(album, "div.disco_release", albumTable.section, &albumTable.rows, &albumTable.links)
			}
		})
		c.OnError(func(r *colly.Response, err error) {
			fmt.Println(err)
		})

	}
	c.Visit(visitLink)
	c.Wait()
	for _, albumTable := range query.albumTables {
		fmt.Println(albumTable.links)
		data.Rows = append(data.Rows, albumTable.rows...)
		data.Links = append(data.Links, albumTable.links...)
	}
	return rAlbumlistColWidths[:], rAlbumlistColTitles[:]
}

// https://rateyourmusic.com/release/album/the-velvet-underground-nico/the-velvet-underground-and-nico/
func (r *RateYourMusic) Album(data *ScrapedData) ([]int, []string) {
	c := createCrawler(r.Delay, r.Cookies, r.UserAgent)
	data.Metadata = orderedmap.New[string, string]()

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
		value := strings.Join(strings.Fields(strings.ReplaceAll(h.ChildText("td"), "\n", "")), " ")
		if key != "Share" {
			data.Metadata.Set(key, value)
		}
	})
	c.OnHTML("div.album_title > input.album_shortcut", func(h *colly.HTMLElement) {
		albumId := h.Attr("value")
		data.Metadata.Set("ID", albumId[6:len(albumId)-1])
	})
	c.OnHTML("div#column_container_left div.section_tracklisting ul#tracks", func(h *colly.HTMLElement) {
		h.ForEach("li.track", func(_ int, h *colly.HTMLElement) {
			if len(h.ChildText("span.tracklist_total")) > 0 {
				value := strings.Fields(h.ChildText("span.tracklist_total"))
				data.Metadata.Set("Total Length", value[len(value)-1])
			} else {
				number := h.ChildText("span.tracklist_num")
				title := h.ChildText("span.tracklist_title a.song")
				duration := h.ChildText("span.tracklist_duration")
				data.Rows = append(data.Rows, []string{number, title, duration})
			}
		})
	})
	c.Visit(r.Link)
	c.Wait()
	return rAlbumColWidths[:], rAlbumColTitles[:]
}

func (r *RateYourMusic) StyleColor() string {
	return RYMSTYLECOLOR
}

func (r *RateYourMusic) SetLink(link string) {
	r.Link = link
}

// https://rateyourmusic.com/release/album/the-velvet-underground-nico/the-velvet-underground-and-nico/reviews/1/
// Recursively scrape all reviews (may generate problems for very popular albums)
func (r *RateYourMusic) ReviewsList(data *ScrapedData) ([]int, []string) {
	c := createCrawler(r.Delay, r.Cookies, r.UserAgent)
	data.Links = make([]string, 0)

	c.OnHTML("span.navspan a.navlinknext", func(h *colly.HTMLElement) {
		h.Request.Visit(h.Attr("href"))
	})
	c.OnHTML("div.review > div.review_header ", func(h *colly.HTMLElement) {
		user := h.ChildText("span.review_user")
		date := h.ChildText("span.review_date")
		rating := strings.Split(h.ChildAttr("span.review_rating > img", "alt"), " ")[0]
		data.Rows = append(data.Rows, []string{user, date, rating})
	})
	c.OnHTML("div.review > div.review_body ", func(h *colly.HTMLElement) {
		data.Links = append(data.Links, h.ChildText("span.rendered_text"))
	})

	c.Visit(r.Link + "reviews/1")
	c.Wait()
	return rReviewColWidths[:], rReviewColTitles[:]
}

func (r *RateYourMusic) Credits() *orderedmap.OrderedMap[string, string] {
	c := createCrawler(r.Delay, r.Cookies, r.UserAgent)
	credits := orderedmap.New[string, string]()

	c.OnHTML("div.section_credits > ul.credits", func(h *colly.HTMLElement) {
		h.ForEach("li[class!='expand_button']:not([style='display:none;'])", func(_ int, h *colly.HTMLElement) {
			artist := h.ChildText("a.artist")
			if len(artist) == 0 {
				artist = h.ChildText("span:not([class])")
			}
			credit := []string{}
			h.ForEach("span.role_name ", func(i int, h *colly.HTMLElement) {
				h.DOM.Contents().Not("span.role_tracks").Each(func(_ int, s *goquery.Selection) {
					credit = append(credit, strings.ToUpper(s.Text()[:1])+s.Text()[1:])
				})
			})
			credits.Set(artist, strings.Join(credit, ", "))
		})
	})

	c.Visit(r.Link)
	c.Wait()
	return credits
}

var loginForm = map[string][]byte{
	"remember":         []byte("false"),
	"maintain_session": []byte("true"),
	"rym_ajax_req":     []byte("1"),
	"action":           []byte("Login"),
}

// For reference, inspect https://rateyourmusic.com/account/login
func (r *RateYourMusic) Login() {
	user, password, err := credentials()
	if err != nil {
		panic(err)
	}
	loginForm["user"] = []byte(user)
	loginForm["password"] = []byte(password)

	// r.Cookies = make(map[string]string)
	r.Cookies["username"] = user
	c := createCrawler(r.Delay, r.Cookies, r.UserAgent)

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)
	})
	c.OnResponse(func(response *colly.Response) {
		cookies := response.Headers.Values("Set-Cookie")
		for _, cookieStr := range cookies {
			cookie := strings.Split(strings.Split(cookieStr, "; ")[0], "=")
			r.Cookies[cookie[0]] = cookie[1]
		}
	})
	c.PostMultipart(LOGIN, loginForm)
	c.Wait()
}

var ratingForm = map[string][]byte{
	"rym_ajax_req": []byte("1"),
	"action":       []byte("CatalogSetRating"),
	"type":         []byte("l"),
}

func (r *RateYourMusic) sendRating(rating string, id string) {
	c := createCrawler(r.Delay, r.Cookies, r.UserAgent)

	ratingForm["assoc_id"] = []byte(id)
	ratingForm["rating"] = []byte(rating)
	if token, err := url.PathUnescape(r.Cookies["ulv"]); err == nil {
		ratingForm["request_token"] = []byte(token)
	}

	c.OnResponse(func(r *colly.Response) {
		fmt.Println(r.StatusCode, "Vote has been uploaded.")
	})
	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)
	})

	c.PostMultipart("https://rateyourmusic.com/httprequest/CatalogSetRating", ratingForm)
	c.Wait()
}

func (r *RateYourMusic) ListChoices() []string {
	if r.Cookies != nil {
		return append(listMenuDefaultChoices, "Set rating")
	}
	return listMenuDefaultChoices
}

func (r *RateYourMusic) AdditionalFunctions() map[string]interface{} {
	return map[string]interface{}{"Set rating": r.sendRating}
}
