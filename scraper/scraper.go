package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"log"
	"net/http"
	"strings"

	//"github.com/gocolly/colly"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type ColumnData struct {
	Title []string
	Width []int
}

type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Solution struct {
	Url       string   `json:"url"`
	Cookies   []Cookie `json:"cookies"`
	UserAgent string   `json:"userAgent"`
}

type FlaresolverrData struct {
	Status         string   `json:"status"`
	Message        string   `json:"message"`
	Solution       Solution `json:"solution"`
	StartTimestamp int      `json:"startTimestamp"`
	EndTimestamp   int      `json:"endtimestamp"`
	Version        string   `json:"version"`
}

type ScrapedData struct {
	Rows    [][]string
	Columns ColumnData
	// optionals
	Links    []string
	Metadata *orderedmap.OrderedMap[string, string]
	Image    image.Image
}

type Scraper interface {
	SearchBand(*ScrapedData) ([]int, []string)
	AlbumList(*ScrapedData) ([]int, []string)
	Album(*ScrapedData) ([]int, []string)
	ReviewsList(*ScrapedData) ([]int, []string)
	Credits() *orderedmap.OrderedMap[string, string]
	StyleColor() string
	SetLink(string)
	ListChoices() []string
	AdditionalFunctions() map[string]any
}

var listMenuDefaultChoices = []string{"Go back", "Show credits", "Show reviews"}

type tableConstructor func(*ScrapedData) ([]int, []string)

func ScrapeData(method tableConstructor) ScrapedData {
	data := ScrapedData{}
	data.Rows = make([][]string, 0)
	colWidths, colTitles := method(&data)
	data.Columns = ColumnData{
		Title: colTitles[:],
		Width: computeColumnWidth(colWidths, colTitles, data.Rows),
	}
	return data
}

func computeColumnWidth(maxWidth []int, colTitles []string, rows [][]string) []int {
	widths := []int{}
	padding := 2
	for i, w := range maxWidth {
		maxLength := 0
		for _, row := range rows {
			cell := row[i]
			if len(cell) > maxLength {
				maxLength = len(cell)
			}
		}
		maxLength += padding
		switch {
		case maxLength > w:
			widths = append(widths, w)
		case maxLength < len(colTitles[i]):
			widths = append(widths, len(colTitles[i]))
		default:
			widths = append(widths, maxLength)
		}
	}
	return widths
}

func createCookieHeader(cookies map[string]string) string {
	cookieString := make([]string, 0)
	for name, value := range cookies {
		cookieString = append(cookieString, fmt.Sprintf("%s=%s", name, value))
	}
	return strings.Join(cookieString, "; ")
}

func GetCloudflareCookies(flaresolverrUrl string, url string) (map[string]string, string) {
	// c := createCrawler(0, nil)
	cloudflareData := map[string]string{}
	payload := map[string]string{
		"url":        fmt.Sprintf("%s", url),
		"cmd":        "request.get",
		"maxTimeout": "60000",
	}
	postBody, _ := json.Marshal(payload)
	responseBody := bytes.NewBuffer(postBody)
	//Leverage Go's HTTP Post function to make request
	resp, err := http.Post(fmt.Sprintf("%s/v1", flaresolverrUrl), "application/json", responseBody)
	//Handle Error
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	var responsePayload FlaresolverrData
	err = json.Unmarshal(body, &responsePayload)
	if err != nil {
		panic("Error in Flaresolverr response: ")
	}
	for _, cookie := range responsePayload.Solution.Cookies {
		cloudflareData[cookie.Name] = cookie.Value
	}
	return cloudflareData, responsePayload.Solution.UserAgent
}
