# musicScraper

CLI tool for scraping information from musical website (Rateyourmusic, Metal
Archives), with nice album ASCII art.

## Features

- Search for your favorite artists on Metallum and RateYourMusic (so far)

- Show discography and album tracklist

- Show album credits

- Show user reviews.

## Usage

Clone the repo and build the package with `go build`, with Go version >= 1.18.

To list your RYM album rating, download your profile data and save it in the
`.config/musicScrapper` folder as `user_albums_export.csv`.

```shell

musicScraper [OPTIONS] "name_of_artist"

-credits
        Display RYM credits
  -website string
        Desired Website ('metallum' or 'rym')
```

## Screenshots

![1](./images/1688463493.png)

![2](./images/1688464348.png)
