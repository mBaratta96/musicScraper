package main

import (
	"github.com/mBaratta96/music-scrapper/cli"
	"github.com/mBaratta96/music-scrapper/scraper"
)

func main() {
	rows := scraper.PrintRows()
	cli.PrintRows(rows)
}
