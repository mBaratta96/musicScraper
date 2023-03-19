package main

import (
	"cli"
	"fmt"
	"metallum"
	"os"
	"rym"
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
		m := metallum.Metallum{search}
		rows, columns, links := m.FindBand()
		if len(links) == 0 {
			fmt.Println("No result for your search")
			os.Exit(0)
		}
		index := 0
		if len(links) > 1 {
			index = cli.PrintRows(rows, columns)
		}
		rows, columns, keys, values, links := metallum.CreateRows(links[index])
		cli.PrintMetadata(keys, values)
		index = cli.PrintRows(rows, columns)
		rows, columns, keys, values, img := metallum.GetAlbum(links[index])
		if img != nil {
			cli.PrintImage(img)
		}
		cli.PrintMetadata(keys, values)
		cli.PrintLink(links[index])
		_ = cli.PrintRows(rows, columns)
	case "rym":
		r := metallum.RateYourMusic{search}
		rows, columns, links := r.FindBand()
		if len(links) == 0 {
			fmt.Println("No result for your search")
			os.Exit(0)
		}
		index := 0
		if len(links) > 1 {
			index = cli.PrintRows(rows, columns)
		}
		rows, columns, links = rym.GetAlbumList(links[index])
		index = cli.PrintRows(rows, columns)
		rows, columns, keys, values, img := rym.GetAlbum(links[index])
		if img != nil {
			cli.PrintImage(img)
		}
		cli.PrintMetadata(keys, values)
		cli.PrintLink(links[index])
		_ = cli.PrintRows(rows, columns)
	}
}
