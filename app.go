package main

import (
	"os"

	"github.com/mBaratta96/music-scrapper/metallum"
)

func main() {
	search := os.Args[1]
	rows, columns, links := metallum.FindBand(search)
	if len(links) == 1 {
		rows, columns = metallum.CreateRows(links[0])
	} else {
		index := metallum.PrintRows(rows, columns)
		rows, columns = metallum.CreateRows(links[index])
	}
	_ = metallum.PrintRows(rows, columns)
}
