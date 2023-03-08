package main

import (
	"cli"
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
	// if album == true {
	// 	fmt.Println("ALBUM")
	// 	metallum.GetAlbum(search)
	// } else {
	switch website {
	case "metallum":
		rows, columns, links := metallum.FindBand(search)
		index := 0
		if len(links) > 1 {
			index = cli.PrintRows(rows, columns)
		}
		rows, columns, links = metallum.CreateRows(links[index])
		index = cli.PrintRows(rows, columns)
		rows, columns = metallum.GetAlbum(links[index])
		_ = cli.PrintRows(rows, columns)
	case "rym":
		rows, columns, links := rym.SearchArtist(search)
		index := 0
		if len(links) > 1 {
			index = cli.PrintRows(rows, columns)
		}
		rows, columns, links = rym.GetAlbumList(links[index])
		index = cli.PrintRows(rows, columns)
		rows, columns = rym.GetAlbum(links[index])
		_ = cli.PrintRows(rows, columns)
	}
}
