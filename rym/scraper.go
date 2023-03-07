package rym

import (
	"log"

	"github.com/gocolly/colly"
)

func Login() {
	c := colly.NewCollector()
	// authenticate
	err := c.PostMultipart("https://rateyourmusic.com/httprequest/Login",
		map[string][]byte{"user": []byte("baratz96"), "password": []byte("1qazxsw2"), "action": []byte("Login")})
	if err != nil {
		log.Fatal(err)
	}

	// attach callbacks after login
	c.OnResponse(func(r *colly.Response) {
		log.Println("response received", r.StatusCode)
	})

	// start scraping
	c.Visit("https://rateyourmusic.com/account/login")
}
