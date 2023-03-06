package main

import (
	"os"

	"github.com/mBaratta96/music-scrapper/cli"
	"github.com/mBaratta96/music-scrapper/scraper"
)

func main() {
	search := os.Args[1]
	rows, columns := scraper.FindBand(search)
	cli.PrintRows(rows, columns)
	rows, columns = scraper.PrintRows()
	cli.PrintRows(rows, columns)
}
