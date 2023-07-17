package main

import (
	"cli"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"scraper"
	"strconv"
)

func checkIndex(index int) int {
	if index == -1 {
		os.Exit(0)
	}
	return index
}

func app(s scraper.Scraper) {
	data := scraper.ScrapeData(s.SearchBand)
	index := -1
	if len(data.Links) == 0 {
		fmt.Println("No result for your search")
	} else {
		index = cli.PrintTable(data.Rows, data.Columns.Title, data.Columns.Width)
	}
	index = checkIndex(index)
	s.SetLink(data.Links[index])
	data = scraper.ScrapeData(s.AlbumList)
	for true {
		cli.CallClear()
		cli.PrintMap(data.Metadata, s.StyleColor())
		index = checkIndex(cli.PrintTable(data.Rows, data.Columns.Title, data.Columns.Width))
		s.SetLink(data.Links[index])
		albumData := scraper.ScrapeData(s.Album)
		cli.CallClear()
		if albumData.Image != nil {
			cli.PrintImage(albumData.Image)
		}
		cli.PrintMap(albumData.Metadata, s.StyleColor())
		cli.PrintLink(data.Links[index])
		_ = checkIndex(cli.PrintTable(albumData.Rows, albumData.Columns.Title, albumData.Columns.Width))
		listIndex := checkIndex(cli.PrintList(s.ListChoices()))
		if listIndex == 0 {
			continue
		}
		gotCredits := false
		gotReviews := false
		var creditsData map[string]string
		var reviewData scraper.ScrapedData
		for true {
			switch listIndex {
			case 1:
				if !gotCredits {
					creditsData = s.Credits()
					gotCredits = true
				}
				cli.PrintMap(creditsData, s.StyleColor())

			case 2:
				if !gotReviews {
					reviewData = scraper.ScrapeData(s.ReviewsList)
					gotReviews = true
				}
				reviewIndex := checkIndex(
					cli.PrintTable(reviewData.Rows, reviewData.Columns.Title, reviewData.Columns.Width),
				)
				cli.PrintReview(reviewData.Links[reviewIndex])
			case 3:
				var rating string
				for true {
					fmt.Println("Insert rating (0 to 10):")
					if _, err := fmt.Scanln(&rating); err != nil {
						panic(err)
					}
					if i, err := strconv.Atoi(rating); err == nil {
						if i >= 0 && i <= 10 {
							break
						}
					}
					fmt.Println("Wrong rating value")
				}
				s.AdditionalFunctions()[3].(func(string, string))(rating, albumData.Metadata["ID"])
			}
			listIndex = checkIndex(cli.PrintList(s.ListChoices()))
			if listIndex == 0 {
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
	loginFilePath := filepath.Join(configFolder, "musicScraper", ".login.json")

	if *website == "metallum" {
		app(&scraper.Metallum{Link: search})
	} else {
		r := &scraper.RateYourMusic{}
		r.Link = search
		r.GetCredits = *rymCredits
		r.Login(loginFilePath)
		if r.Cookies != nil {
			r.DownloadUserData()
		}
		app(r)
	}
}
