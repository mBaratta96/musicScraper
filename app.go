package main

import (
	"cli"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"scraper"

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

func app(s scraper.Scraper) {
	rows, columns, links := s.FindBand()
	if len(links) == 0 {
		fmt.Println("No result for your search")
		os.Exit(0)
	}
	index := 0
	if len(links) > 1 {
		index = cli.PrintRows(rows, columns.Title, columns.Width)
	}
	if index == -1 {
		os.Exit(1)
	}
	rows, columns, links, metadata := s.GetAlbumList(links[index])
	for true {
		cli.CallClear()
		cli.PrintMetadata(metadata, s.GetStyleColor())
		index = cli.PrintRows(rows, columns.Title, columns.Width)
		if index == -1 {
			break
		}
		rows, columns, metadata, img := s.GetAlbum(links[index])
		cli.CallClear()
		if img != nil {
			cli.PrintImage(img)
		}
		cli.PrintMetadata(metadata, s.GetStyleColor())
		cli.PrintLink(links[index])
		_ = cli.PrintRows(rows, columns.Title, columns.Width)
	}
}

func main() {
	website := flag.String("website", "", "Desired Website")
	rymCredits := flag.Bool("credits", false, "Display RYM credits")

	flag.Parse()
	if len(flag.Args()) == 0 {
		os.Exit(1)
	}
	if !(*website == "metallum" || *website == "rym") {
		fmt.Println("Wrong website")
		os.Exit(1)
	}
	search := flag.Arg(0)
	configFolder, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Cannot determine config folder")
		os.Exit(1)
	}
	configFilePath := filepath.Join(configFolder, "musicScrapper", "user_albums_export.csv")
	if *website == "metallum" {
		app(scraper.Metallum{Search: search})
	} else {
		readRYMRatings(configFilePath)
		app(scraper.RateYourMusic{Search: search, Credits: *rymCredits})
	}
}
