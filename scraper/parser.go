package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/gocarina/gocsv"
)

type RYMRating struct {
	RYMAlbumId string `csv:"RYM Album"`
	Rating     string `csv:"Rating"`
}

type RYMRatingSlice struct {
	Ids     []int
	Ratings []int
}

type RYMCookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func ReadRYMRatings(path string) RYMRatingSlice {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return RYMRatingSlice{}
	}
	ratingsFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Println("Error in leading ratings: ", err)
		os.Exit(1)
	}
	defer ratingsFile.Close()

	data := make([]RYMRating, 0)
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		return gocsv.LazyCSVReader(in) // Allows use of quotes in CSV
	})
	if err := gocsv.UnmarshalFile(ratingsFile, &data); err != nil { // Load clients from file
		fmt.Println("Error in leading ratings: ", err)
		os.Exit(1)
	}
	ids := []int{}
	ratings := []int{}
	for _, d := range data {
		if id, err := strconv.Atoi(d.RYMAlbumId); err == nil {
			ids = append(ids, id)
		} else {
			panic(err)
		}
		if rating, err := strconv.Atoi(d.Rating); err == nil {
			ratings = append(ratings, rating)
		} else {
			panic(err)
		}
	}

	return RYMRatingSlice{Ids: ids, Ratings: ratings}
}

func ReadRYMCookies(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ""
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		panic("Error when opening file: ")
	}
	var cookies []RYMCookie
	err = json.Unmarshal(content, &cookies)
	if err != nil {
		panic(err)
	}
	cookie_list := make([]string, 0)
	for _, cookie := range cookies {
		cookie_list = append(cookie_list, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}
	cookie_header := strings.Join(cookie_list, "; ")
	fmt.Println(cookie_header)
	return cookie_header
}
