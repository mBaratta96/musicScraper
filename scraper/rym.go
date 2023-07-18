package scraper

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strconv"
	"strings"
	"time"

	//"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

const (
	DOMAIN        string = "https://rateyourmusic.com"
	RYMSTYLECOLOR string = "#427b58"
	LOGIN         string = "https://rateyourmusic.com/httprequest/Login"
	RATING        string = "https://rateyourmusic.com/httprequest/CatalogSetRating"
	USERDATA      string = "https://rateyourmusic.com/user_albums_export?album_list_id="
)

var (
	rBandColumnTitles      = [3]string{"Band Name", "Genre", "Country"}
	rBandColumnWidths      = [3]int{64, 64, 32}
	rAlbumlistColumnTitles = [8]string{"Rec.", "Title", "Year", "Reviews", "Ratings", "Average", "Type", "Vote"}
	rAlbumlistColumnWidths = [8]int{4, 64, 4, 7, 7, 7, 12, 5}
	rAlbumColumnTitles     = [3]string{"N.", "Title", "Duration"}
	rAlbumColumnWidths     = [3]int{4, 64, 8}
	rReviewColumnTitles    = [3]string{"User", "Date", "Rating"}
	rReviewColumnWidths    = [3]int{64, 16, 7}
)

type RateYourMusic struct {
	Delay      int
	Link       string
	Ratings    map[int]int
	Cookies    map[string]string
	GetCredits bool
}

type AlbumTable struct {
	Query   string
	Section string
}

func createCrawler(delay int) *colly.Collector {
	c := colly.NewCollector(
		colly.Async(true),
		colly.UserAgent(
			"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0",
		))
	if delay > 0 {
		c.Limit(&colly.LimitRule{
			DomainGlob:  "*",
			Parallelism: 2,
			RandomDelay: time.Duration(delay) * time.Second,
		})
	}
	return c
}

