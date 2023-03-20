package main

import (
	"cli"
	"fmt"
	"metallum"
	"os"
	"rym"
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
		m := scraper.Metallum{search}
		rows, links := m.FindBand()
		if len(links) == 0 {
			fmt.Println("No result for your search")
			os.Exit(0)
		}
		index := 0
		if len(links) > 1 {
			index = cli.PrintRows(rows, []string{"Band Name", "Genre", "Country"}, []int{64, 64, 32})
		}
		rows, keys, values, links := metallum.CreateRows(links[index])
		cli.PrintMetadata(keys, values)
		index = cli.PrintRows(rows, []string{"Name", "Type", "Year", "Country"}, []int{64, 16, 4, 8})
		rows, keys, values, img := metallum.GetAlbum(links[index])

		if img != nil {
			cli.PrintImage(img)
		}
		cli.PrintMetadata(keys, values)
		cli.PrintLink(links[index])
		_ = cli.PrintRows(rows, []string{"N.", "Title", "Duration", "Lyric"}, []int{4, 64, 8, 16})
	case "rym":
		rows, links := rym.SearchArtist(search)
		if len(links) == 0 {
			fmt.Println("No result for your search")
			os.Exit(0)
		}
		index := 0
		if len(links) > 1 {
			index = cli.PrintRows(rows, []string{"Band Name", "Genre", "Country"}, []int{64, 64, 32})
		}
		rows, links = rym.GetAlbumList(links[index])

		index = cli.PrintRows(
			rows,
			[]string{"Rec.", "Title", "Year", "Reviews", "Ratings", "Average", "Type"},
			[]int{4, 64, 4, 7, 7, 7, 12},
		)
		rows, keys, values, img := rym.GetAlbum(links[index])
		if img != nil {
			cli.PrintImage(img)
		}
		cli.PrintMetadata(keys, values)
		cli.PrintLink(links[index])
		_ = cli.PrintRows(rows, []string{"N.", "Title", "Duration"}, []int{4, 64, 8})
	}
}
