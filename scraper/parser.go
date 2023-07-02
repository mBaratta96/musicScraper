package scraper

import (
	"fmt"
	"io"
	"os"

	"github.com/gocarina/gocsv"
)

type RYMRating struct {
	RYMAlbumId string `csv:"RYM Album"`
	Rating     string `csv:"Rating"`
}

func readRYMRatings(path string) []RYMRating {
	ratingsFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Println("Error in leading ratings: ", err)
		os.Exit(1)
	}
	defer ratingsFile.Close()

	ratings := make([]RYMRating, 0)
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		// return csv.NewReader(in)
		return gocsv.LazyCSVReader(in) // Allows use of quotes in CSV
	})
	if err := gocsv.UnmarshalFile(ratingsFile, &ratings); err != nil { // Load clients from file
		fmt.Println("Error in leading ratings: ", err)
		os.Exit(1)
	}
	return ratings
}
