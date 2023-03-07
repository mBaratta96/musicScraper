package main

import (
	"fmt"
	"metallum"
	"os"
	//"rym"
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
		if len(links) == 1 {
			rows, columns, links = metallum.CreateRows(links[index])
		} else {
			index = metallum.PrintRows(rows, columns, false)
			rows, columns, links = metallum.CreateRows(links[index])
		}
		index = metallum.PrintRows(rows, columns, false)
		rows, columns = metallum.GetAlbum(links[index])
		fmt.Println(links[index])
		_ = metallum.PrintRows(rows, columns, true)
	//
	case "rym":
		fmt.Println("RYM")

	}
}
