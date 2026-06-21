package letterboxd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type userTestCase struct {
	name     string
	filename string
	validate func(*testing.T, *User)
}

func TestParseUser(t *testing.T) {
	tests := []userTestCase{
		{
			name:     "abbynormalz",
			filename: "abbynormalz.html",
			validate: validateAbbyNormalz,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", "users", tt.filename)
			f, err := os.Open(path)
			if err != nil {
				t.Fatalf("failed to read test data %s: %v", tt.filename, err)
			}
			doc, err := getDocument(f)
			if err != nil {
				t.Fatalf("failed to parse test data doc: %v", err)
			}
			user, err := parseUser(doc)
			if err != nil {
				t.Fatalf("failed to parse list: %v", err)
			}

			tt.validate(t, user)

		})
	}

}

func validateAbbyNormalz(t *testing.T, actual *User) {
	expected := &User{
		UserName:       "AbbyNormalz",
		DisplayName:    "AbbyNormalz",
		TotalFilms:     3353,
		TotalFollowing: 0,
		TotalFollowers: 0,
		TotalWatchList: 0,
		TotalLists:     1,
		TotalDiary:     2003,
		Badge:          "Pro",
	}

	if diff := cmp.Diff(
		expected,
		actual,
		cmp.FilterPath(func(p cmp.Path) bool {
			switch p.String() {
			case "TopFour":
				return true
			}
			return false
		}, cmp.Ignore()),
	); diff != "" {
		t.Errorf("parseUser() mismatch (-expected +actual):\n%s", diff)
	}

	if len(actual.TopFour) != 4 {
		t.Errorf("parseUser() TopFour length expected '4', got '%d'", len(actual.TopFour))
	}
	if len(actual.TopFour) == 4 {
		movie := actual.TopFour[0]
		if movie.Title != "Young Frankenstein" {
			t.Errorf("parseUser() TopFour[0] Title expected 'Young Frankenstein', got '%s'", movie.Title)
		}
		if movie.ReleaseYear != 1974 {
			t.Errorf("parseUser() TopFour[0] ReleaseYear expected '1974', got %d", movie.ReleaseYear)
		}
		if movie.Slug != "young-frankenstein" {
			t.Errorf("parseUser() TopFour[0] Slug expected 'young-frankenstein', got %s", movie.Slug)
		}

	}
}
