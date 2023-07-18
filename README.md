# musicScraper

CLI tool for scraping information from musical website (Rateyourmusic, Metal
Archives), with nice album ASCII art.

## Features

- Search for your favorite artists on Metallum and RateYourMusic (so far)

- Show discography and album tracklist

- Show album credits

- Show user reviews.

- **New RYM feature:** Login and rate an album.

## Usage

Clone the repo and build the package with `go build`, with Go version >= 1.18. Put the binary file in `~/.local/bin`.

To list your RYM album rating, download your profile data and save it in the
`.config/musicScrapper` folder as `user_albums_export.csv`.

To set a rating in RYM, you'll have to provide your authentication data. Create
a `.login.json` file in `~/.congig/musicScraper` and write this simple login file:

```json 
{ 
    "user": "yourRYMusername",
    "password": "yourRYMPassword"
}

```

```shell

musicScraper [OPTIONS] "name_of_artist"

-credits
        Display RYM credits
-delay int
        Set value of random delay for RYM (default 1). Set to 0 to disable.
-website string
        Desired Website ('metallum' or 'rym')
```

## Credits

Made with [Colly](https://github.com/gocolly/colly) and [Bubbletea](https://github.com/charmbracelet/bubbletea).

## Screenshots

![1](./images/1688463493.png)

![2](./images/1688464348.png)
