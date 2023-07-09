package main

import (
	"cli"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"scraper"
)

func checkIndex(index int) int {
	if index == -1 {
		os.Exit(1)
	}
	return index
}

func app(s scraper.Scraper) {
	data := scraper.ScrapeData(s.FindBand)
	if len(data.Links) == 0 {
		fmt.Println("No result for your search")
		os.Exit(0)
	}
	index := -1
	if len(data.Links) > 1 {
		index = cli.PrintRows(data.Rows, data.Columns.Title, data.Columns.Width)
	}
	index = checkIndex(index)
	s.SetLink(data.Links[index])
	data = scraper.ScrapeData(s.GetAlbumList)
	for true {
		cli.CallClear()
		cli.PrintMetadata(data.Metadata, s.GetStyleColor())
		index = checkIndex(cli.PrintRows(data.Rows, data.Columns.Title, data.Columns.Width))
		s.SetLink(data.Links[index])
		albumData := scraper.ScrapeData(s.GetAlbum)
		cli.CallClear()
		if albumData.Image != nil {
			cli.PrintImage(albumData.Image)
		}
		cli.PrintMetadata(albumData.Metadata, s.GetStyleColor())
		cli.PrintLink(data.Links[index])
		index = checkIndex(cli.PrintRows(albumData.Rows, albumData.Columns.Title, albumData.Columns.Width))
		listIndex := checkIndex(cli.PrintList())
		if listIndex == 2 {
			continue
		}
		for true {
			if listIndex == 0 {
				creditsData := scraper.ScrapeData(s.GetCredits)
				cli.PrintMetadata(creditsData.Metadata, s.GetStyleColor())
			}
			if listIndex == 1 {
				reviewData := scraper.ScrapeData(s.GetReviewsList)
				reviewIndex := checkIndex(
					cli.PrintRows(reviewData.Rows, reviewData.Columns.Title, reviewData.Columns.Width),
				)
				cli.PrintReview(reviewData.Links[reviewIndex])
			}
			listIndex = checkIndex(cli.PrintList())
			if listIndex == 2 {
				break
			}
		}
	}
}

func main() {
	website := flag.String("website", "", "Desired Website ('metallum' or 'rym')")
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
		app(&scraper.Metallum{Link: search})
	} else {
		app(&scraper.RateYourMusic{
			Link:    search,
			Credits: *rymCredits,
			Ratings: scraper.ReadRYMRatings(configFilePath),
		})
	}
}
