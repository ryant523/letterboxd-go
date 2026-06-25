package letterboxd

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Person represents a genric cast or crew member associated with a film.
type Person struct {
	Name string
	Slug string // The unique URL slug identifier of the person on Letterboxd
}

type Movie struct {
	Title            string
	ReleaseYear      int
	Runtime          int // The film runtime in minutes.
	Slug             string
	ImdbID           string
	TmdbID           int
	Language         string
	Directors        []Person
	Writers          []Person
	OriginalWriters  []Person
	Actors           []Person
	Composers        []Person
	Cinematographers []Person
	Genres           []string
}

// GetMovieBySlug requests a film's public page from Letterboxd by its unique slug identifier
// and parses it into a [*Movie]
func (c *Client) GetMovieBySlug(ctx context.Context, slug string) (*Movie, error) {
	path := "/film/" + slug
	return c.getMovie(ctx, path)
}

// GetMovieByImdb requests a film's public page from Letterboxd by its Imdb ID
func (c *Client) GetMovieByImdb(ctx context.Context, imdbID string) (*Movie, error) {
	path := "imdb/" + imdbID
	return c.getMovie(ctx, path)
}

// getMovie will make the request to the url and parse the Movie
func (c *Client) getMovie(ctx context.Context, url string) (*Movie, error) {
	html, err := c.getHtml(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to get movie HTML for url %s: %w", url, err)
	}
	movie, err := parseMovie(html)
	if err != nil {
		return nil, fmt.Errorf("failed to parse movie HTML for url %s: %w", url, err)
	}
	return movie, nil
}

type movieParser struct {
	doc        *goquery.Document
	posterSel  *goquery.Selection
	detailsSel *goquery.Selection
}

func newMovieParser(doc *goquery.Document) *movieParser {
	return &movieParser{
		doc:        doc,
		posterSel:  doc.Find("#poster-modal .react-component"),
		detailsSel: doc.Find("#tab-panel-details"),
	}
}

// crewData acts as an internal structural translation matrix to isolate separate
// professional roles found within the single Letterboxd HTML crew panel context.
type crewData struct {
	Directors        []Person
	Composers        []Person
	Writers          []Person
	OriginalWriters  []Person
	Cinematographers []Person
}

// parseMovie maps a raw HTML goquery DOM tree directly into a strongly typed [*Movie] struct.
// It relies on predefined internal element configurations to cleanly extract individual data fields.
func parseMovie(doc *goquery.Document) (*Movie, error) {
	p := newMovieParser(doc)
	title := p.title()
	if title == "" {
		return nil, fmt.Errorf("could not find movie title; page layout may have changed or page is invalid")
	}

	crew := p.extractCrew()

	return &Movie{
		Title:            title,
		Slug:             p.slug(),
		ReleaseYear:      p.releaseYear(),
		ImdbID:           p.imdbID(),
		TmdbID:           p.tmdbID(),
		Runtime:          p.runtime(),
		Language:         p.language(),
		Genres:           p.genres(),
		Directors:        crew.Directors,
		Composers:        crew.Composers,
		Writers:          crew.Writers,
		OriginalWriters:  crew.OriginalWriters,
		Cinematographers: crew.Cinematographers,
		Actors:           p.actors(),
	}, nil
}

// extractCrew iterates through the film's crew panel structure, mapping categorized professional
// headings ("Director", "Writer", etc.) directly to arrays of [Person] structs.
func (p *movieParser) extractCrew() crewData {
	var crew crewData

	crewSel := p.doc.Find("#tab-panel-crew")

	crewSel.Find("h3").Each(func(i int, s *goquery.Selection) {
		roleText := strings.TrimSpace(s.Find(".crewrole.-full").Text())

		var people []Person
		s.NextFiltered("div.text-sluglist").Find("a.text-slug").Each(func(j int, anchor *goquery.Selection) {
			name := cleanName(anchor.Text())
			href, exists := anchor.Attr("href")

			if name != "" && exists {
				people = append(people, Person{
					Name: name,
					Slug: extractPersonSlug(href),
				})
			}
		})

		switch roleText {
		case "Director", "Directors":
			crew.Directors = people
		case "Composer", "Composers":
			crew.Composers = people
		case "Writer", "Writers":
			crew.Writers = people
		case "Original Writer", "Original Writers":
			crew.OriginalWriters = people
		case "Cinematography":
			crew.Cinematographers = people
		}
	})

	return crew
}

// title extracts the text element containing the primary movie headline.
func (p *movieParser) title() string {
	return strings.TrimSpace(p.doc.Find("h1.headline-1.primaryname").Text())
}

// movieSlug reads tracking elements inside the poster DOM element to determine the canonical film path.
func (p *movieParser) slug() string {
	if slug, exists := p.posterSel.Attr("data-item-slug"); exists {
		return strings.TrimSpace(slug)
	}

	if link, exists := p.posterSel.Attr("data-item-link"); exists {
		parts := strings.Split(strings.Trim(link, "/"), "/")
		if len(parts) > 1 {
			return parts[1]
		}
	}
	return ""
}

