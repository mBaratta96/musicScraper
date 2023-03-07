package main

import (
	//"flag"
	"fmt"
	"os"

	"github.com/mBaratta96/music-scrapper/metallum"
)

func main() {
	// var album bool
	// flag.BoolVar(&album, "a", true, "Specify is search for an album")
	// flag.Parse()
	search := os.Args[1]
	// if album == true {
	// 	fmt.Println("ALBUM")
	// 	metallum.GetAlbum(search)
	// } else {
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
	//}
}
