package letterboxd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// Todo add Mock Roundtripper to test Movies() iterator

type listTestCase struct {
	name     string
	filename string
	validate func(*testing.T, *List)
}

func TestParseList(t *testing.T) {
	tests := []listTestCase{
		{
			name:     "Letterbox's Top 500 Films",
			filename: "top500-page1.html",
			validate: validateTop500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", "lists", tt.filename)
			f, err := os.Open(path)
			if err != nil {
				t.Fatalf("failed to read test data %s: %v", tt.filename, err)
			}
			doc, err := getDocument(f)
			if err != nil {
				t.Fatalf("failed to parse test data doc: %v", err)
			}
			list, err := parseMovieList(doc)
			if err != nil {
				t.Fatalf("failed to parse list: %v", err)
			}

			tt.validate(t, list)

		})
	}

}

func validateTop500(t *testing.T, actual *List) {
	layout := "2006-01-02 15:04:05.000Z"
	createdAt, _ := time.Parse(layout, "2013-11-08 10:38:22.466Z")
	updatedAt, _ := time.Parse(layout, "2026-06-09 09:33:55.374Z")
	expected := &List{
		Title:     "Letterboxd's Top 500 Films",
		UserNames: []string{"official", "dave"},
		Url:       "https://letterboxd.com/official/list/letterboxds-top-500-films/",
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if diff := cmp.Diff(expected, actual, cmpopts.IgnoreUnexported(List{})); diff != "" {
		t.Errorf("parseMovieList() mismatch (-expected +actual):\n%s", diff)
	}

	if actual.nextLink != "/official/list/letterboxds-top-500-films/page/2/" {
		t.Errorf("List nextLink expected '/official/list/letterboxds-top-500-films/page/2/', got '%s'", actual.nextLink)
	}
	if len(actual.firstMovies) != 100 {
		t.Errorf("List firstMovies expected length '100', got '%d'", len(actual.firstMovies))
	}
	if len(actual.firstMovies) > 0 {
		movie := actual.firstMovies[0]
		if movie.Slug != "harakiri" {
			t.Errorf("List firstMovies[0] expected Slug 'harakiri', got '%s'", movie.Slug)
		}
		if movie.Title != "Harakiri" {
			t.Errorf("List firstMovies[0] expected Title 'Harakiri', got '%s'", movie.Title)
		}
		if movie.ReleaseYear != 1962 {
			t.Errorf("List firstMovies[0] expected ReleaseYear '1962', got '%d'", movie.ReleaseYear)
		}
	}
}
