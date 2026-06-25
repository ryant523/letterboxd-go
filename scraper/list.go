package letterboxd

import (
	"context"
	"fmt"
	"iter"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ListMovie represents a single movie entry extracted from a Letterboxd list overview page.
// To get the full movie use [Client.GetMovie] with the Slug.
type ListMovie struct {
	Title       string
	ReleaseYear int
	Slug        string
	Rank        int
}

// List represents a parsed Letterboxd movie list container along with its metadata
// and internal pagination tracking capabilities.
type List struct {
	Url       string
	UserNames []string
	Title     string
	Length    int
	CreatedAt time.Time
	UpdatedAt time.Time

	client      *Client
	firstMovies []*ListMovie
	nextLink    string
}

// Movies returns a lazy-loaded iterator sequence for all entries in the list.
// It automatically flushes the pre-fetched first page, then handles on-demand
// HTTP calls for subsequent pages only as the loop advances.
func (l *List) Movies(ctx context.Context) iter.Seq2[*ListMovie, error] {
	return func(yield func(*ListMovie, error) bool) {
		for _, m := range l.firstMovies {
			if !yield(m, nil) {
				return // User broke out of loop early
			}
		}
		if l.nextLink == "" {
			return
		}

		nextLink := l.nextLink
		for {
			path := nextLink
			doc, err := l.client.getHtml(ctx, path)
			if err != nil {
				yield(&ListMovie{}, fmt.Errorf("failed fetching list: %s, %w", path, err))
				return
			}
			var movies []*ListMovie
			movies, nextLink = parseListItems(doc)
			if len(movies) == 0 {
				return
			}
			for _, m := range movies {
				if !yield(m, nil) {
					return
				}
			}
			if nextLink == "" {
				return
			}
		}
	}

}

// Get All movies in the list
func (l *List) GetAllMovies(ctx context.Context) ([]*ListMovie, error) {
	movies := make([]*ListMovie, 0, l.Length)
	for m, err := range l.Movies(ctx) {
		if err != nil {
			return nil, err
		}
		movies = append(movies, m)
	}
	return movies, nil
}

// GetList requests a Letterboxd list by its fully qualified URL, parses the metadata,
// aggregates the initial chunk of movies, and initializes the tracking struct pointer.
func (c *Client) GetList(ctx context.Context, u string, opts ...Option) (*List, error) {
	baseUrl, _ := url.Parse(u)
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	if optionsPath := options.build(); optionsPath != "" {
		baseUrl.Path = path.Join(baseUrl.Path, optionsPath)
	}
	doc, err := c.getHtml(ctx, baseUrl.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch list base page: %w", err)
	}

	list, err := parseMovieList(doc)
	if err != nil {
		return nil, err
	}
	list.client = c
	return list, nil
}

// parseMovieList extracts all high-level list metadata and the first batch
// of film items from a loaded goquery document tree.
func parseMovieList(doc *goquery.Document) (*List, error) {

	headerSel := doc.Find("header.person-header")

	items, nextLink := parseListItems(doc)

	return &List{
		Url:         getUrlFromList(doc), // Left as doc since it checks <head> meta tags
		UserNames:   getUserNamesFromList(headerSel),
		Title:       getTitleFromList(doc),
		Length:      getNumberOfMoviesFromList(doc),
		nextLink:    nextLink,
		firstMovies: items,
		CreatedAt:   getPublishedTime(doc),
		UpdatedAt:   getUpdatedTime(doc),
	}, nil
}

// parseListItems processes a single page's poster grid and extracts its items,
// alongside searching for any valid forward pagination link.
func parseListItems(doc *goquery.Document) ([]*ListMovie, string) {
	posterGridSel := doc.Find("ul.poster-list")
	paginationSel := doc.Find("div.pagination")
	// Pre-allocate slice space guessing Letterboxd's standard grid size (typically 20-50 per page)
	movies := make([]*ListMovie, 0, 50)

	posterGridSel.Find("li.posteritem").Each(func(i int, s *goquery.Selection) {
		l := &ListMovie{}
		movieDiv := s.Find("div.react-component")
		if slug, exists := movieDiv.Attr("data-item-slug"); exists {
			l.Slug = slug
		}
		if name, exists := movieDiv.Attr("data-item-name"); exists {
			title, year := extractMovieInfo(name)
			l.Title = title
			l.ReleaseYear = year
		}
		l.Rank = convertTextToInt(s.Find("p.list-number").Text())
		movies = append(movies, l)

	})

	return movies, getNextLinkFromList(paginationSel)
}

// getUrlFromList extracts the canonical OpenGraph URL configuration.
func getUrlFromList(doc *goquery.Document) string {
	return doc.Find("meta[property='og:url']").AttrOr("content", "")
}

// getTitleFromList extracts the list's title from the OpenGraph header tags.
func getTitleFromList(doc *goquery.Document) string {
	return doc.Find("meta[property='og:title']").AttrOr("content", "")
}

// getNumberOfMoviesFromList extracts the number of movies
func getNumberOfMoviesFromList(doc *goquery.Document) int {
	desc := doc.Find("meta[name='description']").AttrOr("content", "")
	if desc == "" {
		return 0
	}

	re := regexp.MustCompile(`A list of (\d+) films compiled`)
	matches := re.FindStringSubmatch(desc)
	if len(matches) < 2 {
		return 0
	}
	countStr := matches[1]
	filmCount, err := strconv.Atoi(countStr)
	if err != nil {
		return 0
	}
	return filmCount
}

// getUserNamesFromList safely extracts unique usernames from the list header.
// Curated list pages can occasionally contain multiple names/avatars.
func getUserNamesFromList(headerSel *goquery.Selection) []string {
	var userNames []string
	seen := make(map[string]bool)

	// Just look for the anchors right inside the pre-located header selection
	headerSel.Find("a.name, a.avatar").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			userName := strings.Trim(href, "/")

			if userName != "" && !seen[userName] {
				seen[userName] = true
				userNames = append(userNames, userName)
			}
		}
	})

	return userNames
}

