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

func ExampleClient_GetDiary() {
	ctx := context.Background()
	client := letterboxd.NewClient()
	defer client.Close()

	diary, err := client.GetDiary(
		ctx,
		"userName",
		letterboxd.WithWatchedYear("2026"),
		letterboxd.WithDirector("akira-kurosawa"),
		letterboxd.WithRating(letterboxd.RatingFourHalf),
	)
	if err != nil {
		log.Panic(err)
	}

	for entry, err := range diary.Entries(ctx) {
		if err != nil {
			log.Panic(err)
		}
		fmt.Printf("%s - %s (%d)\n", entry.DateWatched.Format("20060102"), entry.Title, entry.ReleaseYear)

	}
}

func ExampleClient_GetUser() {
	ctx := context.Background()
	client := letterboxd.NewClient()
	defer client.Close()

	user, err := client.GetUser(ctx, "userName")
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(user.DisplayName)
	movies := make([]*letterboxd.Movie, 0, 4)
	for _, top4 := range user.TopFour {
		m, err := client.GetMovieBySlug(ctx, top4.Slug)
		if err != nil {
			log.Panic(err)
		}
		movies = append(movies, m)
	}

	fmt.Printf("%+v\n", movies)
}
