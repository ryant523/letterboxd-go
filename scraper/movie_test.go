package letterboxd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type movieTestCase struct {
	name     string
	filename string
	validate func(*testing.T, *Movie)
}

func TestParseMovie(t *testing.T) {
	tests := []movieTestCase{
		{
			name:     "Incendies - Comprehensive Metadata Parsing",
			filename: "incendies.html",
			validate: validateIncendies,
		}, {
			name:     "Movie 43 - High Director, Composer and Cinematographer Count",
			filename: "movie-43.html",
			validate: validateMovie43,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", "movies", tt.filename)
			f, err := os.Open(path)
			if err != nil {
				t.Fatalf("failed to read test data %s: %v", tt.filename, err)
			}
			doc, err := getDocument(f)
			if err != nil {
				t.Fatalf("failed to parse test data doc: %v", err)
			}
			movie, err := parseMovie(doc)
			if err != nil {
				t.Fatalf("failed to parse movie: %v", err)
			}

			tt.validate(t, movie)

		})
	}

}

func validateIncendies(t *testing.T, actual *Movie) {
	expected := &Movie{
		Title:       "Incendies",
		ReleaseYear: 2010,
		Slug:        "incendies",
		ImdbID:      "tt1255953",
		TmdbID:      46738,
		Runtime:     131,
		Language:    "French",
		Directors: []Person{
			{Name: "Denis Villeneuve", Slug: "denis-villeneuve"},
		},
		Composers: []Person{
			{Name: "Grégoire Hetzel", Slug: "gregoire-hetzel"}, // Fixed to match expected slug convention
		},
		Writers: []Person{
			{Name: "Valérie Beaugrand-Champagne", Slug: "valerie-beaugrand-champagne"},
			{Name: "Denis Villeneuve", Slug: "denis-villeneuve"},
		},
		OriginalWriters: []Person{
			{Name: "Wajdi Mouawad", Slug: "wajdi-mouawad"},
		},
		Cinematographers: []Person{
			{Name: "André Turpin", Slug: "andre-turpin"},
		},
		Genres: []string{"Mystery", "Drama", "War"},
		// ignore Actors here.
	}

	// 2. Compare everything in a single line (ignoring the long Actors slice for this check)
	if diff := cmp.Diff(expected, actual, cmp.FilterPath(func(p cmp.Path) bool {
		return p.String() == "Actors" // We will check actors separately since there are 25
	}, cmp.Ignore())); diff != "" {
		t.Errorf("parseMovie() mismatch (-expected +actual):\n%s", diff)
	}

	// 3. Testing that large slice becomes trivial
	if len(actual.Actors) != 25 {
		t.Errorf("expected 25 actors, got %d", len(actual.Actors))
	} else if actual.Actors[0].Name != "Lubna Azabal" {
		t.Errorf("expected first actor 'Lubna Azabal', got '%s'", actual.Actors[0].Name)
	}
}

func validateMovie43(t *testing.T, actual *Movie) {
	expected := &Movie{
		Title:       "Movie 43",
		ReleaseYear: 2013,
		Slug:        "movie-43",
		ImdbID:      "tt1333125",
		TmdbID:      87818,
		Runtime:     98,
		Language:    "English",
		Directors: []Person{
			{Name: "Steven Brill", Slug: "steven-brill"},
			{Name: "Elizabeth Banks", Slug: "elizabeth-banks"},
			{Name: "Steve Carr", Slug: "steve-carr"},
			{Name: "James Duffy", Slug: "james-duffy"},
			{Name: "Griffin Dunne", Slug: "griffin-dunne"},
			{Name: "Peter Farrelly", Slug: "peter-farrelly"},
			{Name: "Patrik Forsberg", Slug: "patrik-forsberg"},
			{Name: "James Gunn", Slug: "james-gunn"},
			{Name: "Brett Ratner", Slug: "brett-ratner-1"},
			{Name: "Will Graham", Slug: "will-graham"},
			{Name: "Jonathan van Tulleken", Slug: "jonathan-van-tulleken"},
			{Name: "Rusty Cundieff", Slug: "rusty-cundieff"},
		},
		Composers: []Person{
			{Name: "Christophe Beck", Slug: "christophe-beck"},
			{Name: "Tyler Bates", Slug: "tyler-bates"},
			{Name: "Leo Birenberg", Slug: "leo-birenberg"},
			{Name: "William Goodrum", Slug: "william-goodrum"},
			{Name: "Dave Hodge", Slug: "dave-hodge"},
			{Name: "Matt Jantzen", Slug: "matt-jantzen"},
		},
		Cinematographers: []Person{
			{Name: "Frankie DeMarco", Slug: "frankie-demarco"},
			{Name: "Steve Gainer", Slug: "steve-gainer"},
			{Name: "Matthew F. Leonetti", Slug: "matthew-f-leonetti"},
			{Name: "William Rexer", Slug: "william-rexer"},
			{Name: "Newton Thomas Sigel", Slug: "newton-thomas-sigel"},
			{Name: "Tim Suhrstedt", Slug: "tim-suhrstedt"},
			{Name: "Daryn Okada", Slug: "daryn-okada"},
			{Name: "Mattias Rudh", Slug: "mattias-rudh"},
			{Name: "Eric Scherbarth", Slug: "eric-scherbarth"},
		},
		// ignore Actors, Writers and Genres
	}

	// 2. Compare everything in a single line (ignoring the long Actors slice for this check)
	if diff := cmp.Diff(expected, actual, cmp.FilterPath(func(p cmp.Path) bool {
		switch p.String() {
		case "Actors":
			return true
		case "Writers":
			return true
		case "Genres":
			return true
		}
		return false
	}, cmp.Ignore())); diff != "" {
		t.Errorf("parseMovie() mismatch (-expected +actual):\n%s", diff)
	}

}