func getVote(divId string, ratings map[int]int) string {
	if len(ratings) == 0 {
		return ""
	}
	splitted := strings.Split(divId, "_")
	isListened := false
	var vote int
	if releaseId, err := strconv.Atoi(splitted[len(splitted)-1]); err == nil {
		if rating, ok := ratings[releaseId]; ok {
			isListened = true
			vote = rating
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
	userRatings map[int]int,
	delay int,
) {
	c := createCrawler(delay)

	c.OnHTML("div#column_container_right div.section_artist_image > a > div", func(h *colly.HTMLElement) {
		data.Metadata["Top Album"] = h.Text
	})
	if hasBio {
		c.OnHTML(
			"div#column_container_right div.section_artist_biography > span.rendered_text",
			func(h *colly.HTMLElement) {
				data.Metadata["Biography"] = h.Text
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
	c.Wait()
}

func (r *RateYourMusic) SearchBand(data *ScrapedData) ([]int, []string) {
	data.Links = make([]string, 0)
	c := createCrawler(r.Delay)

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
	c.Wait()
	return rBandColumnWidths[:], rBandColumnTitles[:]
}

func (r *RateYourMusic) AlbumList(data *ScrapedData) ([]int, []string) {
	var albumTables []AlbumTable
	var tableQuery string
	var hasBio bool
	var visitLink string
	data.Links = make([]string, 0)
	data.Metadata = make(map[string]string)

	if r.GetCredits {
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
	getAlbumListDiscography(data, visitLink, tableQuery, albumTables, hasBio, r.Ratings, r.Delay)
	return rAlbumlistColumnWidths[:], rAlbumlistColumnTitles[:]
}

func (r *RateYourMusic) Album(data *ScrapedData) ([]int, []string) {
	c := createCrawler(r.Delay)
	data.Metadata = make(map[string]string)

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
		if len(r.Ratings) > 0 {
			albumId := h.Attr("value")
			if id, err := strconv.Atoi(albumId[6 : len(albumId)-1]); err == nil {
				data.Metadata["ID"] = fmt.Sprintf("%d", id)
				if rating, ok := r.Ratings[id]; ok {
					data.Metadata["Vote"] = fmt.Sprintf("%.1f", float32(rating)/2)
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
	c.Wait()
	return rAlbumColumnWidths[:], rAlbumColumnTitles[:]
}

func (r *RateYourMusic) StyleColor() string {
	return RYMSTYLECOLOR
}

func (r *RateYourMusic) SetLink(link string) {
	r.Link = link
}

func (r *RateYourMusic) ReviewsList(data *ScrapedData) ([]int, []string) {
	c := createCrawler(r.Delay)
	data.Links = make([]string, 0)

	c.OnHTML("span.navspan a.navlinknext", func(h *colly.HTMLElement) {
		h.Request.Visit(h.Attr("href"))
	})

	c.OnHTML("div.review > div.review_header ", func(h *colly.HTMLElement) {
		var row [3]string
		row[0] = h.ChildText("span.review_user")
		row[1] = h.ChildText("span.review_date")
		row[2] = strings.Split(h.ChildAttr("span.review_rating > img", "alt"), " ")[0]
		data.Rows = append(data.Rows, row[:])
	})

	c.OnHTML("div.review > div.review_body ", func(h *colly.HTMLElement) {
		data.Links = append(data.Links, h.ChildText("span.rendered_text"))
	})

	c.Visit(r.Link + "reviews/1")
	c.Wait()
	return rReviewColumnWidths[:], rReviewColumnTitles[:]
}

func (r *RateYourMusic) Credits() map[string]string {
	c := createCrawler(r.Delay)
	credits := make(map[string]string)

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
			credits[artist] = strings.Join(credit, ", ")
		})
	})

	c.Visit(r.Link)
	c.Wait()
	return credits
}

func (r *RateYourMusic) Login(path string) {
	user, password, err := readUserLoginData(path)
	if err != nil {
		return
	}
	formRequest := map[string][]byte{
		"user":             []byte(user),
		"password":         []byte(password),
		"remember":         []byte("false"),
		"maintain_session": []byte("true"),
		"rym_ajax_req":     []byte("1"),
		"action":           []byte("Login"),
	}
	r.Cookies = make(map[string]string)
	c := createCrawler(r.Delay)

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

	c.PostMultipart(LOGIN, formRequest)
	c.Wait()
}

func createCookieHeader(cookies map[string]string) string {
	cookieString := make([]string, 0)
	for cookieName, cookieValue := range cookies {
		cookieString = append(cookieString, fmt.Sprintf("%s=%s", cookieName, cookieValue))
	}
	return strings.Join(cookieString, "; ")
}

func (r *RateYourMusic) DownloadUserData() {
	var userId string
	c := createCrawler(r.Delay)

	c.OnHTML("div.bubble_header.profile_header ", func(h *colly.HTMLElement) {
		headerText := strings.Fields(h.Text)
		for _, text := range headerText {
			if strings.HasPrefix(text, "#") {
				userId = text[1:]
			}
		}
		h.Request.Visit(USERDATA + userId + "&noreview")
	})

	c.OnRequest(func(request *colly.Request) {
		request.Headers.Set("Cookie", createCookieHeader(r.Cookies))
	})
	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)
	})
	c.OnResponse(func(response *colly.Response) {
		if response.Headers.Get("content-type") == "text/plain; charset=utf-8" {
			r.Ratings = readRYMRatings(response.Body)
		}
	})

	c.Visit(DOMAIN + "/~" + r.Cookies["username"])
	c.Wait()
}

func (r *RateYourMusic) SendRating(rating string, id string) {
	c := createCrawler(r.Delay)

	formRequest := map[string][]byte{
		"type":          []byte("l"),
		"assoc_id":      []byte(id),
		"rating":        []byte(rating),
		"action":        []byte("CatalogSetRating"),
		"rym_ajax_req":  []byte("1"),
		"request_token": []byte(strings.ReplaceAll(r.Cookies["ulv"], "%2e", ".")),
	}

	c.OnRequest(func(req *colly.Request) {
		req.Headers.Set("Cookie", createCookieHeader(r.Cookies))
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println(r.StatusCode, "Vote has been uploaded.")
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)
	})

	c.PostMultipart("https://rateyourmusic.com/httprequest/CatalogSetRating", formRequest)
	c.Wait()
}

func (r *RateYourMusic) ListChoices() []string {
	if r.Cookies != nil {
		return append(listMenuDefaultChoices, "Set rating")
	}
	return listMenuDefaultChoices
}

func (r *RateYourMusic) AdditionalFunctions() map[int]interface{} {
	return map[int]interface{}{3: r.SendRating}
}
