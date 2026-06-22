package letterboxd

import (
	"context"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// User represents a letterboxd user to get basic stats along with the Top Four.
// Use Diary to get a user's diary.
type User struct {
	UserName       string
	DisplayName    string
	TotalFilms     int
	TotalLists     int
	TotalFollowing int
	TotalFollowers int
	TotalWatchList int // only if watchlist is public
	TotalDiary     int
	TopFour        []*ListMovie
	Badge          string // Pro, Patron
}

// GetUser will retrieve and scrape the user profile page.
func (c *Client) GetUser(ctx context.Context, username string) (*User, error) {
	html, err := c.getHtml(ctx, username)
	if err != nil {
		return nil, err
	}
	return parseUser(html)
}

func parseUser(doc *goquery.Document) (*User, error) {
	headerSel := doc.Find("section.profile-header")
	statsSel := doc.Find(".profile-stats")
	favsSel := doc.Find("#favourites")

	return &User{
		UserName:       getUserName(headerSel),
		DisplayName:    getUserDisplayName(headerSel),
		TotalFilms:     getStatCount(statsSel, "/films/"),
		TotalLists:     getStatCount(statsSel, "/lists/"),
		TotalFollowing: getStatCount(statsSel, "/following/"),
		TotalFollowers: getStatCount(statsSel, "/followers/"),
		TotalWatchList: getUserTotalWatchList(doc), // Left as doc since it's an aside
		TotalDiary:     getUserTotalDiaryEntries(doc),
		TopFour:        getUserTopFourFilms(favsSel),
		Badge:          getBadge(headerSel), // Badges usually live in the header
	}, nil
}

func getStatCount(statsSel *goquery.Selection, pathSuffix string) int {
	// We search ONLY within the already-located .profile-stats element
	rawCount := statsSel.Find("a[href$='" + pathSuffix + "'] .value").Text()
	return convertTextToInt(rawCount)
}

func getUserName(headerSel *goquery.Selection) string {
	return headerSel.AttrOr("data-person", "")
}
func getUserDisplayName(headerSel *goquery.Selection) string {
	return headerSel.Find(".profile-avatar img").AttrOr("alt", "")
}

func getUserTotalWatchList(doc *goquery.Document) int {
	rawCount := doc.Find(".watchlist-aside .all-link").Text()
	if rawCount == "" {
		return 0
	}
	return convertTextToInt(rawCount)
}

func getUserTotalDiaryEntries(doc *goquery.Document) int {
	rawCount := doc.Find("section .all-link[href$='/diary/']").Text()
	return convertTextToInt(rawCount)
}

func getUserTopFourFilms(favsSel *goquery.Selection) []*ListMovie {
	movies := make([]*ListMovie, 0, 4)

	// We already have #favourites, so just look for the components inside it
	favsSel.Find(".react-component[data-item-slug]").Each(func(i int, s *goquery.Selection) {
		l := &ListMovie{}
		if slug, exists := s.Attr("data-item-slug"); exists {
			l.Slug = slug
		}
		if name, exists := s.Attr("data-item-name"); exists {
			title, year := extractMovieInfo(name)
			l.Title = title
			l.ReleaseYear = year
		}
		movies = append(movies, l)
	})

	return movies
}

func getBadge(headerSel *goquery.Selection) string {
	return headerSel.Find(".badge").Text()
}

// convertTextToInt will cleanup whitespace around text and convert it to int
func convertTextToInt(text string) int {
	cleanText := strings.TrimSpace(text)
	if strings.Contains(cleanText, ",") {
		cleanText = strings.ReplaceAll(cleanText, ",", "")
	}

	count, err := strconv.Atoi(cleanText)
	if err != nil {
		return 0
	}
	return count
}