// getNextLinkFromList checks the bottom pagination element to grab the forward link.
func getNextLinkFromList(paginationSel *goquery.Selection) string {
	return paginationSel.Find("a.next").AttrOr("href", "")
}

// getPublishedTime attempts to pull the list's native creation timestamp in ISO format.
// Returns a zero-value time.Time{} if absent or unparseable.
func getPublishedTime(doc *goquery.Document) time.Time {
	publishedTimeStr, exists := doc.Find("span.published time").Attr("datetime")
	if !exists {
		return time.Time{}
	}
	parsedTime, err := time.Parse(time.RFC3339, publishedTimeStr)
	if err != nil {
		return time.Time{}
	}
	return parsedTime
}

// getUpdatedTime attempts to pull the list's native modification timestamp in ISO format.
// Returns a zero-value time.Time{} if the list has never been modified.
func getUpdatedTime(doc *goquery.Document) time.Time {
	updatedTimeStr, exists := doc.Find("span.updated time").Attr("datetime")
	if !exists {
		return time.Time{}
	}

	parsedTime, err := time.Parse(time.RFC3339, updatedTimeStr)
	if err != nil {
		return time.Time{}
	}

	return parsedTime
}

// extractMovieInfo will extract the movie title and year from a string like "Birdman or (The Unexpected Virtue of Ignorance) (2014)"
func extractMovieInfo(input string) (string, int) {
	// Compile the regular expression
	re := regexp.MustCompile(`^(.+)\s+\((\d{4})\)$`)

	// Find the submatches
	matches := re.FindStringSubmatch(input)

	// If we don't get exactly 3 elements (the full match + 2 caputure groups),
	// the string didn't match the expected format.
	if len(matches) != 3 {
		return "", 0
	}

	// Clean up any trailing/leading whitespace just in case
	title := strings.TrimSpace(matches[1])
	year, err := strconv.Atoi(matches[2])
	if err != nil {
		year = 0
	}

	return title, year
}
