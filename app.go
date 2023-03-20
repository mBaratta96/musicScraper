package main

import (
	"cli"
	"fmt"
	"os"
	"scraper"
)

func main() {
	// flag.Parse()
	if len(os.Args) < 3 {
		os.Exit(1)
	}
	website := os.Args[1]
	search := os.Args[2]
	switch website {
	case "metallum":
		m := scraper.Metallum{Search: search}
		rows, links := m.FindBand()
		if len(links) == 0 {
			fmt.Println("No result for your search")
			os.Exit(0)
		}
		index := 0
		if len(links) > 1 {
			index = cli.PrintRows(rows, []string{"Band Name", "Genre", "Country"}, []int{64, 64, 32})
		}
		rows, links, metadata := m.GetAlbumList(links[index])
		cli.PrintMetadata(metadata)
		index = cli.PrintRows(rows, []string{"Name", "Type", "Year", "Country"}, []int{64, 16, 4, 8})
		rows, metadata, img := m.GetAlbum(links[index])

		if img != nil {
			cli.PrintImage(img)
		}
		cli.PrintMetadata(metadata)
		cli.PrintLink(links[index])
		_ = cli.PrintRows(rows, []string{"N.", "Title", "Duration", "Lyric"}, []int{4, 64, 8, 16})
	case "rym":
		r := scraper.RateYourMusic{Search: search}
		rows, links := r.FindBand()
		if len(links) == 0 {
			fmt.Println("No result for your search")
			os.Exit(0)
		}
		index := 0
		if len(links) > 1 {
			index = cli.PrintRows(rows, []string{"Band Name", "Genre", "Country"}, []int{64, 64, 32})
		}
		rows, links, metadata := r.GetAlbumList(links[index])
		cli.PrintMetadata(metadata)
		index = cli.PrintRows(
			rows,
			[]string{"Rec.", "Title", "Year", "Reviews", "Ratings", "Average", "Type"},
			[]int{4, 64, 4, 7, 7, 7, 12},
		)
		rows, metadata, img := r.GetAlbum(links[index])
		if img != nil {
			cli.PrintImage(img)
		}
		cli.PrintMetadata(metadata)
		cli.PrintLink(links[index])
		_ = cli.PrintRows(rows, []string{"N.", "Title", "Duration"}, []int{4, 64, 8})
	}
}