// releaseYear pulls the primary link value containing the designated production calendar year.
func (p *movieParser) releaseYear() int {
	yearStr := strings.TrimSpace(p.doc.Find(".releaseyear a").First().Text())
	if yearStr == "" {
		yearStr = strings.TrimSpace(p.doc.Find("span.releasedate").Text())
	}

	if year, err := strconv.Atoi(yearStr); err == nil {
		return year
	}
	return 0
}

// imdbID isolates the standardized tracking ID extracted directly from the outgoing IMDb link.
func (p *movieParser) imdbID() string {
	if imdbHref, exists := p.doc.Find(`a[data-track-action="IMDb"]`).Attr("href"); exists {
		if idx := strings.Index(imdbHref, "?"); idx != -1 {
			imdbHref = imdbHref[:idx]
		}
		parts := strings.Split(imdbHref, "/")
		if len(parts) > 4 {
			return parts[4]
		}
	}
	return ""
}

// tmdbID parses the numerical tracking identifier retrieved from outbound TMDB redirect references.
func (p *movieParser) tmdbID() int {
	if tmdbHref, exists := p.doc.Find(`a[data-track-action="TMDB"]`).Attr("href"); exists {
		parts := strings.Split(tmdbHref, "/")
		if len(parts) > 4 {
			if tid, err := strconv.Atoi(parts[4]); err == nil {
				return tid
			}
		}
	}
	return 0
}

// runtime parses the movie runtime in minutes
func (p *movieParser) runtime() int {
	footerText := p.doc.Find("p.text-link.text-footer").Text()

	// Normalize the string. HTML spaces (&nbsp;) often convert into special
	// unicode spaces (\u00a0). Replacing them with standard spaces makes parsing safer.
	footerText = strings.ReplaceAll(footerText, "\u00a0", " ")

	// Extract the text before "mins"
	// Example: " 136 mins   More at IMDb..." -> split by "mins" -> index 0 is " 136 "
	parts := strings.Split(footerText, "mins")
	if len(parts) < 2 {
		return 0
	}

	cleanNumStr := strings.TrimSpace(parts[0])

	runtime, err := strconv.Atoi(cleanNumStr)
	if err != nil {
		return 0
	}

	return runtime
}

// language crawls through detail list blocks looking explicitly for the primary dialogue classification field
func (p *movieParser) language() string {
	primaryAnchor := p.detailsSel.Find("h3:contains('Primary Language')").Next().Find("a.text-slug")
	if primaryAnchor.Length() > 0 {
		return strings.TrimSpace(primaryAnchor.First().Text())
	}

	fallbackAnchor := p.detailsSel.Find("h3:contains('Language')").First().Next().Find("a.text-slug")
	if fallbackAnchor.Length() > 0 {
		return strings.TrimSpace(fallbackAnchor.First().Text())
	}

	return ""
}

// genres scans categorical metadata anchor elements tagged within the genre panel element.
func (p *movieParser) genres() []string {
	var genres []string
	// Find anchor tags inside the container whose href contains "/genre/"
	p.doc.Find("#tab-panel-genres a[href*='/genre/']").Each(func(i int, anchor *goquery.Selection) {
		name := strings.TrimSpace(anchor.Text())
		if name != "" {
			genres = append(genres, name)
		}
	})

	return genres
}

// actors isolates individual cast name components mapped within the corresponding billing layout container.
func (p *movieParser) actors() []Person {
	var actors []Person

	castSel := p.doc.Find("#tab-panel-cast")

	castSel.Find("a[href*='/actor/']").Each(func(i int, anchor *goquery.Selection) {
		rawName := anchor.Text()
		name := cleanName(rawName)
		href, exists := anchor.Attr("href")

		if name != "" && exists {
			actors = append(actors, Person{
				Name: name,
				Slug: extractPersonSlug(href),
			})
		}
	})

	return actors
}

// extractPersonSlug isolates the trailing subdirectory slug value representing an actor or crew entity.
func extractPersonSlug(href string) string {
	// If href is "/director/denis-villeneuve/", Clean() handles the slashes
	// and Base() snaps right to the last segment: "denis-villeneuve"
	return path.Base(path.Clean(href))
}

// cleanName strips duplicate whitespace patterns, padding characters, and line breaks from parsed DOM names.
func cleanName(name string) string {
	// If it doesn't contain tabs, newlines, or double spaces, it's already clean.
	// (Letterboxd names rarely have trailing/leading spaces unless wrapped in newlines)
	if !strings.ContainsAny(name, "\n\t") && !strings.Contains(name, "  ") {
		return strings.TrimSpace(name)
	}

	words := strings.Fields(name)
	return strings.Join(words, " ")
}
