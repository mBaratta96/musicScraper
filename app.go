package main

import (
	"os"

	"github.com/mBaratta96/music-scrapper/cli"
	"github.com/mBaratta96/music-scrapper/scraper"
)

func main() {
	search := os.Args[1]
	scraper.FindBand(search)
	rows, columns := scraper.PrintRows()
	cli.PrintRows(rows, columns)
}
