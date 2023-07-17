package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/gocarina/gocsv"
)

type RYMRating struct {
	RYMAlbumId string `csv:"RYM Album"`
	Rating     string `csv:"Rating"`
}

type LoginData struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type RYMCookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func readRYMRatings(payload []byte) map[int]int {
	data := make([]RYMRating, 0)
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		return gocsv.LazyCSVReader(in) // Allows use of quotes in CSV
	})
	if err := gocsv.UnmarshalBytes(payload, &data); err != nil { // Load clients from file
		fmt.Println("Error in leading ratings: ", err)
		os.Exit(1)
	}

	ratings := make(map[int]int)
	for _, d := range data {
		if id, err := strconv.Atoi(d.RYMAlbumId); err != nil {
			panic(err)
		} else if rating, err := strconv.Atoi(d.Rating); err != nil {
			panic(err)
		} else {
			ratings[id] = rating
		}
	}

	return ratings
}

func readUserLoginData(path string) (string, string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic(err)
	}
	loginFile, err := ioutil.ReadFile(path)
	if err != nil {
		panic("Error when opening file: ")
	}
	var loginData LoginData
	err = json.Unmarshal(loginFile, &loginData)
	if err != nil {
		panic(err)
	}
	return loginData.User, loginData.Password
}
