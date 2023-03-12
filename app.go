package main

import (
	"cli"
	"fmt"
	"metallum"
	"os"
	"rym"

	"github.com/joho/godotenv"
)

func main() {
	// flag.Parse()
	if len(os.Args) < 3 {
		os.Exit(1)
	}
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	website := os.Args[1]
	search := os.Args[2]
	switch website {
	case "metallum":
		rows, columns, links := metallum.FindBand(search)
		if len(links) == 0 {
			fmt.Println("No result for your search")
			os.Exit(0)
		}
		index := 0
		if len(links) > 1 {
			index = cli.PrintRows(rows, columns)
		}
		rows, columns, links = metallum.CreateRows(links[index])
		index = cli.PrintRows(rows, columns)
		rows, columns = metallum.GetAlbum(links[index])
		_ = cli.PrintRows(rows, columns)
	case "rym":
		rym.Login(os.Getenv("RYM_USERNAME"), os.Getenv("RYM_PASSWORD"))
		rows, columns, links := rym.SearchArtist(search)
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
		rows, columns = rym.GetAlbum(links[index])
		_ = cli.PrintRows(rows, columns)
	}
}
