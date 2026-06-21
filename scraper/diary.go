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

type Diary struct {
	Username string

	url       string
	nextLink  string
	firstPage []*DiaryEntry
	client    *Client
}

type DiaryEntry struct {
	DateWatched time.Time
	Slug        string
	Rating      int
	Title       string
	ReleaseYear int
}

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

func (c *Client) GetDiary(ctx context.Context, userName string, opts ...Option) (*Diary, error) {
	u, _ := url.Parse(fmt.Sprintf("/%s/diary", userName))
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	if optionsPath := options.Build(); optionsPath != "" {
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

func getDiaryMovieRating(row *goquery.Selection) int {
	if ratingStr, exists := row.Find("input.rateit-field").Attr("value"); exists {
		if ratingVal, err := strconv.Atoi(ratingStr); err == nil {
			return ratingVal
		}
	}
	return 0
}

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

/*
<div class="pagination">
<div class="paginate-nextprev paginate-disabled">
<span class="previous">Newer</span></div>
<div class="paginate-nextprev">
<a class="next" href="/abbynormalz/diary/films/page/2/">Older</a></div>
*/
