# letterboxd

A Letterboxd web scraper written in Go.

# Installation

`go get -u github.com/ryant523/letterboxd-go/scraper`

# Usage

To get started you must create a new Client

```go
client := letterboxd.NewClient()
```

This uses [github.com/sardanioss/httpcloak/client]("github.com/sardanioss/httpcloak/client") for fingerprinting to prevent being blocked by Cloudflare.

## Movies

```go
ctx := context.Background()
movie, err := client.GetMovieBySlug(ctx, "interstellar")

movie, err := client.GetMovieByImdb(ctx, "tt0816692")
```

## Lists

The List returns basic metadata such as "Url", "UserNames", "Title", "CreatedAt" and "UpdatedAt".

```go
list, err := client.GetList(ctx, "https://letterboxd.com/official/list/letterboxds-top-500-films/")
```

The Movies() returns an iterator to get the movies of the list. The movie contains just a "Title", "ReleaseYear" and "Slug". To get the full metadata of the movie, you must use GetMovieBySlug()

```go
for _, m := range list.Movies(ctx) {
    fmt.Printf("%s (%d) [%s]\n", m.Title, m.ReleaseYear, m.Slug)
}

```