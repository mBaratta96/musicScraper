package main

import (
	"cli"
	"flag"
	"fmt"
	"os"
	"scraper"
)

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
	rows, columns, links, metadata := s.GetAlbumList(links[index])
	cli.CallClear()
	cli.PrintMetadata(metadata, s.GetStyleColor())
	index = cli.PrintRows(rows, columns.Title, columns.Width)
	rows, columns, metadata, img := s.GetAlbum(links[index])
	cli.CallClear()
	if img != nil {
		cli.PrintImage(img)
	}
	cli.PrintMetadata(metadata, s.GetStyleColor())
	cli.PrintLink(links[index])
	_ = cli.PrintRows(rows, columns.Title, columns.Width)
}

func main() {
	website := flag.String("website", "", "Desired Website")
	rym_credits := flag.Bool("credits", false, "Display RYM credits")
	flag.Parse()
	if len(flag.Args()) == 0 {
		os.Exit(1)
	}
	if !(*website == "metallum" || *website == "rym") {
		fmt.Println("Wrong website")
		os.Exit(1)
	}
	search := flag.Arg(0)
	if *website == "metallum" {
		app(scraper.Metallum{Search: search})
	} else {
		app(scraper.RateYourMusic{Search: search, Credits: *rym_credits})
	}
}
