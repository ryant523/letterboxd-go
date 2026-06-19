package letterboxd_test

import (
	"context"
	"fmt"
	"log"

	letterboxd "github.com/yourusername/letterboxd-go/scraper"
)

func ExampleClient_GetMovie() {
	client := letterboxd.NewClient(letterboxd.WithTimeout(10),
		letterboxd.WithRetry(3),
	)
	movie, err := client.GetMovie(context.Background(), "la-la-land")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(movie.Title)
	fmt.Println(movie.ReleaseYear)
}
