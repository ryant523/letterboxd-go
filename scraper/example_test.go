package letterboxd_test

import (
	"context"
	"fmt"
	"log"

	letterboxd "github.com/ryant523/letterboxd-go/scraper"
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

func ExampleClient_GetList() {
	ctx := context.Background()
	client := letterboxd.NewClient()
	list, err := client.GetList(ctx, "https://letterboxd.com/official/list/letterboxds-top-500-films/")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(list.Title)

	// To get movies, use the iterator from list.Movies
	for movie, err := range list.Movies(ctx) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(movie.Title)
	}
}
