package scraper

import "github.com/charmbracelet/bubbles/table"

type Scraper interface {
	FindBand() ([]table.Row, []table.Column, []string)
}
