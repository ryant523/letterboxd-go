package letterboxd

import (
	"context"
	"fmt"
	"iter"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Diary represents a letterboxd diary. The entries can be retrieved with the Entries() iterator
// or using GetAllEntries().
type Diary struct {
	Username string

	url       string
	nextLink  string
	firstPage []*DiaryEntry
	client    *Client
}

// DiaryEntry represents a single movie watched in a user's diary.
type DiaryEntry struct {
	DateWatched time.Time
	Slug        string
	Rating      int
	Title       string
	ReleaseYear int
}

// Entries is an iterator used to retrieve all DiaryEntry structs.
func (d *Diary) Entries(ctx context.Context) iter.Seq2[*DiaryEntry, error] {
	return func(yield func(*DiaryEntry, error) bool) {
		for _, e := range d.firstPage {
			if !yield(e, nil) {
				return // User broke out of loop early
			}
		}
		if d.nextLink == "" {
			return
		}
		nextLink := d.nextLink
		for {
			doc, err := d.client.getHtml(ctx, nextLink)
			if err != nil {
				yield(&DiaryEntry{}, fmt.Errorf("failed fetching diary: %s, %w", nextLink, err))
				return
			}
			var entries []*DiaryEntry
			entries, nextLink = parseDiaryEntries(doc)
			if len(entries) == 0 {
				return
			}
			for _, e := range entries {
				if !yield(e, nil) {
					return
				}
			}
			if nextLink == "" {
				return
			}

		}
	}
}

// GetAllEntries returns a slice of all *DiaryEntry.
func (d *Diary) GetAllEntries(ctx context.Context) ([]*DiaryEntry, error) {
	entries := make([]*DiaryEntry, 0, 50)
	for entry, err := range d.Entries(ctx) {
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil

}

// GetDiary will retrieve a user's diary. This call will parse the first page
// of the diary. The rest of the diary entries are not retrieved until either
// Diary.Entries(ctx) or GetAllEntries(ctx) is called.
//
// This accepts Option that can be used to filter and/or sort the results.
func (c *Client) GetDiary(ctx context.Context, userName string, opts ...Option) (*Diary, error) {
	u, _ := url.Parse(fmt.Sprintf("/%s/diary", userName))
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	if optionsPath := options.build(); optionsPath != "" {
		u.Path = path.Join(u.Path, optionsPath)
	}
	doc, err := c.getHtml(ctx, u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get diary for %s: %w", userName, err)
	}
	d := &Diary{}
	entries, next := parseDiaryEntries(doc)
	d.firstPage = entries
	d.nextLink = next
	d.url = u.String()
	d.client = c
	return d, nil
}

// parseDiaryEntries will parse a single page of the diary and find all the diary entries
// as well as returning the next pagination link.
func parseDiaryEntries(doc *goquery.Document) ([]*DiaryEntry, string) {
	entries := make([]*DiaryEntry, 0)
	doc.Find("tr.diary-entry-row").Each(func(i int, row *goquery.Selection) {
		entry := &DiaryEntry{}
		component := row.Find(".react-component[data-item-slug]")
		entry.Slug = component.AttrOr("data-item-slug", "")
		if name, exists := component.Attr("data-item-name"); exists {
			title, year := extractMovieInfo(name)
			entry.Title = title
			entry.ReleaseYear = year
		}
		entry.DateWatched = getDiaryDateWatched(row)
		entry.Rating = getDiaryMovieRating(row)

		entries = append(entries, entry)
	})

	var next string

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		if s.Find(".paginate-pages").Length() > 0 {
			next = s.Find(".paginate-nextprev a.next").AttrOr("href", "")
		}
	})

	return entries, next
}

// getDiaryMovieRating parses the rating if it exists of the diary entry.
// This rating is 1-10 instead of the 5 stars.
func getDiaryMovieRating(row *goquery.Selection) int {
	if ratingStr, exists := row.Find("input.rateit-field").Attr("value"); exists {
		if ratingVal, err := strconv.Atoi(ratingStr); err == nil {
			return ratingVal
		}
	}
	return 0
}

// getDiaryDateWatched parses the date the diary entry was watched.
func getDiaryDateWatched(row *goquery.Selection) time.Time {
	if href, exists := row.Find("a.daydate").Attr("href"); exists {

		// Splitting "/abbynormalz/diary/films/for/2026/05/17/" by "/"
		// yields a slice where the last pieces are the year, month, and day.
		parts := strings.Split(strings.Trim(href, "/"), "/")

		if len(parts) >= 3 {
			// Grab the last 3 segments: "2026", "05", "17"
			yearStr := parts[len(parts)-3]
			monthStr := parts[len(parts)-2]
			dayStr := parts[len(parts)-1]

			// Combine into a standardized layout: "2026-05-17"
			dateStr := fmt.Sprintf("%s-%s-%s", yearStr, monthStr, dayStr)

			// Parse using Go's ISO-8601 layout
			parsedDate, err := time.Parse("2006-01-02", dateStr)
			if err == nil {
				return parsedDate
			}
		}
	}
	return time.Time{}
}
