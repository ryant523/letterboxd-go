package letterboxd_test

import (
	"context"
	"fmt"
	"log"

	letterboxd "github.com/ryant523/letterboxd-go/scraper"
)

func ExampleClient_GetMovieBySlug() {
	client := letterboxd.NewClient(letterboxd.WithTimeout(10),
		letterboxd.WithRetry(3),
	)
	defer client.Close()
	movie, err := client.GetMovieBySlug(context.Background(), "la-la-land")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(movie.Title)
	fmt.Println(movie.ReleaseYear)
}

func ExampleClient_GetList() {
	ctx := context.Background()
	client := letterboxd.NewClient()
	defer client.Close()
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

	// can also get a slice of all movies in the list

	movies, err := list.GetAllMovies(ctx)
	fmt.Println(len(movies))
}
