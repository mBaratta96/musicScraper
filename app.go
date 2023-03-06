package main

import (
	"os"

	"github.com/mBaratta96/music-scrapper/cli"
	"github.com/mBaratta96/music-scrapper/scraper"
)

func main() {
	search := os.Args[1]
	rows, columns, links := scraper.FindBand(search)
	index := cli.PrintRows(rows, columns)
	rows, columns = scraper.PrintRows(links[index])
	_ = cli.PrintRows(rows, columns)
}
