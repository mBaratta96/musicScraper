package main

import (
	"cli"
	//"flag"
	"fmt"
	"os"
	"path/filepath"
	"scraper"
)

func app(s scraper.Scraper) {
	data := scraper.ScrapeData(s.FindBand)
	if len(data.Links) == 0 {
		fmt.Println("No result for your search")
		os.Exit(0)
	}
	index := 0
	if len(data.Links) > 1 {
		index = cli.PrintRows(data.Rows, data.Columns.Title, data.Columns.Width)
	}
	if index == -1 {
		os.Exit(1)
	}
	s.SetLink(data.Links[index])
	data = scraper.ScrapeData(s.GetAlbumList)
	for true {
		cli.CallClear()
		cli.PrintMetadata(data.Metadata, s.GetStyleColor())
		index = cli.PrintRows(data.Rows, data.Columns.Title, data.Columns.Width)
		if index == -1 {
			break
		}
		s.SetLink(data.Links[index])
		albumData := scraper.ScrapeData(s.GetAlbum)
		cli.CallClear()
		if albumData.Image != nil {
			cli.PrintImage(albumData.Image)
		}
		cli.PrintMetadata(albumData.Metadata, s.GetStyleColor())
		cli.PrintLink(data.Links[index])
		_ = cli.PrintRows(albumData.Rows, albumData.Columns.Title, albumData.Columns.Width)
	}
}

func main() {
	// website := flag.String("website", "", "Desired Website ('metallum' or 'rym')")
	// rymCredits := flag.Bool("credits", false, "Display RYM credits")
	//
	// flag.Parse()
	// if len(flag.Args()) == 0 {
	// 	os.Exit(1)
	// }
	// if !(*website == "metallum" || *website == "rym") {
	// 	fmt.Println("Wrong website")
	// 	os.Exit(1)
	// }
	// search := flag.Arg(0)
	configFolder, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Cannot determine config folder")
		os.Exit(1)
	}
	configFilePath := filepath.Join(configFolder, "musicScrapper", "user_albums_export.csv")
	// if *website == "metallum" {
	// 	app(&scraper.Metallum{Link: search})
	// } else {
	// 	app(&scraper.RateYourMusic{
	// 		Link:    search,
	// 		Credits: *rymCredits,
	// 		Ratings: scraper.ReadRYMRatings(configFilePath),
	// 	})
	// }
	link := "https://rateyourmusic.com/release/album/roland-kirk/i-talk-with-the-spirits/"
	rym := &scraper.RateYourMusic{Link: link, Credits: false, Ratings: scraper.ReadRYMRatings(configFilePath)}
	data := scraper.ScrapeData(rym.GetReviewsList)

	index := cli.PrintRows(data.Rows, data.Columns.Title, data.Columns.Width)
	fmt.Println(data.Links[index])
}
